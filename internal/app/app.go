package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"banking-platform/config"
	"banking-platform/internal/hash"
	"banking-platform/internal/jwt"
	"banking-platform/internal/server"
	"banking-platform/internal/service"
	"banking-platform/internal/storage"
)

type App struct {
	server *server.Server
}

func NewApp() (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := storage.NewDB(cfg.DatabaseURL())
	if err != nil {
		return nil, err
	}

	userRepo := storage.NewUserRepository(db)
	accountRepo := storage.NewAccountRepository(db)
	transactionRepo := storage.NewTransactionRepository(db)
	ledgerRepo := storage.NewLedgerRepository(db)
	refreshTokenRepo := storage.NewRefreshTokenRepository(db)

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
		logger,
	)

	srv := server.NewServer(
		db,
		authService,
		accountService,
		transactionService,
	)

	return &App{
		server: srv,
	}, nil
}

func (a *App) Run() error {
	cfg, _ := config.Load()
	return a.server.Start(cfg.Port)
}

func (a *App) Close() error {
	return a.server.Close()
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	return a.server.Close()
}
