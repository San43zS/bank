package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"banking-platform/internal/apperr"
	"banking-platform/internal/model"
	"banking-platform/internal/storage"
	"github.com/google/uuid"
)

const ExchangeRateUSDtoEUR = 0.92

var systemBankUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type TransactionService struct {
	db              *storage.DB
	accountRepo     *storage.AccountRepository
	transactionRepo *storage.TransactionRepository
	ledgerRepo      *storage.LedgerRepository
	userRepo        *storage.UserRepository
	logger          *slog.Logger
}

func NewTransactionService(
	db *storage.DB,
	accountRepo *storage.AccountRepository,
	transactionRepo *storage.TransactionRepository,
	ledgerRepo *storage.LedgerRepository,
	userRepo *storage.UserRepository,
	logger *slog.Logger,
) *TransactionService {
	return &TransactionService{
		db:              db,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		ledgerRepo:      ledgerRepo,
		userRepo:        userRepo,
		logger:          logger,
	}
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID uuid.UUID, req *model.TransferRequest) (*model.TransactionResponse, error) {
	s.logger.Info("Processing transfer", "from_user_id", fromUserID, "to_user_id", req.ToUserID, "amount", req.Amount, "currency", req.Currency)

	if req.Currency != model.CurrencyUSD && req.Currency != model.CurrencyEUR {
		s.logger.Warn("Invalid currency", "currency", req.Currency)
		return nil, fmt.Errorf("invalid currency: %s", req.Currency)
	}

	amountCents, err := amountToCents(req.Amount)
	if err != nil {
		s.logger.Warn("Invalid amount precision", "amount", req.Amount)
		return nil, apperr.ErrInvalidAmount
	}

	tx, err := s.db.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	fromAccountID, err := s.accountRepo.FindAccountIDTx(tx, fromUserID, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender account: %w", err)
	}
	toAccountID, err := s.accountRepo.FindAccountIDTx(tx, req.ToUserID, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	lockIDs := []uuid.UUID{fromAccountID, toAccountID}
	sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })

	locked := make(map[uuid.UUID]*model.Account, 2)
	for _, id := range lockIDs {
		acc, err := s.accountRepo.LockAccountForUpdate(tx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to lock account: %w", err)
		}
		locked[id] = acc
	}

	fromAccount := locked[fromAccountID]
	toAccount := locked[toAccountID]
	if fromAccount == nil || toAccount == nil {
		return nil, fmt.Errorf("failed to lock accounts")
	}

	if fromAccount.UserID != fromUserID || toAccount.UserID != req.ToUserID {
		return nil, apperr.ErrUnauthorized
	}

	fromBalanceCents, err := amountToCents(fromAccount.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid sender balance in db")
	}
	toBalanceCents, err := amountToCents(toAccount.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient balance in db")
	}

	if fromBalanceCents < amountCents {
		s.logger.Warn("Insufficient funds", "user_id", fromUserID, "balance_cents", fromBalanceCents, "amount_cents", amountCents)
		return nil, apperr.ErrInsufficientFunds
	}

	transactionID := uuid.New()
	now := time.Now()
	amountFloat := centsToFloat(amountCents)
	transaction := &model.Transaction{
		ID:            transactionID,
		Type:          model.TransactionTypeTransfer,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        amountFloat,
		Currency:      req.Currency,
		Description:   fmt.Sprintf("Transfer %s %.2f from %s to %s", req.Currency, amountFloat, fromUserID, req.ToUserID),
		CreatedAt:     now,
	}

	if err := s.transactionRepo.Create(tx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	fromEntry := &model.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     fromAccount.ID,
		Amount:        -amountFloat,
		CreatedAt:     now,
	}
	if err := s.ledgerRepo.CreateEntry(tx, fromEntry); err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	toEntry := &model.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     toAccount.ID,
		Amount:        amountFloat,
		CreatedAt:     now,
	}
	if err := s.ledgerRepo.CreateEntry(tx, toEntry); err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	newFromBalanceCents := fromBalanceCents - amountCents
	newToBalanceCents := toBalanceCents + amountCents

	if err := s.accountRepo.UpdateBalanceString(tx, fromAccount.ID, centsToDecimalString(newFromBalanceCents)); err != nil {
		return nil, fmt.Errorf("failed to update sender balance: %w", err)
	}

	if err := s.accountRepo.UpdateBalanceString(tx, toAccount.ID, centsToDecimalString(newToBalanceCents)); err != nil {
		return nil, fmt.Errorf("failed to update recipient balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transfer transaction", "error", err, "transaction_id", transactionID)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.logger.Info("Transfer completed successfully", "transaction_id", transactionID, "from_user_id", fromUserID, "to_user_id", req.ToUserID, "amount", amountFloat, "currency", req.Currency)

	fromUser, _ := s.userRepo.GetByID(fromUserID)
	toUser, _ := s.userRepo.GetByID(req.ToUserID)

	response := &model.TransactionResponse{
		ID:            transaction.ID,
		Type:          transaction.Type,
		FromAccountID: transaction.FromAccountID,
		ToAccountID:   transaction.ToAccountID,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		Description:   transaction.Description,
		CreatedAt:     transaction.CreatedAt,
	}
	if fromUser != nil {
		response.FromUserEmail = &fromUser.Email
	}
	if toUser != nil {
		response.ToUserEmail = &toUser.Email
	}

	return response, nil
}

