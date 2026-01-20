package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"banking-platform/internal/apperr"
	"banking-platform/internal/domain"
	"banking-platform/internal/jwt"
	"banking-platform/pkg/hash"
)

type AuthService struct {
	userRepo         UserRepo
	accountRepo      AccountRepo
	txRunner         TxRunner
	transactionRepo  TransactionRepo
	ledgerRepo       LedgerRepo
	refreshTokenRepo RefreshTokenRepo
	tokenManager     *jwt.TokenManager
	hasher           *hash.Hasher
	logger           *slog.Logger
}

func NewAuthService(
	userRepo UserRepo,
	accountRepo AccountRepo,
	txRunner TxRunner,
	transactionRepo TransactionRepo,
	ledgerRepo LedgerRepo,
	refreshTokenRepo RefreshTokenRepo,
	tokenManager *jwt.TokenManager,
	hasher *hash.Hasher,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		accountRepo:      accountRepo,
		txRunner:         txRunner,
		transactionRepo:  transactionRepo,
		ledgerRepo:       ledgerRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenManager:     tokenManager,
		hasher:           hasher,
		logger:           logger,
	}
}

// Register creates a user and returns a token pair.
func (s *AuthService) Register(ctx context.Context, in *domain.RegisterInput) (*domain.AuthResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	s.logger.Info("Registering new user", "email", in.Email)

	_, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err == nil {
		s.logger.Warn("User already exists", "email", in.Email)
		return nil, apperr.ErrUserExists
	}
	if err != nil && !errors.Is(err, apperr.ErrUserNotFound) {
		return nil, fmt.Errorf("auth.register: get user by email: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        in.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("auth.register: create user: %w", err)
	}

	usdAccount, eurAccount, err := s.createDefaultAccounts(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to create default accounts", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("auth.register: create default accounts: %w", err)
	}

	if err := s.fundInitialBalancesViaLedger(ctx, user.ID, usdAccount.ID, eurAccount.ID); err != nil {
		s.logger.Error("Failed to fund initial balances via ledger", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("auth.register: fund initial balances: %w", err)
	}

	tokenPair, err := s.generateTokenPair(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("auth.register: generate tokens: %w", err)
	}

	s.logger.Info("User registered successfully", "user_id", user.ID, "email", in.Email)

	return &domain.AuthResult{
		Tokens: *tokenPair,
		User: domain.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// Login validates credentials and returns a token pair.
func (s *AuthService) Login(ctx context.Context, in *domain.LoginInput) (*domain.AuthResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	s.logger.Info("User login attempt", "email", in.Email)

	user, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			s.logger.Warn("User not found", "email", in.Email)
			return nil, apperr.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth.login: get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		s.logger.Warn("Invalid password", "email", in.Email)
		return nil, apperr.ErrInvalidCredentials
	}

	tokenPair, err := s.generateTokenPair(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("auth.login: generate tokens: %w", err)
	}

	s.logger.Info("User logged in successfully", "user_id", user.ID, "email", in.Email)

	return &domain.AuthResult{
		Tokens: *tokenPair,
		User: domain.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (s *AuthService) createDefaultAccounts(ctx context.Context, userID uuid.UUID) (*domain.Account, *domain.Account, error) {
	now := time.Now()

	usdAccount := &domain.Account{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     domain.CurrencyUSD,
		BalanceCents: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.accountRepo.Create(ctx, usdAccount); err != nil {
		s.logger.Error("Failed to create USD account", "error", err, "user_id", userID)
		return nil, nil, err
	}

	eurAccount := &domain.Account{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     domain.CurrencyEUR,
		BalanceCents: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.accountRepo.Create(ctx, eurAccount); err != nil {
		s.logger.Error("Failed to create EUR account", "error", err, "user_id", userID)
		return nil, nil, err
	}

	s.logger.Info("Default accounts created", "user_id", userID)
	return usdAccount, eurAccount, nil
}

func (s *AuthService) fundInitialBalancesViaLedger(ctx context.Context, userID uuid.UUID, userUSDAccountID uuid.UUID, userEURAccountID uuid.UUID) error {
	const initialUSDCents int64 = 1000_00
	const initialEURCents int64 = 500_00

	return s.txRunner.WithTx(ctx, func(tx Tx) error {
		bankUSDAccountID, err := s.accountRepo.FindAccountIDTx(ctx, tx, systemBankUserID, domain.CurrencyUSD)
		if err != nil {
			return fmt.Errorf("find bank USD account: %w", err)
		}
		bankEURAccountID, err := s.accountRepo.FindAccountIDTx(ctx, tx, systemBankUserID, domain.CurrencyEUR)
		if err != nil {
			return fmt.Errorf("find bank EUR account: %w", err)
		}

		if err := s.createFundingTransferTx(ctx, tx, bankUSDAccountID, userUSDAccountID, domain.CurrencyUSD, initialUSDCents, "Initial USD funding"); err != nil {
			return fmt.Errorf("initial USD funding: %w", err)
		}
		if err := s.createFundingTransferTx(ctx, tx, bankEURAccountID, userEURAccountID, domain.CurrencyEUR, initialEURCents, "Initial EUR funding"); err != nil {
			return fmt.Errorf("initial EUR funding: %w", err)
		}
		return nil
	})
}

func (s *AuthService) createFundingTransferTx(ctx context.Context, tx Tx, fromAccountID uuid.UUID, toAccountID uuid.UUID, currency domain.Currency, amountCents int64, purpose string) error {
	if amountCents <= 0 {
		return apperr.BadRequest("amount must be greater than 0")
	}

	// Lock deterministically to reduce deadlock probability.
	lockIDs := []uuid.UUID{fromAccountID, toAccountID}
	sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i].String() < lockIDs[j].String() })

	locked := make(map[uuid.UUID]*domain.Account, 2)
	for _, id := range lockIDs {
		acc, err := s.accountRepo.LockAccountForUpdate(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("lock account: %w", err)
		}
		locked[id] = acc
	}

	fromAccount := locked[fromAccountID]
	toAccount := locked[toAccountID]
	if fromAccount == nil || toAccount == nil {
		return fmt.Errorf("failed to lock accounts")
	}
	if fromAccount.Currency != currency || toAccount.Currency != currency {
		return apperr.ErrInvalidCurrency
	}
	if fromAccount.BalanceCents < amountCents {
		return apperr.ErrLiquidityUnavailable
	}

	transactionID := uuid.New()
	createdAt := time.Now()
	amountStr := domain.CentsToDecimalString(amountCents)
	created := &domain.Transaction{
		ID:            transactionID,
		Type:          domain.TransactionTypeTransfer,
		FromAccountID: &fromAccountID,
		ToAccountID:   toAccountID,
		AmountCents:   amountCents,
		Currency:      currency,
		Description:   fmt.Sprintf("%s: %s %s", purpose, currency, amountStr),
		CreatedAt:     createdAt,
	}

	if err := s.transactionRepo.Create(ctx, tx, created); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	fromEntry := &domain.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     fromAccountID,
		AmountCents:   -amountCents,
		CreatedAt:     createdAt,
	}
	if err := s.ledgerRepo.CreateEntry(ctx, tx, fromEntry); err != nil {
		return fmt.Errorf("create ledger entry (from): %w", err)
	}
	toEntry := &domain.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     toAccountID,
		AmountCents:   amountCents,
		CreatedAt:     createdAt,
	}
	if err := s.ledgerRepo.CreateEntry(ctx, tx, toEntry); err != nil {
		return fmt.Errorf("create ledger entry (to): %w", err)
	}

	if err := s.ledgerRepo.VerifyTransactionBalanceTx(ctx, tx, transactionID); err != nil {
		return err
	}

	newFrom := fromAccount.BalanceCents - amountCents
	newTo := toAccount.BalanceCents + amountCents

	if err := s.accountRepo.UpdateBalanceString(ctx, tx, fromAccountID, domain.CentsToDecimalString(newFrom)); err != nil {
		return fmt.Errorf("update from balance: %w", err)
	}
	if err := s.accountRepo.UpdateBalanceString(ctx, tx, toAccountID, domain.CentsToDecimalString(newTo)); err != nil {
		return fmt.Errorf("update to balance: %w", err)
	}

	return nil
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID uuid.UUID) (*domain.TokenPair, error) {
	accessToken, err := s.tokenManager.GenerateAccessToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth.generate_token_pair: generate access token: %w", err)
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth.generate_token_pair: generate refresh token: %w", err)
	}

	tokenHash := s.hasher.SHA256Hex(refreshToken)

	refreshTokenRecord := &RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.tokenManager.GetRefreshTokenTTL()),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, fmt.Errorf("auth.generate_token_pair: store refresh token: %w", err)
	}

	return &domain.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	claims, err := s.tokenManager.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		s.logger.Warn("Invalid token", "error", err)
		return uuid.Nil, apperr.ErrInvalidToken
	}

	return claims.UserID, nil
}

