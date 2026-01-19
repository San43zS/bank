package storage

import (
	"database/sql"
	"fmt"

	"banking-platform/internal/model"
	"github.com/google/uuid"
)

type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(account *model.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.GetDB().Exec(
		query,
		account.ID, account.UserID, account.Currency, account.Balance,
		account.CreatedAt, account.UpdatedAt,
	)
	return err
}

func (r *AccountRepository) GetByUserID(userID uuid.UUID) ([]*model.Account, error) {
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE user_id = $1 ORDER BY currency
	`

	rows, err := r.db.GetDB().Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		account := &model.Account{}
		if err := rows.Scan(
			&account.ID, &account.UserID, &account.Currency, &account.Balance,
			&account.CreatedAt, &account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

func (r *AccountRepository) GetByID(id uuid.UUID) (*model.Account, error) {
	account := &model.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE id = $1
	`
	err := r.db.GetDB().QueryRow(query, id).Scan(
		&account.ID, &account.UserID, &account.Currency, &account.Balance,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) GetByUserIDAndCurrency(userID uuid.UUID, currency model.Currency) (*model.Account, error) {
	account := &model.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE user_id = $1 AND currency = $2
	`
	err := r.db.GetDB().QueryRow(query, userID, currency).Scan(
		&account.ID, &account.UserID, &account.Currency, &account.Balance,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) FindAccountIDTx(tx *sql.Tx, userID uuid.UUID, currency model.Currency) (uuid.UUID, error) {
	var id uuid.UUID
	query := `SELECT id FROM accounts WHERE user_id = $1 AND currency = $2`
	err := tx.QueryRow(query, userID, currency).Scan(&id)
	if err == sql.ErrNoRows {
		return uuid.Nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *AccountRepository) UpdateBalanceString(tx *sql.Tx, accountID uuid.UUID, newBalance string) error {
	query := `UPDATE accounts SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.Exec(query, newBalance, accountID)
	return err
}

func (r *AccountRepository) LockAccountForUpdate(tx *sql.Tx, accountID uuid.UUID) (*model.Account, error) {
	account := &model.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE id = $1 FOR UPDATE
	`
	err := tx.QueryRow(query, accountID).Scan(
		&account.ID, &account.UserID, &account.Currency, &account.Balance,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) LockAccount(tx *sql.Tx, userID uuid.UUID, currency model.Currency) (*model.Account, error) {
	account := &model.Account{}
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts WHERE user_id = $1 AND currency = $2 FOR UPDATE
	`
	err := tx.QueryRow(query, userID, currency).Scan(
		&account.ID, &account.UserID, &account.Currency, &account.Balance,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}
