package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"banking-platform/internal/apperr"
	"banking-platform/internal/hash"
	"banking-platform/internal/jwt"
	"banking-platform/internal/model"
	"banking-platform/internal/storage"
)

type AuthService struct {
	userRepo         *storage.UserRepository
	accountRepo      *storage.AccountRepository
	refreshTokenRepo *storage.RefreshTokenRepository
	tokenManager     *jwt.TokenManager
	hasher           *hash.Hasher
	logger           *slog.Logger
}

func NewAuthService(
	userRepo *storage.UserRepository,
	accountRepo *storage.AccountRepository,
	refreshTokenRepo *storage.RefreshTokenRepository,
	tokenManager *jwt.TokenManager,
	hasher *hash.Hasher,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		accountRepo:      accountRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenManager:     tokenManager,
		hasher:           hasher,
		logger:           logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	s.logger.Info("Registering new user", "email", req.Email)

	_, err := s.userRepo.GetByEmail(req.Email)
	if err == nil {
		s.logger.Warn("User already exists", "email", req.Email)
		return nil, apperr.ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := &model.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.createDefaultAccounts(user.ID); err != nil {
		s.logger.Error("Failed to create default accounts", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to create default accounts: %w", err)
	}

	tokenPair, err := s.generateTokenPair(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.logger.Info("User registered successfully", "user_id", user.ID, "email", req.Email)

	return &model.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	s.logger.Info("User login attempt", "email", req.Email)

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		s.logger.Warn("User not found", "email", req.Email)
		return nil, apperr.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("Invalid password", "email", req.Email)
		return nil, apperr.ErrInvalidCredentials
	}

	tokenPair, err := s.generateTokenPair(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.logger.Info("User logged in successfully", "user_id", user.ID, "email", req.Email)

	return &model.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) createDefaultAccounts(userID uuid.UUID) error {
	now := time.Now()

	usdAccount := &model.Account{
		ID:        uuid.New(),
		UserID:    userID,
		Currency:  model.CurrencyUSD,
		Balance:   1000.00,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.accountRepo.Create(usdAccount); err != nil {
		s.logger.Error("Failed to create USD account", "error", err, "user_id", userID)
		return err
	}

	eurAccount := &model.Account{
		ID:        uuid.New(),
		UserID:    userID,
		Currency:  model.CurrencyEUR,
		Balance:   500.00,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.accountRepo.Create(eurAccount); err != nil {
		s.logger.Error("Failed to create EUR account", "error", err, "user_id", userID)
		return err
	}

	s.logger.Info("Default accounts created", "user_id", userID)
	return nil
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID uuid.UUID) (*model.TokenResponse, error) {
	accessToken, err := s.tokenManager.GenerateAccessToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenHash := s.hasher.SHA256Hex(refreshToken)

	refreshTokenRecord := &storage.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.tokenManager.GetRefreshTokenTTL()),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	claims, err := s.tokenManager.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		s.logger.Warn("Invalid token", "error", err)
		return uuid.Nil, apperr.ErrInvalidToken
	}

	return claims.UserID, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Warn("User not found", "user_id", userID)
		return nil, apperr.ErrUserNotFound
	}
	return user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, error) {
	s.logger.Info("Refreshing token")

	tokenHash := s.hasher.SHA256Hex(refreshToken)

	tokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		s.logger.Warn("Refresh token not found or expired", "error", err)
		return nil, apperr.ErrInvalidToken
	}

	if err := s.refreshTokenRepo.Delete(ctx, tokenHash); err != nil {
		s.logger.Warn("Failed to delete old refresh token", "error", err)
	}

	tokenPair, err := s.generateTokenPair(ctx, tokenRecord.UserID)
	if err != nil {
		s.logger.Error("Failed to generate new tokens", "error", err, "user_id", tokenRecord.UserID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.logger.Info("Token refreshed successfully", "user_id", tokenRecord.UserID)

	return tokenPair, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	s.logger.Info("User logout")

	claims, err := s.tokenManager.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		s.logger.Warn("Invalid access token during logout", "error", err)
	} else {
		s.logger.Info("Logging out user", "user_id", claims.UserID)
	}

	tokenHash := s.hasher.SHA256Hex(refreshToken)
	if err := s.refreshTokenRepo.Delete(ctx, tokenHash); err != nil {
		s.logger.Warn("Failed to delete refresh token", "error", err)
	}

	s.logger.Info("User logged out successfully")
	return nil
}
