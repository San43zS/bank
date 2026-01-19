package apperr

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user with this email already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidAmount      = errors.New("amount must have at most 2 decimal places")
	ErrAccountNotFound    = errors.New("account not found")
)
