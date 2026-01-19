package model

import (
	"time"

	"github.com/google/uuid"
)

type LedgerEntry struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TransactionID uuid.UUID `json:"transaction_id" db:"transaction_id"`
	AccountID     uuid.UUID `json:"account_id" db:"account_id"`
	Amount        float64   `json:"amount" db:"amount"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type LedgerEntryWithAccount struct {
	LedgerEntry
	AccountCurrency Currency  `json:"account_currency" db:"account_currency"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
}