func (s *TransactionService) Exchange(ctx context.Context, userID uuid.UUID, req *model.ExchangeRequest) (*model.TransactionResponse, error) {
	s.logger.Info("Processing exchange", "user_id", userID, "from_currency", req.FromCurrency, "to_currency", req.ToCurrency, "amount", req.Amount)

	if req.FromCurrency == req.ToCurrency {
		s.logger.Warn("Same currency for exchange", "currency", req.FromCurrency)
		return nil, fmt.Errorf("from and to currencies must be different")
	}

	amountCents, err := amountToCents(req.Amount)
	if err != nil {
		s.logger.Warn("Invalid amount precision", "amount", req.Amount)
		return nil, apperr.ErrInvalidAmount
	}

	exchangeRate, convertedCents, err := convertExchange(amountCents, req.FromCurrency, req.ToCurrency)
	if err != nil {
		return nil, err
	}
	convertedAmount := centsToFloat(convertedCents)

	tx, err := s.db.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	userFromID, err := s.accountRepo.FindAccountIDTx(tx, userID, req.FromCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get from account: %w", err)
	}
	userToID, err := s.accountRepo.FindAccountIDTx(tx, userID, req.ToCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get to account: %w", err)
	}
	bankFromID, err := s.accountRepo.FindAccountIDTx(tx, systemBankUserID, req.FromCurrency)
	if err != nil {
		return nil, fmt.Errorf("bank account not found: %w", err)
	}
	bankToID, err := s.accountRepo.FindAccountIDTx(tx, systemBankUserID, req.ToCurrency)
	if err != nil {
		return nil, fmt.Errorf("bank account not found: %w", err)
	}

	lockIDs := []uuid.UUID{userFromID, userToID, bankFromID, bankToID}
	sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })

	locked := make(map[uuid.UUID]*model.Account, 4)
	for _, id := range lockIDs {
		acc, err := s.accountRepo.LockAccountForUpdate(tx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to lock account: %w", err)
		}
		locked[id] = acc
	}

	fromAccount := locked[userFromID]
	toAccount := locked[userToID]
	bankFrom := locked[bankFromID]
	bankTo := locked[bankToID]
	if fromAccount == nil || toAccount == nil || bankFrom == nil || bankTo == nil {
		return nil, fmt.Errorf("failed to lock accounts")
	}

	if fromAccount.UserID != userID || toAccount.UserID != userID || bankFrom.UserID != systemBankUserID || bankTo.UserID != systemBankUserID {
		return nil, apperr.ErrUnauthorized
	}

	fromBalanceCents, err := amountToCents(fromAccount.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid from balance in db")
	}
	toBalanceCents, err := amountToCents(toAccount.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid to balance in db")
	}
	bankFromBalanceCents, err := amountToCents(bankFrom.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid bank balance in db")
	}
	bankToBalanceCents, err := amountToCents(bankTo.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid bank balance in db")
	}

	if fromBalanceCents < amountCents {
		s.logger.Warn("Insufficient funds for exchange", "user_id", userID, "balance_cents", fromBalanceCents, "amount_cents", amountCents)
		return nil, apperr.ErrInsufficientFunds
	}
	if bankToBalanceCents < convertedCents {
		s.logger.Error("Bank has insufficient liquidity", "currency", req.ToCurrency, "bank_balance_cents", bankToBalanceCents, "needed_cents", convertedCents)
		return nil, fmt.Errorf("exchange liquidity unavailable")
	}

	transactionID := uuid.New()
	now := time.Now()
	amountFloat := centsToFloat(amountCents)
	transaction := &model.Transaction{
		ID:              transactionID,
		Type:            model.TransactionTypeExchange,
		FromAccountID:   &fromAccount.ID,
		ToAccountID:     toAccount.ID,
		Amount:          amountFloat,
		Currency:        req.FromCurrency,
		ExchangeRate:    &exchangeRate,
		ConvertedAmount: &convertedAmount,
		Description:     fmt.Sprintf("Exchange %.2f %s to %.2f %s", amountFloat, req.FromCurrency, convertedAmount, req.ToCurrency),
		CreatedAt:       now,
	}

	if err := s.transactionRepo.Create(tx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	fromEntry := &model.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     fromAccount.ID,
		Amount:        -amountFloat,
		CreatedAt:     now,
	}
	if err := s.ledgerRepo.CreateEntry(tx, fromEntry); err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	bankFromEntry := &model.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     bankFrom.ID,
		Amount:        amountFloat,
		CreatedAt:     now,
	}
	if err := s.ledgerRepo.CreateEntry(tx, bankFromEntry); err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	bankToEntry := &model.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     bankTo.ID,
		Amount:        -convertedAmount,
		CreatedAt:     now,
	}
	if err := s.ledgerRepo.CreateEntry(tx, bankToEntry); err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	toEntry := &model.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     toAccount.ID,
		Amount:        convertedAmount,
		CreatedAt:     now,
	}
	if err := s.ledgerRepo.CreateEntry(tx, toEntry); err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	newFromBalanceCents := fromBalanceCents - amountCents
	newToBalanceCents := toBalanceCents + convertedCents
	newBankFromBalanceCents := bankFromBalanceCents + amountCents
	newBankToBalanceCents := bankToBalanceCents - convertedCents

	if err := s.accountRepo.UpdateBalanceString(tx, fromAccount.ID, centsToDecimalString(newFromBalanceCents)); err != nil {
		return nil, fmt.Errorf("failed to update from account balance: %w", err)
	}

	if err := s.accountRepo.UpdateBalanceString(tx, toAccount.ID, centsToDecimalString(newToBalanceCents)); err != nil {
		return nil, fmt.Errorf("failed to update to account balance: %w", err)
	}

	if err := s.accountRepo.UpdateBalanceString(tx, bankFrom.ID, centsToDecimalString(newBankFromBalanceCents)); err != nil {
		return nil, fmt.Errorf("failed to update bank from balance: %w", err)
	}

	if err := s.accountRepo.UpdateBalanceString(tx, bankTo.ID, centsToDecimalString(newBankToBalanceCents)); err != nil {
		return nil, fmt.Errorf("failed to update bank to balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit exchange transaction", "error", err, "transaction_id", transactionID)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Exchange completed successfully", "transaction_id", transactionID, "user_id", userID, "amount", req.Amount, "converted_amount", convertedAmount)

	user, _ := s.userRepo.GetByID(userID)

	response := &model.TransactionResponse{
		ID:              transaction.ID,
		Type:            transaction.Type,
		FromAccountID:   transaction.FromAccountID,
		ToAccountID:     transaction.ToAccountID,
		Amount:          transaction.Amount,
		Currency:        transaction.Currency,
		ExchangeRate:    transaction.ExchangeRate,
		ConvertedAmount: transaction.ConvertedAmount,
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt,
	}
	if user != nil {
		response.FromUserEmail = &user.Email
		response.ToUserEmail = &user.Email
	}

	return response, nil
}

