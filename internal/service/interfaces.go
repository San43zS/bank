package service

import (
	"context"

	"banking-platform/internal/model"
	"github.com/google/uuid"
)

type IAuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

type IAccountService interface {
	GetUserAccounts(ctx context.Context, userID uuid.UUID) ([]*model.AccountResponse, error)
	GetAccountBalance(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (float64, error)
}

type ITransactionService interface {
	Transfer(ctx context.Context, fromUserID uuid.UUID, req *model.TransferRequest) (*model.TransactionResponse, error)
	Exchange(ctx context.Context, userID uuid.UUID, req *model.ExchangeRequest) (*model.TransactionResponse, error)
	GetUserTransactions(ctx context.Context, userID uuid.UUID, filter *model.TransactionFilter) ([]*model.TransactionResponse, error)
}
