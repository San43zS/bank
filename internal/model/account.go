package model

import (
	"time"

	"github.com/google/uuid"
)

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

type Account struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Currency  Currency  `json:"currency" db:"currency"`
	Balance   float64   `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type AccountResponse struct {
	ID       uuid.UUID `json:"id"`
	Currency Currency  `json:"currency"`
	Balance  float64   `json:"balance"`
}
