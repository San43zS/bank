package storage

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"banking-platform/internal/model"
)

type TransactionRepository struct {
	db *DB
}

func NewTransactionRepository(db *DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *sql.Tx, transaction *model.Transaction) error {
	query := `
		INSERT INTO transactions (id, type, from_account_id, to_account_id, amount, currency, exchange_rate, converted_amount, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := tx.Exec(
		query,
		transaction.ID, transaction.Type, transaction.FromAccountID, transaction.ToAccountID,
		transaction.Amount, transaction.Currency, transaction.ExchangeRate,
		transaction.ConvertedAmount, transaction.Description, transaction.CreatedAt,
	)
	return err
}

func (r *TransactionRepository) GetByUserID(userID uuid.UUID, filter *model.TransactionFilter) ([]*model.TransactionResponse, error) {
	query := `
		SELECT 
			t.id, t.type, t.from_account_id, t.to_account_id, t.amount, t.currency,
			t.exchange_rate, t.converted_amount, t.description, t.created_at,
			from_user.email as from_user_email,
			to_user.email as to_user_email
		FROM transactions t
		LEFT JOIN accounts from_acc ON t.from_account_id = from_acc.id
		LEFT JOIN users from_user ON from_acc.user_id = from_user.id
		JOIN accounts to_acc ON t.to_account_id = to_acc.id
		JOIN users to_user ON to_acc.user_id = to_user.id
		WHERE (from_acc.user_id = $1 OR to_acc.user_id = $1)
	`
	
	args := []interface{}{userID}
	argIndex := 2

	if filter.Type != "" {
		query += fmt.Sprintf(" AND t.type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}

	query += " ORDER BY t.created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.Limit
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, offset)
		}
	}

	rows, err := r.db.GetDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.TransactionResponse
	for rows.Next() {
		t := &model.TransactionResponse{}
		var fromUserEmail, toUserEmail sql.NullString
		
		if err := rows.Scan(
			&t.ID, &t.Type, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Currency,
			&t.ExchangeRate, &t.ConvertedAmount, &t.Description, &t.CreatedAt,
			&fromUserEmail, &toUserEmail,
		); err != nil {
			return nil, err
		}
		
		if fromUserEmail.Valid {
			t.FromUserEmail = &fromUserEmail.String
		}
		if toUserEmail.Valid {
			t.ToUserEmail = &toUserEmail.String
		}
		
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

func (r *TransactionRepository) GetByID(id uuid.UUID) (*model.Transaction, error) {
	transaction := &model.Transaction{}
	query := `
		SELECT id, type, from_account_id, to_account_id, amount, currency, exchange_rate, converted_amount, description, created_at
		FROM transactions WHERE id = $1
	`
	
	var fromAccountID sql.NullString
	err := r.db.GetDB().QueryRow(query, id).Scan(
		&transaction.ID, &transaction.Type, &fromAccountID, &transaction.ToAccountID,
		&transaction.Amount, &transaction.Currency, &transaction.ExchangeRate,
		&transaction.ConvertedAmount, &transaction.Description, &transaction.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, err
	}
	
	if fromAccountID.Valid {
		parsedID, _ := uuid.Parse(fromAccountID.String)
		transaction.FromAccountID = &parsedID
	}
	
	return transaction, nil
}
