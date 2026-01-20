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
	"banking-platform/internal/domain"
	"banking-platform/internal/http/dto"
	"github.com/google/uuid"
)

var systemBankUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type TransactionService struct {
	txRunner        TxRunner
	accountRepo     AccountRepo
	transactionRepo TransactionRepo
	ledgerRepo      LedgerRepo
	userRepo        UserRepo
	logger          *slog.Logger

	exchangeRateUSDtoEURNum int64
	exchangeRateUSDtoEURDen int64
}

// Money is cents; balance changes are transactional; each transaction must be ledger-balanced.

func NewTransactionService(
	txRunner TxRunner,
	accountRepo AccountRepo,
	transactionRepo TransactionRepo,
	ledgerRepo LedgerRepo,
	userRepo UserRepo,
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

func (s *TransactionService) Transfer(ctx context.Context, fromUserID uuid.UUID, req *dto.TransferRequest) (*dto.TransactionResponse, error) {
	var toUserID uuid.UUID
	if req.ToUserID != nil && req.ToUserEmail != nil {
		return nil, apperr.BadRequest("provide either to_user_id or to_user_email")
	}
	if req.ToUserID != nil {
		toUserID = *req.ToUserID
	} else if req.ToUserEmail != nil {
		email := strings.ToLower(strings.TrimSpace(*req.ToUserEmail))
		if email == "" {
			return nil, apperr.BadRequest("to_user_email cannot be empty")
		}
		u, err := s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("transaction.transfer: get recipient by email: %w", err)
		}
		toUserID = u.ID
	} else {
		return nil, apperr.BadRequest("recipient is required")
	}
	if toUserID == fromUserID {
		return nil, apperr.ErrCannotTransferToSelf
	}

	s.logger.Info("Processing transfer", "from_user_id", fromUserID, "to_user_id", toUserID, "amount_cents", req.AmountCents, "currency", req.Currency)

	if req.Currency != domain.CurrencyUSD && req.Currency != domain.CurrencyEUR {
		s.logger.Warn("Invalid currency", "currency", req.Currency)
		return nil, apperr.ErrInvalidCurrency
	}

	amountCents := req.AmountCents

	var created *domain.Transaction
	var fromAccountID uuid.UUID
	var toAccountID uuid.UUID
	var createdAt time.Time

	if err := s.txRunner.WithTx(ctx, func(tx Tx) error {
		var err error

		fromAccountID, err = s.accountRepo.FindAccountIDTx(ctx, tx, fromUserID, req.Currency)
		if err != nil {
			return fmt.Errorf("transaction.transfer: find sender account: %w", err)
		}
		toAccountID, err = s.accountRepo.FindAccountIDTx(ctx, tx, toUserID, req.Currency)
		if err != nil {
			return fmt.Errorf("transaction.transfer: find recipient account: %w", err)
		}

		// Lock deterministically to avoid deadlocks.
		lockIDs := []uuid.UUID{fromAccountID, toAccountID}
		sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })

		locked := make(map[uuid.UUID]*domain.Account, 2)
		for _, id := range lockIDs {
			acc, err := s.accountRepo.LockAccountForUpdate(ctx, tx, id)
			if err != nil {
				return fmt.Errorf("transaction.transfer: lock account: %w", err)
			}
			locked[id] = acc
		}

		fromAccount := locked[fromAccountID]
		toAccount := locked[toAccountID]
		if fromAccount == nil || toAccount == nil {
			return fmt.Errorf("transaction.transfer: failed to lock accounts")
		}

		if fromAccount.UserID != fromUserID || toAccount.UserID != toUserID {
			return apperr.ErrUnauthorized
		}

		fromBalanceCents := fromAccount.BalanceCents
		toBalanceCents := toAccount.BalanceCents

		if fromBalanceCents < amountCents {
			s.logger.Warn("Insufficient funds", "user_id", fromUserID, "balance_cents", fromBalanceCents, "amount_cents", amountCents)
			return apperr.ErrInsufficientFunds
		}

		transactionID := uuid.New()
		createdAt = time.Now()
		amountStr := domain.CentsToDecimalString(amountCents)
		created = &domain.Transaction{
			ID:            transactionID,
			Type:          domain.TransactionTypeTransfer,
			FromAccountID: &fromAccount.ID,
			ToAccountID:   toAccount.ID,
			AmountCents:   amountCents,
			Currency:      req.Currency,
			Description:   fmt.Sprintf("Transfer %s %s from %s to %s", req.Currency, amountStr, fromUserID, toUserID),
			CreatedAt:     createdAt,
		}

		if err := s.transactionRepo.Create(ctx, tx, created); err != nil {
			return fmt.Errorf("transaction.transfer: create transaction: %w", err)
		}

		fromEntry := &domain.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     fromAccount.ID,
			AmountCents:   -amountCents,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, fromEntry); err != nil {
			return fmt.Errorf("transaction.transfer: create ledger entry (from): %w", err)
		}

		toEntry := &domain.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     toAccount.ID,
			AmountCents:   amountCents,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, toEntry); err != nil {
			return fmt.Errorf("transaction.transfer: create ledger entry (to): %w", err)
		}

		if err := s.ledgerRepo.VerifyTransactionBalanceTx(ctx, tx, transactionID); err != nil {
			s.logger.Error("Ledger not balanced (transfer)", "error", err, "transaction_id", transactionID)
			return err
		}

		newFromBalanceCents := fromBalanceCents - amountCents
		newToBalanceCents := toBalanceCents + amountCents

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, fromAccount.ID, domain.CentsToDecimalString(newFromBalanceCents)); err != nil {
			return fmt.Errorf("transaction.transfer: update sender balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, toAccount.ID, domain.CentsToDecimalString(newToBalanceCents)); err != nil {
			return fmt.Errorf("transaction.transfer: update recipient balance: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	s.logger.Info("Transfer completed successfully", "transaction_id", created.ID, "from_user_id", fromUserID, "to_user_id", toUserID, "amount_cents", created.AmountCents, "currency", req.Currency)

	fromUser, _ := s.userRepo.GetByID(ctx, fromUserID)
	toUser, _ := s.userRepo.GetByID(ctx, toUserID)

	resp := &dto.TransactionResponse{
		ID:            created.ID,
		Type:          created.Type,
		FromAccountID: created.FromAccountID,
		ToAccountID:   created.ToAccountID,
		AmountCents:   created.AmountCents,
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

func (s *TransactionService) Exchange(ctx context.Context, userID uuid.UUID, req *dto.ExchangeRequest) (*dto.TransactionResponse, error) {
	s.logger.Info("Processing exchange", "user_id", userID, "from_currency", req.FromCurrency, "to_currency", req.ToCurrency, "amount_cents", req.AmountCents)

	if req.FromCurrency == req.ToCurrency {
		s.logger.Warn("Same currency for exchange", "currency", req.FromCurrency)
		return nil, apperr.ErrCurrenciesMustDiffer
	}

	amountCents := req.AmountCents

	exchangeRate, convertedCents, err := convertExchange(amountCents, req.FromCurrency, req.ToCurrency, s.exchangeRateUSDtoEURNum, s.exchangeRateUSDtoEURDen)
	if err != nil {
		return nil, fmt.Errorf("transaction.exchange: convert: %w", err)
	}
	convertedAmountCents := convertedCents

	var created *domain.Transaction
	var createdAt time.Time
	if err := s.txRunner.WithTx(ctx, func(tx Tx) error {
		userFromID, err := s.accountRepo.FindAccountIDTx(ctx, tx, userID, req.FromCurrency)
		if err != nil {
			return fmt.Errorf("transaction.exchange: find user from account: %w", err)
		}
		userToID, err := s.accountRepo.FindAccountIDTx(ctx, tx, userID, req.ToCurrency)
		if err != nil {
			return fmt.Errorf("transaction.exchange: find user to account: %w", err)
		}
		bankFromID, err := s.accountRepo.FindAccountIDTx(ctx, tx, systemBankUserID, req.FromCurrency)
		if err != nil {
			return fmt.Errorf("transaction.exchange: find bank from account: %w", err)
		}
		bankToID, err := s.accountRepo.FindAccountIDTx(ctx, tx, systemBankUserID, req.ToCurrency)
		if err != nil {
			return fmt.Errorf("transaction.exchange: find bank to account: %w", err)
		}

		// Lock deterministically to avoid deadlocks.
		lockIDs := []uuid.UUID{userFromID, userToID, bankFromID, bankToID}
		sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })

		locked := make(map[uuid.UUID]*domain.Account, 4)
		for _, id := range lockIDs {
			acc, err := s.accountRepo.LockAccountForUpdate(ctx, tx, id)
			if err != nil {
				return fmt.Errorf("transaction.exchange: lock account: %w", err)
			}
			locked[id] = acc
		}

		fromAccount := locked[userFromID]
		toAccount := locked[userToID]
		bankFrom := locked[bankFromID]
		bankTo := locked[bankToID]
		if fromAccount == nil || toAccount == nil || bankFrom == nil || bankTo == nil {
			return fmt.Errorf("transaction.exchange: failed to lock accounts")
		}

		if fromAccount.UserID != userID || toAccount.UserID != userID || bankFrom.UserID != systemBankUserID || bankTo.UserID != systemBankUserID {
			return apperr.ErrUnauthorized
		}

		fromBalanceCents := fromAccount.BalanceCents
		toBalanceCents := toAccount.BalanceCents
		bankFromBalanceCents := bankFrom.BalanceCents
		bankToBalanceCents := bankTo.BalanceCents

		if fromBalanceCents < amountCents {
			s.logger.Warn("Insufficient funds for exchange", "user_id", userID, "balance_cents", fromBalanceCents, "amount_cents", amountCents)
			return apperr.ErrInsufficientFunds
		}
		if bankToBalanceCents < convertedAmountCents {
			s.logger.Error("Bank has insufficient liquidity", "currency", req.ToCurrency, "bank_balance_cents", bankToBalanceCents, "needed_cents", convertedCents)
			return apperr.ErrLiquidityUnavailable
		}

		transactionID := uuid.New()
		createdAt = time.Now()
		amountStr := domain.CentsToDecimalString(amountCents)
		convertedStr := domain.CentsToDecimalString(convertedAmountCents)
		created = &domain.Transaction{
			ID:                   transactionID,
			Type:                 domain.TransactionTypeExchange,
			FromAccountID:        &fromAccount.ID,
			ToAccountID:          toAccount.ID,
			AmountCents:          amountCents,
			Currency:             req.FromCurrency,
			ExchangeRate:         &exchangeRate,
			ConvertedAmountCents: &convertedAmountCents,
			Description:          fmt.Sprintf("Exchange %s %s to %s %s", amountStr, req.FromCurrency, convertedStr, req.ToCurrency),
			CreatedAt:            createdAt,
		}

		if err := s.transactionRepo.Create(ctx, tx, created); err != nil {
			return fmt.Errorf("transaction.exchange: create transaction: %w", err)
		}

		fromEntry := &domain.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     fromAccount.ID,
			AmountCents:   -amountCents,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, fromEntry); err != nil {
			return fmt.Errorf("transaction.exchange: create ledger entry (user from): %w", err)
		}

		bankFromEntry := &domain.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     bankFrom.ID,
			AmountCents:   amountCents,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, bankFromEntry); err != nil {
			return fmt.Errorf("transaction.exchange: create ledger entry (bank from): %w", err)
		}

		bankToEntry := &domain.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     bankTo.ID,
			AmountCents:   -convertedAmountCents,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, bankToEntry); err != nil {
			return fmt.Errorf("transaction.exchange: create ledger entry (bank to): %w", err)
		}

		toEntry := &domain.LedgerEntry{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     toAccount.ID,
			AmountCents:   convertedAmountCents,
			CreatedAt:     createdAt,
		}
		if err := s.ledgerRepo.CreateEntry(ctx, tx, toEntry); err != nil {
			return fmt.Errorf("transaction.exchange: create ledger entry (user to): %w", err)
		}

		if err := s.ledgerRepo.VerifyTransactionBalanceTx(ctx, tx, transactionID); err != nil {
			s.logger.Error("Ledger not balanced (exchange)", "error", err, "transaction_id", transactionID)
			return err
		}

		newFromBalanceCents := fromBalanceCents - amountCents
		newToBalanceCents := toBalanceCents + convertedAmountCents
		newBankFromBalanceCents := bankFromBalanceCents + amountCents
		newBankToBalanceCents := bankToBalanceCents - convertedAmountCents

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, fromAccount.ID, domain.CentsToDecimalString(newFromBalanceCents)); err != nil {
			return fmt.Errorf("transaction.exchange: update user from balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, toAccount.ID, domain.CentsToDecimalString(newToBalanceCents)); err != nil {
			return fmt.Errorf("transaction.exchange: update user to balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, bankFrom.ID, domain.CentsToDecimalString(newBankFromBalanceCents)); err != nil {
			return fmt.Errorf("transaction.exchange: update bank from balance: %w", err)
		}

		if err := s.accountRepo.UpdateBalanceString(ctx, tx, bankTo.ID, domain.CentsToDecimalString(newBankToBalanceCents)); err != nil {
			return fmt.Errorf("transaction.exchange: update bank to balance: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	s.logger.Info("Exchange completed successfully", "transaction_id", created.ID, "user_id", userID, "amount_cents", req.AmountCents, "converted_amount_cents", convertedAmountCents)

	user, _ := s.userRepo.GetByID(ctx, userID)

	response := &dto.TransactionResponse{
		ID:                   created.ID,
		Type:                 created.Type,
		FromAccountID:        created.FromAccountID,
		ToAccountID:          created.ToAccountID,
		AmountCents:          created.AmountCents,
		Currency:             created.Currency,
		ExchangeRate:         created.ExchangeRate,
		ConvertedAmountCents: created.ConvertedAmountCents,
		Description:          created.Description,
		CreatedAt:            createdAt,
	}
	if user != nil {
		response.FromUserEmail = &user.Email
		response.ToUserEmail = &user.Email
	}

	return response, nil
}

func (s *TransactionService) GetUserTransactions(ctx context.Context, userID uuid.UUID, filter *dto.TransactionFilter) ([]*dto.TransactionResponse, error) {
	f := &domain.TransactionFilter{
		Type:  filter.Type,
		Page:  filter.Page,
		Limit: filter.Limit,
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 50
	}

	items, err := s.transactionRepo.GetByUserID(ctx, userID, f)
	if err != nil {
		return nil, fmt.Errorf("transaction.list: %w", err)
	}
	out := make([]*dto.TransactionResponse, 0, len(items))
	for _, it := range items {
		tx := it.Transaction
		out = append(out, &dto.TransactionResponse{
			ID:                  tx.ID,
			Type:                tx.Type,
			FromAccountID:        tx.FromAccountID,
			ToAccountID:          tx.ToAccountID,
			AmountCents:          tx.AmountCents,
			Currency:             tx.Currency,
			ExchangeRate:         tx.ExchangeRate,
			ConvertedAmountCents: tx.ConvertedAmountCents,
			Description:          tx.Description,
			CreatedAt:            tx.CreatedAt,
			FromUserEmail:        it.FromUserEmail,
			ToUserEmail:          it.ToUserEmail,
		})
	}
	return out, nil
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
	scale := int64(1_000_000)
	n := int64(f*float64(scale) + 0.5)
	if n <= 0 {
		return 92, 100
	}
	return n, scale
}

func convertExchange(amountCents int64, from domain.Currency, to domain.Currency, rateNum int64, rateDen int64) (float64, int64, error) {
	if rateNum <= 0 || rateDen <= 0 {
		return 0, 0, fmt.Errorf("invalid exchange rate")
	}

	if from == domain.CurrencyUSD && to == domain.CurrencyEUR {
		converted := (amountCents*rateNum + rateDen/2) / rateDen
		return float64(rateNum) / float64(rateDen), converted, nil
	}
	if from == domain.CurrencyEUR && to == domain.CurrencyUSD {
		converted := (amountCents*rateDen + rateNum/2) / rateNum
		return float64(rateDen) / float64(rateNum), converted, nil
	}
	return 0, 0, apperr.ErrInvalidCurrency
}