// GetUserByID returns a client-facing user DTO.
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			s.logger.Warn("User not found", "user_id", userID)
			return nil, apperr.ErrUserNotFound
		}
		return nil, fmt.Errorf("auth.get_user_by_id: %w", err)
	}
	return &domain.UserInfo{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// RefreshToken rotates refresh token and issues a new token pair.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	s.logger.Info("Refreshing token")

	tokenHash := s.hasher.SHA256Hex(refreshToken)

	tokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidToken) {
			s.logger.Warn("Refresh token not found or expired", "error", err)
			return nil, apperr.ErrInvalidToken
		}
		return nil, fmt.Errorf("auth.refresh_token: get refresh token: %w", err)
	}

	if err := s.refreshTokenRepo.Delete(ctx, tokenHash); err != nil {
		s.logger.Warn("Failed to delete old refresh token", "error", err)
	}

	tokenPair, err := s.generateTokenPair(ctx, tokenRecord.UserID)
	if err != nil {
		s.logger.Error("Failed to generate new tokens", "error", err, "user_id", tokenRecord.UserID)
		return nil, fmt.Errorf("auth.refresh_token: generate tokens: %w", err)
	}

	s.logger.Info("Token refreshed successfully", "user_id", tokenRecord.UserID)

	return tokenPair, nil
}

// Logout revokes the provided refresh token.
func (s *AuthService) Logout(ctx context.Context, in *domain.LogoutInput) error {
	s.logger.Info("User logout")

	claims, err := s.tokenManager.ValidateAccessToken(ctx, in.AccessToken)
	if err != nil {
		s.logger.Warn("Invalid access token during logout", "error", err)
	} else {
		s.logger.Info("Logging out user", "user_id", claims.UserID)
	}

	tokenHash := s.hasher.SHA256Hex(in.RefreshToken)
	if err := s.refreshTokenRepo.Delete(ctx, tokenHash); err != nil {
		s.logger.Warn("Failed to delete refresh token", "error", err)
	}

	s.logger.Info("User logged out successfully")
	return nil
}
