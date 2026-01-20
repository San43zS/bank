package service

import (
	"context"
	"database/sql"
	"time"

	"banking-platform/internal/domain"
	"github.com/google/uuid"
)

type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Commit() error
	Rollback() error
}

type TxRunner interface {
	WithTx(ctx context.Context, fn func(tx Tx) error) error
}

type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetAll(ctx context.Context) ([]*domain.User, error)
}

type AccountRepo interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetByUserIDAndCurrency(ctx context.Context, userID uuid.UUID, currency domain.Currency) (*domain.Account, error)

	FindAccountIDTx(ctx context.Context, tx Tx, userID uuid.UUID, currency domain.Currency) (uuid.UUID, error)
	UpdateBalanceString(ctx context.Context, tx Tx, accountID uuid.UUID, newBalance string) error
	LockAccountForUpdate(ctx context.Context, tx Tx, accountID uuid.UUID) (*domain.Account, error)
}

type TransactionRepo interface {
	Create(ctx context.Context, tx Tx, transaction *domain.Transaction) error
	GetByUserID(ctx context.Context, userID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.TransactionWithEmails, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
}

type LedgerRepo interface {
	CreateEntry(ctx context.Context, tx Tx, entry *domain.LedgerEntry) error
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*domain.LedgerEntry, error)
	VerifyTransactionBalanceTx(ctx context.Context, tx Tx, transactionID uuid.UUID) error

	FindUnbalancedTransactionIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
	FindAccountBalanceMismatches(ctx context.Context, limit int) ([]*domain.AccountBalanceMismatch, error)
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Delete(ctx context.Context, tokenHash string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

