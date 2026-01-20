package model

import "github.com/google/uuid"

type AccountBalanceMismatch struct {
	AccountID uuid.UUID `json:"account_id"`
	UserID    uuid.UUID `json:"user_id"`
	Currency  Currency  `json:"currency"`
	Balance   float64   `json:"balance"`
	LedgerSum float64   `json:"ledger_sum"`
	Diff      float64   `json:"diff"`
}
