package storage

import (
	"context"

	"banking-platform/internal/model"
	"github.com/google/uuid"
)

type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetAll(ctx context.Context) ([]*model.User, error)
}

type AccountRepo interface {
	Create(ctx context.Context, account *model.Account) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error)
	GetByUserIDAndCurrency(ctx context.Context, userID uuid.UUID, currency model.Currency) (*model.Account, error)

	FindAccountIDTx(ctx context.Context, tx Tx, userID uuid.UUID, currency model.Currency) (uuid.UUID, error)
	UpdateBalanceString(ctx context.Context, tx Tx, accountID uuid.UUID, newBalance string) error
	LockAccountForUpdate(ctx context.Context, tx Tx, accountID uuid.UUID) (*model.Account, error)
}

type TransactionRepo interface {
	Create(ctx context.Context, tx Tx, transaction *model.Transaction) error
	GetByUserID(ctx context.Context, userID uuid.UUID, filter *model.TransactionFilter) ([]*model.TransactionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
}

type LedgerRepo interface {
	CreateEntry(ctx context.Context, tx Tx, entry *model.LedgerEntry) error
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*model.LedgerEntry, error)
	VerifyTransactionBalanceTx(ctx context.Context, tx Tx, transactionID uuid.UUID) error

	FindUnbalancedTransactionIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
	FindAccountBalanceMismatches(ctx context.Context, limit int) ([]*model.AccountBalanceMismatch, error)
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Delete(ctx context.Context, tokenHash string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
