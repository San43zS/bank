package domain

import "github.com/google/uuid"

// RegisterInput is the input for user registration.
type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// LoginInput is the input for user login.
type LoginInput struct {
	Email    string
	Password string
}

// LogoutInput is the input for logout.
type LogoutInput struct {
	AccessToken  string
	RefreshToken string
}

// TransferInput is the input for a transfer.
type TransferInput struct {
	ToUserID    *uuid.UUID
	ToUserEmail *string
	Currency    Currency
	AmountCents int64
}

// ExchangeInput is the input for currency exchange.
type ExchangeInput struct {
	FromCurrency Currency
	ToCurrency   Currency
	AmountCents  int64
}
