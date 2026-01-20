package repo

import (
	"context"
	"database/sql"
	"fmt"

	"banking-platform/internal/apperr"
	"banking-platform/internal/domain"
	"banking-platform/internal/service"
	"github.com/google/uuid"
)

type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create inserts a new account. Balance is stored as DECIMAL(15,2) in DB.
func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.GetDB().ExecContext(
		ctx,
		query,
		account.ID, account.UserID, account.Currency, domain.CentsToDecimalString(account.BalanceCents),
		account.CreatedAt, account.UpdatedAt,
	)
	return err
}

// GetByUserID loads all accounts for a user.
func (r *AccountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Account, error) {
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE user_id = $1 ORDER BY currency
	`

	rows, err := r.db.GetDB().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		account := &domain.Account{}
		var balanceStr string
		if err := rows.Scan(
			&account.ID, &account.UserID, &account.Currency, &balanceStr,
			&account.CreatedAt, &account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		bc, err := domain.DecimalStringToCents(balanceStr)
		if err != nil {
			return nil, fmt.Errorf("invalid balance in db for account %s: %w", account.ID.String(), err)
		}
		account.BalanceCents = bc
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

// GetByID loads a single account by id.
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	account := &domain.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE id = $1
	`
	var balanceStr string
	err := r.db.GetDB().QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.UserID, &account.Currency, &balanceStr,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	bc, err := domain.DecimalStringToCents(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("invalid balance in db for account %s: %w", account.ID.String(), err)
	}
	account.BalanceCents = bc
	return account, nil
}

// GetByUserIDAndCurrency loads a single account for a given user and currency.
func (r *AccountRepository) GetByUserIDAndCurrency(ctx context.Context, userID uuid.UUID, currency domain.Currency) (*domain.Account, error) {
	account := &domain.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE user_id = $1 AND currency = $2
	`
	var balanceStr string
	err := r.db.GetDB().QueryRowContext(ctx, query, userID, currency).Scan(
		&account.ID, &account.UserID, &account.Currency, &balanceStr,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	bc, err := domain.DecimalStringToCents(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("invalid balance in db for account %s: %w", account.ID.String(), err)
	}
	account.BalanceCents = bc
	return account, nil
}

// FindAccountIDTx finds account id within an existing transaction (used by services).
func (r *AccountRepository) FindAccountIDTx(ctx context.Context, tx service.Tx, userID uuid.UUID, currency domain.Currency) (uuid.UUID, error) {
	var id uuid.UUID
	query := `SELECT id FROM accounts WHERE user_id = $1 AND currency = $2`
	err := tx.QueryRowContext(ctx, query, userID, currency).Scan(&id)
	if err == sql.ErrNoRows {
		return uuid.Nil, apperr.ErrAccountNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// UpdateBalanceString updates balance using a decimal string (e.g. "10.25").
func (r *AccountRepository) UpdateBalanceString(ctx context.Context, tx service.Tx, accountID uuid.UUID, newBalance string) error {
	query := `UPDATE accounts SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, newBalance, accountID)
	return err
}

// LockAccountForUpdate locks the account row FOR UPDATE and returns the current state.
func (r *AccountRepository) LockAccountForUpdate(ctx context.Context, tx service.Tx, accountID uuid.UUID) (*domain.Account, error) {
	account := &domain.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE id = $1 FOR UPDATE
	`
	var balanceStr string
	err := tx.QueryRowContext(ctx, query, accountID).Scan(
		&account.ID, &account.UserID, &account.Currency, &balanceStr,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	bc, err := domain.DecimalStringToCents(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("invalid balance in db for account %s: %w", account.ID.String(), err)
	}
	account.BalanceCents = bc
	return account, nil
}

func (r *AccountRepository) LockAccount(ctx context.Context, tx service.Tx, userID uuid.UUID, currency domain.Currency) (*domain.Account, error) {
	account := &domain.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE user_id = $1 AND currency = $2 FOR UPDATE
	`
	var balanceStr string
	err := tx.QueryRowContext(ctx, query, userID, currency).Scan(
		&account.ID, &account.UserID, &account.Currency, &balanceStr,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	bc, err := domain.DecimalStringToCents(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("invalid balance in db for account %s: %w", account.ID.String(), err)
	}
	account.BalanceCents = bc
	return account, nil
}
