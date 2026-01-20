package service

import (
	"context"
	"fmt"
	"log/slog"

	"banking-platform/internal/storage"
)

type LedgerConsistencyService struct {
	ledgerRepo *storage.LedgerRepository
	logger     *slog.Logger
}

func NewLedgerConsistencyService(ledgerRepo *storage.LedgerRepository, logger *slog.Logger) *LedgerConsistencyService {
	return &LedgerConsistencyService{
		ledgerRepo: ledgerRepo,
		logger:     logger,
	}
}

func (s *LedgerConsistencyService) CheckLedgerBalance(ctx context.Context, limit int) error {
	ids, err := s.ledgerRepo.FindUnbalancedTransactionIDs(ctx, limit)
	if err != nil {
		return fmt.Errorf("ledger consistency check failed: %w", err)
	}
	if len(ids) == 0 {
		s.logger.Info("Ledger consistency check OK: no unbalanced transactions")
		return nil
	}
	s.logger.Error("Ledger consistency check FAILED: unbalanced transactions found", "count", len(ids), "transaction_ids", ids)
	return fmt.Errorf("unbalanced transactions found: %d", len(ids))
}

func (s *LedgerConsistencyService) CheckAccountBalanceConsistency(ctx context.Context, limit int) error {
	mismatches, err := s.ledgerRepo.FindAccountBalanceMismatches(ctx, limit)
	if err != nil {
		return fmt.Errorf("account balance consistency check failed: %w", err)
	}
	if len(mismatches) == 0 {
		s.logger.Info("Account balance consistency check OK: no mismatches")
		return nil
	}
	s.logger.Error("Account balance consistency check FAILED: mismatches found", "count", len(mismatches), "mismatches", mismatches)
	return fmt.Errorf("account balance mismatches found: %d", len(mismatches))
}
