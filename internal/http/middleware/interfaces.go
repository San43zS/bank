package middleware

import (
	"context"

	"github.com/google/uuid"
)

type AuthService interface {
	ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error)
}

