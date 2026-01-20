package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"banking-platform/internal/apperr"
	"banking-platform/internal/model"
	"banking-platform/internal/storage"
	"github.com/google/uuid"
)

var systemBankUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type TransactionService struct {
	txRunner        storage.TxRunner
	accountRepo     storage.AccountRepo
	transactionRepo storage.TransactionRepo
	ledgerRepo      storage.LedgerRepo
	userRepo        storage.UserRepo
	logger          *slog.Logger

	exchangeRateUSDtoEURNum int64
	exchangeRateUSDtoEURDen int64
}

func NewTransactionService(
	txRunner storage.TxRunner,
	accountRepo storage.AccountRepo,
	transactionRepo storage.TransactionRepo,
	ledgerRepo storage.LedgerRepo,
	userRepo storage.UserRepo,
	exchangeRateUSDtoEUR string,
	logger *slog.Logger,
) *TransactionService {
	num, den := parseRateToFraction(exchangeRateUSDtoEUR)
	return &TransactionService{
		txRunner:                txRunner,
		accountRepo:             accountRepo,
		transactionRepo:         transactionRepo,
		ledgerRepo:              ledgerRepo,
		userRepo:                userRepo,
		logger:                  logger,
		exchangeRateUSDtoEURNum: num,
		exchangeRateUSDtoEURDen: den,
	}
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID uuid.UUID, req *model.TransferRequest) (*model.TransactionResponse, error) {
	var toUserID uuid.UUID
	if req.ToUserID != nil && req.ToUserEmail != nil {
		return nil, fmt.Errorf("provide either to_user_id or to_user_email")
	}
	if req.ToUserID != nil {
		toUserID = *req.ToUserID
	} else if req.ToUserEmail != nil {
		email := strings.ToLower(strings.TrimSpace(*req.ToUserEmail))
		if email == "" {
			return nil, fmt.Errorf("to_user_email cannot be empty")
		}
		u, err := s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("recipient not found")
		}
		toUserID = u.ID
	} else {
		return nil, fmt.Errorf("recipient is required")
	}
	if toUserID == fromUserID {
		return nil, fmt.Errorf("cannot transfer to self")
	}

	s.logger.Info("Processing transfer", "from_user_id", fromUserID, "to_user_id", toUserID, "amount", req.Amount, "currency", req.Currency)

	if req.Currency != model.CurrencyUSD && req.Currency != model.CurrencyEUR {
		s.logger.Warn("Invalid currency", "currency", req.Currency)
		return nil, fmt.Errorf("invalid currency: %s", req.Currency)
	}

	amountCents, err := amountToCents(req.Amount)
	if err != nil {
		s.logger.Warn("Invalid amount precision", "amount", req.Amount)
		return nil, apperr.ErrInvalidAmount
	}

	var created *model.Transaction
	var fromAccountID uuid.UUID
	var toAccountID uuid.UUID
	var createdAt time.Time

	if err := s.txRunner.WithTx(ctx, func(tx storage.Tx) error {
		var err error

		fromAccountID, err = s.accountRepo.FindAccountIDTx(ctx, tx, fromUserID, req.Currency)
		if err != nil {
			return fmt.Errorf("failed to get sender account: %w", err)
		}
		toAccountID, err = s.accountRepo.FindAccountIDTx(ctx, tx, toUserID, req.Currency)
		if err != nil {
			return fmt.Errorf("account not found: %w", err)
		}

		lockIDs := []uuid.UUID{fromAccountID, toAccountID}
		sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })

		locked := make(map[uuid.UUID]*model.Account, 2)
		for _, id := range lockIDs {
			acc, err := s.accountRepo.LockAccountForUpdate(ctx, tx, id)
			if err != nil {
				return fmt.Errorf("failed to lock account: %w", err)
			}
			locked[id] = acc
		}

		fromAccount := locked[fromAccountID]
		toAccount := locked[toAccountID]
		if fromAccount == nil || toAccount == nil {
			return fmt.Errorf("failed to lock accounts")
		}

		if fromAccount.UserID != fromUserID || toAccount.UserID != toUserID {
			return apperr.ErrUnauthorized
		}

		fromBalanceCents, err := amountToCents(fromAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid sender balance in db")
		}
		toBalanceCents, err := amountToCents(toAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid recipient balance in db")
		}

		if fromBalanceCents < amountCents {
			s.logger.Warn("Insufficient funds", "user_id", fromUserID, "balance_cents", fromBalanceCents, "amount_cents", amountCents)
			return apperr.ErrInsufficientFunds
		}

		transactionID := uuid.New()
		createdAt = time.Now()
		amountFloat := centsToFloat(amountCents)
		created = &model.Transaction{
			ID:            transactionID,
			Type:          model.TransactionTypeTransfer,
			FromAccountID: &fromAccount.ID,
			ToAccountID:   toAccount.ID,
			Amount:        amountFloat,
			Currency:      req.Currency,
			Description:   fmt.Sprintf("Transfer %s %.2f from %s to %s", req.Currency, amountFloat, fromUserID, toUserID),
			CreatedAt:     createdAt,
		}

		if err := s.transactionRepo.Create(ctx, tx, created); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		fromEntry := &model.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     fromAccount.ID,
			Amount:        -amountFloat,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, fromEntry); err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		toEntry := &model.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     toAccount.ID,
			Amount:        amountFloat,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, toEntry); err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		if err := s.ledgerRepo.VerifyTransactionBalanceTx(ctx, tx, transactionID); err != nil {
			s.logger.Error("Ledger not balanced (transfer)", "error", err, "transaction_id", transactionID)
			return err
		}

		newFromBalanceCents := fromBalanceCents - amountCents
		newToBalanceCents := toBalanceCents + amountCents

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, fromAccount.ID, centsToDecimalString(newFromBalanceCents)); err != nil {
			return fmt.Errorf("failed to update sender balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, toAccount.ID, centsToDecimalString(newToBalanceCents)); err != nil {
			return fmt.Errorf("failed to update recipient balance: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	s.logger.Info("Transfer completed successfully", "transaction_id", created.ID, "from_user_id", fromUserID, "to_user_id", toUserID, "amount", created.Amount, "currency", req.Currency)

	fromUser, _ := s.userRepo.GetByID(ctx, fromUserID)
	toUser, _ := s.userRepo.GetByID(ctx, toUserID)

	resp := &model.TransactionResponse{
		ID:            created.ID,
		Type:          created.Type,
		FromAccountID: created.FromAccountID,
		ToAccountID:   created.ToAccountID,
		Amount:        created.Amount,
		Currency:      created.Currency,
		Description:   created.Description,
		CreatedAt:     createdAt,
	}
	if fromUser != nil {
		resp.FromUserEmail = &fromUser.Email
	}
	if toUser != nil {
		resp.ToUserEmail = &toUser.Email
	}
	return resp, nil
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

	exchangeRate, convertedCents, err := convertExchange(amountCents, req.FromCurrency, req.ToCurrency, s.exchangeRateUSDtoEURNum, s.exchangeRateUSDtoEURDen)
	if err != nil {
		return nil, err
	}
	convertedAmount := centsToFloat(convertedCents)

	var created *model.Transaction
	var createdAt time.Time
	if err := s.txRunner.WithTx(ctx, func(tx storage.Tx) error {
		userFromID, err := s.accountRepo.FindAccountIDTx(ctx, tx, userID, req.FromCurrency)
		if err != nil {
			return fmt.Errorf("failed to get from account: %w", err)
		}
		userToID, err := s.accountRepo.FindAccountIDTx(ctx, tx, userID, req.ToCurrency)
		if err != nil {
			return fmt.Errorf("failed to get to account: %w", err)
		}
		bankFromID, err := s.accountRepo.FindAccountIDTx(ctx, tx, systemBankUserID, req.FromCurrency)
		if err != nil {
			return fmt.Errorf("bank account not found: %w", err)
		}
		bankToID, err := s.accountRepo.FindAccountIDTx(ctx, tx, systemBankUserID, req.ToCurrency)
		if err != nil {
			return fmt.Errorf("bank account not found: %w", err)
		}

		lockIDs := []uuid.UUID{userFromID, userToID, bankFromID, bankToID}
		sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })

		locked := make(map[uuid.UUID]*model.Account, 4)
		for _, id := range lockIDs {
			acc, err := s.accountRepo.LockAccountForUpdate(ctx, tx, id)
			if err != nil {
				return fmt.Errorf("failed to lock account: %w", err)
			}
			locked[id] = acc
		}

		fromAccount := locked[userFromID]
		toAccount := locked[userToID]
		bankFrom := locked[bankFromID]
		bankTo := locked[bankToID]
		if fromAccount == nil || toAccount == nil || bankFrom == nil || bankTo == nil {
			return fmt.Errorf("failed to lock accounts")
		}

		if fromAccount.UserID != userID || toAccount.UserID != userID || bankFrom.UserID != systemBankUserID || bankTo.UserID != systemBankUserID {
			return apperr.ErrUnauthorized
		}

		fromBalanceCents, err := amountToCents(fromAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid from balance in db")
		}
		toBalanceCents, err := amountToCents(toAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid to balance in db")
		}
		bankFromBalanceCents, err := amountToCents(bankFrom.Balance)
		if err != nil {
			return fmt.Errorf("invalid bank balance in db")
		}
		bankToBalanceCents, err := amountToCents(bankTo.Balance)
		if err != nil {
			return fmt.Errorf("invalid bank balance in db")
		}

		if fromBalanceCents < amountCents {
			s.logger.Warn("Insufficient funds for exchange", "user_id", userID, "balance_cents", fromBalanceCents, "amount_cents", amountCents)
			return apperr.ErrInsufficientFunds
		}
		if bankToBalanceCents < convertedCents {
			s.logger.Error("Bank has insufficient liquidity", "currency", req.ToCurrency, "bank_balance_cents", bankToBalanceCents, "needed_cents", convertedCents)
			return fmt.Errorf("exchange liquidity unavailable")
		}

		transactionID := uuid.New()
		createdAt = time.Now()
		amountFloat := centsToFloat(amountCents)
		created = &model.Transaction{
			ID:              transactionID,
			Type:            model.TransactionTypeExchange,
			FromAccountID:   &fromAccount.ID,
			ToAccountID:     toAccount.ID,
			Amount:          amountFloat,
			Currency:        req.FromCurrency,
			ExchangeRate:    &exchangeRate,
			ConvertedAmount: &convertedAmount,
			Description:     fmt.Sprintf("Exchange %.2f %s to %.2f %s", amountFloat, req.FromCurrency, convertedAmount, req.ToCurrency),
			CreatedAt:       createdAt,
		}

		if err := s.transactionRepo.Create(ctx, tx, created); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		fromEntry := &model.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     fromAccount.ID,
			Amount:        -amountFloat,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, fromEntry); err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		bankFromEntry := &model.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     bankFrom.ID,
			Amount:        amountFloat,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, bankFromEntry); err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		bankToEntry := &model.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     bankTo.ID,
			Amount:        -convertedAmount,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, bankToEntry); err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		toEntry := &model.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     toAccount.ID,
			Amount:        convertedAmount,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, toEntry); err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		if err := s.ledgerRepo.VerifyTransactionBalanceTx(ctx, tx, transactionID); err != nil {
			s.logger.Error("Ledger not balanced (exchange)", "error", err, "transaction_id", transactionID)
			return err
		}

		newFromBalanceCents := fromBalanceCents - amountCents
		newToBalanceCents := toBalanceCents + convertedCents
		newBankFromBalanceCents := bankFromBalanceCents + amountCents
		newBankToBalanceCents := bankToBalanceCents - convertedCents

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, fromAccount.ID, centsToDecimalString(newFromBalanceCents)); err != nil {
			return fmt.Errorf("failed to update from account balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, toAccount.ID, centsToDecimalString(newToBalanceCents)); err != nil {
			return fmt.Errorf("failed to update to account balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, bankFrom.ID, centsToDecimalString(newBankFromBalanceCents)); err != nil {
			return fmt.Errorf("failed to update bank from balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, bankTo.ID, centsToDecimalString(newBankToBalanceCents)); err != nil {
			return fmt.Errorf("failed to update bank to balance: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	s.logger.Info("Exchange completed successfully", "transaction_id", created.ID, "user_id", userID, "amount", req.Amount, "converted_amount", convertedAmount)

	user, _ := s.userRepo.GetByID(ctx, userID)

	response := &model.TransactionResponse{
		ID:              created.ID,
		Type:            created.Type,
		FromAccountID:   created.FromAccountID,
		ToAccountID:     created.ToAccountID,
		Amount:          created.Amount,
		Currency:        created.Currency,
		ExchangeRate:    created.ExchangeRate,
		ConvertedAmount: created.ConvertedAmount,
		Description:     created.Description,
		CreatedAt:       createdAt,
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

	return s.transactionRepo.GetByUserID(ctx, userID, filter)
}

func amountToCents(amount float64) (int64, error) {
	cents := int64(amount*100 + 0.5)

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

func parseRateToFraction(raw string) (int64, int64) {
	r := strings.TrimSpace(raw)
	if r == "" {
		return 92, 100
	}
	f, err := strconv.ParseFloat(r, 64)
	if err != nil || f <= 0 {
		return 92, 100
	}
	// Support up to 6 decimals from env: scale to an integer fraction.
	scale := int64(1_000_000)
	n := int64(f*float64(scale) + 0.5)
	if n <= 0 {
		return 92, 100
	}
	return n, scale
}

func convertExchange(amountCents int64, from model.Currency, to model.Currency, rateNum int64, rateDen int64) (float64, int64, error) {
	if rateNum <= 0 || rateDen <= 0 {
		return 0, 0, fmt.Errorf("invalid exchange rate")
	}

	if from == model.CurrencyUSD && to == model.CurrencyEUR {
		converted := (amountCents*rateNum + rateDen/2) / rateDen
		return float64(rateNum) / float64(rateDen), converted, nil
	}
	if from == model.CurrencyEUR && to == model.CurrencyUSD {
		converted := (amountCents*rateDen + rateNum/2) / rateNum
		return float64(rateDen) / float64(rateNum), converted, nil
	}
	return 0, 0, fmt.Errorf("unsupported currency pair: %s to %s", from, to)
}
