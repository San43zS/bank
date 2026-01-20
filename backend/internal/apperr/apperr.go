package apperr

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUserExists           = errors.New("user with this email already exists")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidAmount        = errors.New("amount must have at most 2 decimal places")
	ErrAccountNotFound      = errors.New("account not found")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrCurrenciesMustDiffer = errors.New("from and to currencies must be different")
	ErrCannotTransferToSelf = errors.New("cannot transfer to self")
	ErrLiquidityUnavailable = errors.New("exchange liquidity unavailable")
)

// PublicError is a client-facing error with an associated HTTP status code.
type PublicError struct {
	Status  int
	Message string
	Err     error
}

func (e *PublicError) Error() string { return e.Message }
func (e *PublicError) Unwrap() error { return e.Err }

func BadRequest(message string) error {
	return &PublicError{Status: 400, Message: message}
}

// RootCause unwraps err until it cannot be unwrapped any further.
func RootCause(err error) error {
	if err == nil {
		return nil
	}
	for {
		u := errors.Unwrap(err)
		if u == nil {
			return err
		}
		err = u
	}
}
