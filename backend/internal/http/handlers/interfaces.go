package handler

import (
	"context"

	"banking-platform/internal/domain"
	"github.com/google/uuid"
)

// AuthService defines auth operations used by HTTP handlers.
type AuthService interface {
	Register(ctx context.Context, in *domain.RegisterInput) (*domain.AuthResult, error)
	Login(ctx context.Context, in *domain.LoginInput) (*domain.AuthResult, error)
	ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.UserInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, in *domain.LogoutInput) error
}

// AccountService defines account operations used by HTTP handlers.
type AccountService interface {
	GetUserAccounts(ctx context.Context, userID uuid.UUID) ([]*domain.Account, error)
	GetAccountBalance(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (int64, error)
}

// TransactionService defines transaction operations used by HTTP handlers.
type TransactionService interface {
	Transfer(ctx context.Context, fromUserID uuid.UUID, in *domain.TransferInput) (*domain.TransactionInfo, error)
	Exchange(ctx context.Context, userID uuid.UUID, in *domain.ExchangeInput) (*domain.TransactionInfo, error)
	GetUserTransactions(ctx context.Context, userID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.TransactionInfo, error)
}
