package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Account struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Currency     Currency
	BalanceCents int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Transaction struct {
	ID                   uuid.UUID
	Type                 TransactionType
	FromAccountID        *uuid.UUID
	ToAccountID          uuid.UUID
	AmountCents          int64
	Currency             Currency
	ExchangeRate         *float64
	ConvertedAmountCents *int64
	Description          string
	CreatedAt            time.Time
}

type LedgerEntry struct {
	ID            uuid.UUID
	TransactionID uuid.UUID
	AccountID     uuid.UUID
	AmountCents   int64
	CreatedAt     time.Time
}

type AccountBalanceMismatch struct {
	AccountID      uuid.UUID
	UserID         uuid.UUID
	Currency       Currency
	BalanceCents   int64
	LedgerSumCents int64
	DiffCents      int64
}
