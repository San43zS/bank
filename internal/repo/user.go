package repo

import (
	"context"
	"database/sql"
	"strings"

	"banking-platform/internal/apperr"
	"banking-platform/internal/domain"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a user. Email is normalized to lower-case before storing.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	query := `
		INSERT INTO users (id, email, password, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// GetByEmail returns a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	email = strings.ToLower(strings.TrimSpace(email))
	query := `SELECT id, email, password, first_name, last_name, created_at, updated_at
			  FROM users WHERE lower(email) = lower($1)`

	err := r.db.GetDB().QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
		&user.LastName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByID returns a user by UUID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password, first_name, last_name, created_at, updated_at
			  FROM users WHERE id = $1`

	err := r.db.GetDB().QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
		&user.LastName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, email, password, first_name, last_name, created_at, updated_at
			  FROM users ORDER BY created_at`

	rows, err := r.db.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
			&user.LastName, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
