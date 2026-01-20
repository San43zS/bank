package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserInfo is a safe user representation without password hash.
type UserInfo struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TokenPair holds access/refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// AuthResult is returned by register/login.
type AuthResult struct {
	Tokens TokenPair
	User   UserInfo
}

// TransactionInfo is a transaction representation used for API responses.
type TransactionInfo struct {
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
	FromUserEmail        *string
	ToUserEmail          *string
}

