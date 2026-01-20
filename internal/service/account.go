package service

import (
	"context"
	"fmt"
	"log/slog"

	"banking-platform/internal/apperr"
	"banking-platform/internal/model"
	"banking-platform/internal/storage"
	"github.com/google/uuid"
)

type AccountService struct {
	accountRepo storage.AccountRepo
	logger      *slog.Logger
}

func NewAccountService(accountRepo storage.AccountRepo, logger *slog.Logger) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
		logger:      logger,
	}
}

func (s *AccountService) GetUserAccounts(ctx context.Context, userID uuid.UUID) ([]*model.AccountResponse, error) {
	s.logger.Info("Getting user accounts", "user_id", userID)

	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get accounts", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	responses := make([]*model.AccountResponse, len(accounts))
	for i, acc := range accounts {
		responses[i] = &model.AccountResponse{
			ID:       acc.ID,
			Currency: acc.Currency,
			Balance:  acc.Balance,
		}
	}

	s.logger.Info("Retrieved user accounts", "user_id", userID, "count", len(responses))
	return responses, nil
}

func (s *AccountService) GetAccountBalance(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (float64, error) {
	s.logger.Info("Getting account balance", "account_id", accountID, "user_id", userID)

	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		s.logger.Warn("Account not found", "account_id", accountID)
		return 0, apperr.ErrAccountNotFound
	}

	if account.UserID != userID {
		s.logger.Warn("Unauthorized access to account", "account_id", accountID, "user_id", userID)
		return 0, apperr.ErrUnauthorized
	}

	s.logger.Info("Retrieved account balance", "account_id", accountID, "balance", account.Balance)
	return account.Balance, nil
}
