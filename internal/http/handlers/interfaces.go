package handler

import (
	"context"

	"banking-platform/internal/http/dto"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

type AccountService interface {
	GetUserAccounts(ctx context.Context, userID uuid.UUID) ([]*dto.AccountResponse, error)
	GetAccountBalance(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (int64, error)
}

type TransactionService interface {
	Transfer(ctx context.Context, fromUserID uuid.UUID, req *dto.TransferRequest) (*dto.TransactionResponse, error)
	Exchange(ctx context.Context, userID uuid.UUID, req *dto.ExchangeRequest) (*dto.TransactionResponse, error)
	GetUserTransactions(ctx context.Context, userID uuid.UUID, filter *dto.TransactionFilter) ([]*dto.TransactionResponse, error)
}