func (s *TransactionService) GetUserTransactions(ctx context.Context, userID uuid.UUID, filter *model.TransactionFilter) ([]*model.TransactionResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}

	return s.transactionRepo.GetByUserID(userID, filter)
}

func amountToCents(amount float64) (int64, error) {
	cents := int64(amount*100 + 0.5)
	if centsToFloat(cents) != float64(int64(amount*100+0.5))/100 {
	}
	scaled := amount * 100
	nearest := float64(int64(scaled + 0.5))
	if diff := scaled - nearest; diff > 1e-9 || diff < -1e-9 {
		return 0, fmt.Errorf("amount has more than 2 decimals")
	}
	return cents, nil
}

func centsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}

func centsToDecimalString(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func convertExchange(amountCents int64, from model.Currency, to model.Currency) (float64, int64, error) {
	if from == model.CurrencyUSD && to == model.CurrencyEUR {
		converted := (amountCents*92 + 50) / 100
		return ExchangeRateUSDtoEUR, converted, nil
	}
	if from == model.CurrencyEUR && to == model.CurrencyUSD {
		converted := (amountCents*100 + 46) / 92
		return 1.0 / ExchangeRateUSDtoEUR, converted, nil
	}
	return 0, 0, fmt.Errorf("unsupported currency pair: %s to %s", from, to)
}
