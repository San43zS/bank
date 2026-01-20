package repo

import (
	"context"
	"database/sql"

	"banking-platform/internal/apperr"
	"banking-platform/internal/service"
	"github.com/google/uuid"
)

type RefreshTokenRepository struct {
	db *DB
}

func NewRefreshTokenRepository(db *DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create inserts a refresh token record.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *service.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.GetDB().ExecContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	return err
}

// GetByTokenHash returns a non-expired refresh token record by hash.
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*service.RefreshToken, error) {
	token := &service.RefreshToken{}
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	err := r.db.GetDB().QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Delete removes a refresh token record by hash.
func (r *RefreshTokenRepository) Delete(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.db.GetDB().ExecContext(ctx, query, tokenHash)
	return err
}

// DeleteByUserID removes all refresh tokens for a given user.
func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.GetDB().ExecContext(ctx, query, userID)
	return err
}
