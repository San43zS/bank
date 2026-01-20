package dto

import (
	"banking-platform/internal/domain"
	"github.com/google/uuid"
)

type AccountResponse struct {
	ID           uuid.UUID       `json:"id"`
	Currency     domain.Currency `json:"currency"`
	BalanceCents int64           `json:"balance_cents"`
}

type BalanceResponse struct {
	BalanceCents int64 `json:"balance_cents"`
}

