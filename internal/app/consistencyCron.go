package app

import (
	"context"
	"log/slog"
	"time"

	"banking-platform/config"
	"banking-platform/internal/service"
)

type consistencyCron struct {
	cancel context.CancelFunc
	done   chan struct{}
}

func startConsistencyCron(cfg *config.Config, logger *slog.Logger, checker *service.LedgerConsistencyService) *consistencyCron {
	if !cfg.ConsistencyCronEnabled {
		logger.Info("Ledger consistency cron disabled")
		return nil
	}

	interval := cfg.ConsistencyCronInterval
	timeout := cfg.ConsistencyCronTimeout

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	runOnce := func() {
		runCtx, runCancel := context.WithTimeout(ctx, timeout)
		err := checker.CheckLedgerBalance(runCtx, 100)
		runCancel()
		if err != nil {
			logger.Error("Ledger consistency check failed", "error", err)
		}

		runCtx2, runCancel2 := context.WithTimeout(ctx, timeout)
		err2 := checker.CheckAccountBalanceConsistency(runCtx2, 100)
		runCancel2()
		if err2 != nil {
			logger.Error("Account balance consistency check failed", "error", err2)
		}
	}

	go func() {
		defer close(done)
		logger.Info("Ledger consistency cron started", "interval", interval.String(), "timeout", timeout.String())
		runOnce()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runOnce()
			case <-ctx.Done():
				logger.Info("Ledger consistency cron stopped")
				return
			}
		}
	}()

	return &consistencyCron{cancel: cancel, done: done}
}

func (c *consistencyCron) Stop(ctx context.Context) {
	if c == nil {
		return
	}
	if c.cancel != nil {
		c.cancel()
	}
	if c.done == nil {
		return
	}
	select {
	case <-c.done:
	case <-ctx.Done():
	}
}
