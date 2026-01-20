package cron

import (
	"context"
	"log/slog"
	"time"

	"banking-platform/config"
	"banking-platform/internal/service"
)

// ConsistencyCron runs periodic consistency checks in background.
type ConsistencyCron struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// StartConsistencyCron starts periodic checks if enabled in config.
func StartConsistencyCron(cfg *config.Config, logger *slog.Logger, checker *service.LedgerConsistencyService) *ConsistencyCron {
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

		runCtxConsis, runCancelConsis := context.WithTimeout(ctx, timeout)
		errConsis := checker.CheckAccountBalanceConsistency(runCtxConsis, 100)
		runCancelConsis()
		if errConsis != nil {
			logger.Error("Account balance consistency check failed", "error", errConsis)
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

	return &ConsistencyCron{cancel: cancel, done: done}
}

// Stop signals the cron to stop and waits until it finishes (or ctx is done).
func (c *ConsistencyCron) Stop(ctx context.Context) {
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
