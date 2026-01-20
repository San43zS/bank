package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"banking-platform/config"
	"banking-platform/internal/cron"
	"banking-platform/internal/jwt"
	"banking-platform/internal/repo"
	"banking-platform/internal/server"
	"banking-platform/internal/service"
	"banking-platform/pkg/hash"
	"banking-platform/pkg/logger"
)

type App struct {
	cfg    *config.Config
	server *server.Server
	cron   *cron.ConsistencyCron
}

func NewApp() (*App, error) {
	logger := logger.NewJSON(os.Stdout, slog.LevelInfo)
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := repo.NewDB(cfg.DatabaseURL())
	if err != nil {
		return nil, err
	}

	userRepo := repo.NewUserRepository(db)
	accountRepo := repo.NewAccountRepository(db)
	transactionRepo := repo.NewTransactionRepository(db)
	ledgerRepo := repo.NewLedgerRepository(db)
	refreshTokenRepo := repo.NewRefreshTokenRepository(db)

	ledgerConsistencyService := service.NewLedgerConsistencyService(ledgerRepo, logger)

	accessTokenTTL := 15 * time.Minute
	refreshTokenTTL := 7 * 24 * time.Hour
	tokenManager := jwt.NewTokenManager(cfg.JWTSecret, accessTokenTTL, refreshTokenTTL)

	hasher := hash.NewHasher()

	authService := service.NewAuthService(
		userRepo,
		accountRepo,
		refreshTokenRepo,
		tokenManager,
		hasher,
		logger,
	)
	accountService := service.NewAccountService(accountRepo, logger)
	transactionService := service.NewTransactionService(
		db,
		accountRepo,
		transactionRepo,
		ledgerRepo,
		userRepo,
		cfg.ExchangeRateUSDtoEUR,
		logger,
	)

	cronJob := cron.StartConsistencyCron(cfg, logger, ledgerConsistencyService)

	srv := server.NewServer(
		cfg,
		db,
		authService,
		accountService,
		transactionService,
	)

	return &App{
		cfg:    cfg,
		server: srv,
		cron:   cronJob,
	}, nil
}

func (a *App) Run() error {
	return a.server.Start(a.cfg.Port)
}

func (a *App) Close() error {
	if a.cron != nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.cfg.CronStopTimeout)
		a.cron.Stop(ctx)
		cancel()
	}
	return a.server.Close()
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.cron != nil {
		a.cron.Stop(ctx)
	}
	shutdownErr := a.server.Shutdown(ctx)
	closeErr := a.server.Close()
	if shutdownErr != nil {
		return shutdownErr
	}
	return closeErr
}

// ShutdownTimeout returns the configured graceful shutdown timeout.
func (a *App) ShutdownTimeout() time.Duration {
	return a.cfg.ShutdownTimeout
}
