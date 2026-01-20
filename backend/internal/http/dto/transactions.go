package dto

import (
	"time"

	"banking-platform/internal/domain"
	"github.com/google/uuid"
)

type TransferRequest struct {
	ToUserID    *uuid.UUID      `json:"to_user_id,omitempty"`
	ToUserEmail *string         `json:"to_user_email,omitempty"`
	Currency    domain.Currency `json:"currency" binding:"required,oneof=USD EUR"`
	AmountCents int64           `json:"amount_cents" binding:"required,gt=0"`
}

type ExchangeRequest struct {
	FromCurrency domain.Currency `json:"from_currency" binding:"required,oneof=USD EUR"`
	ToCurrency   domain.Currency `json:"to_currency" binding:"required,oneof=USD EUR"`
	AmountCents  int64           `json:"amount_cents" binding:"required,gt=0"`
}

type TransactionResponse struct {
	ID                   uuid.UUID              `json:"id"`
	Type                 domain.TransactionType `json:"type"`
	FromAccountID        *uuid.UUID             `json:"from_account_id,omitempty"`
	ToAccountID          uuid.UUID              `json:"to_account_id"`
	AmountCents          int64                  `json:"amount_cents"`
	Currency             domain.Currency        `json:"currency"`
	ExchangeRate         *float64               `json:"exchange_rate,omitempty"`
	ConvertedAmountCents *int64                 `json:"converted_amount_cents,omitempty"`
	Description          string                 `json:"description"`
	CreatedAt            time.Time              `json:"created_at"`
	FromUserEmail        *string                `json:"from_user_email,omitempty"`
	ToUserEmail          *string                `json:"to_user_email,omitempty"`
}

type TransactionFilter struct {
	Type  domain.TransactionType `form:"type"`
	Page  int                    `form:"page"`
	Limit int                    `form:"limit"`
}
