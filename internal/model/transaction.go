package model

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypeExchange TransactionType = "exchange"
)

type Transaction struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Type            TransactionType `json:"type" db:"type"`
	FromAccountID   *uuid.UUID      `json:"from_account_id,omitempty" db:"from_account_id"`
	ToAccountID     uuid.UUID       `json:"to_account_id" db:"to_account_id"`
	Amount          float64         `json:"amount" db:"amount"`
	Currency        Currency        `json:"currency" db:"currency"`
	ExchangeRate    *float64        `json:"exchange_rate,omitempty" db:"exchange_rate"`
	ConvertedAmount *float64        `json:"converted_amount,omitempty" db:"converted_amount"`
	Description     string          `json:"description" db:"description"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type TransferRequest struct {
	ToUserID uuid.UUID `json:"to_user_id" binding:"required"`
	Currency Currency  `json:"currency" binding:"required,oneof=USD EUR"`
	Amount   float64   `json:"amount" binding:"required,gt=0"`
}

type ExchangeRequest struct {
	FromCurrency Currency `json:"from_currency" binding:"required,oneof=USD EUR"`
	ToCurrency   Currency `json:"to_currency" binding:"required,oneof=USD EUR"`
	Amount       float64  `json:"amount" binding:"required,gt=0"`
}

type TransactionResponse struct {
	ID              uuid.UUID       `json:"id"`
	Type            TransactionType `json:"type"`
	FromAccountID   *uuid.UUID      `json:"from_account_id,omitempty"`
	ToAccountID     uuid.UUID       `json:"to_account_id"`
	Amount          float64         `json:"amount"`
	Currency        Currency        `json:"currency"`
	ExchangeRate    *float64        `json:"exchange_rate,omitempty"`
	ConvertedAmount *float64        `json:"converted_amount,omitempty"`
	Description     string          `json:"description"`
	CreatedAt       time.Time       `json:"created_at"`

	FromUserEmail *string `json:"from_user_email,omitempty"`
	ToUserEmail   *string `json:"to_user_email,omitempty"`
}

type TransactionFilter struct {
	Type  TransactionType `form:"type"`
	Page  int             `form:"page"`
	Limit int             `form:"limit"`
}
