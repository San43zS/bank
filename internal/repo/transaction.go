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

type TransactionRepository struct {
	db *DB
}

func NewTransactionRepository(db *DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a transaction row. Amounts are stored as DECIMAL(15,2) in DB.
func (r *TransactionRepository) Create(ctx context.Context, tx service.Tx, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (id, type, from_account_id, to_account_id, amount, currency, exchange_rate, converted_amount, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	var converted any = nil
	if transaction.ConvertedAmountCents != nil {
		converted = domain.CentsToDecimalString(*transaction.ConvertedAmountCents)
	}
	_, err := tx.ExecContext(
		ctx,
		query,
		transaction.ID, transaction.Type, transaction.FromAccountID, transaction.ToAccountID,
		domain.CentsToDecimalString(transaction.AmountCents), transaction.Currency, transaction.ExchangeRate,
		converted, transaction.Description, transaction.CreatedAt,
	)
	return err
}

// GetByUserID returns a filtered/paginated list of transactions visible to a user.
func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.TransactionWithEmails, error) {
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

	rows, err := r.db.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.TransactionWithEmails
	for rows.Next() {
		out := &domain.TransactionWithEmails{}
		t := &out.Transaction
		var fromUserEmail, toUserEmail sql.NullString
		var amountStr string
		var convertedStr sql.NullString
		var exchangeRate sql.NullFloat64

		if err := rows.Scan(
			&t.ID, &t.Type, &t.FromAccountID, &t.ToAccountID, &amountStr, &t.Currency,
			&exchangeRate, &convertedStr, &t.Description, &t.CreatedAt,
			&fromUserEmail, &toUserEmail,
		); err != nil {
			return nil, err
		}

		ac, err := domain.DecimalStringToCents(amountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction amount in db for %s: %w", t.ID.String(), err)
		}
		t.AmountCents = ac

		if exchangeRate.Valid {
			v := exchangeRate.Float64
			t.ExchangeRate = &v
		}
		if convertedStr.Valid {
			cc, err := domain.DecimalStringToCents(convertedStr.String)
			if err != nil {
				return nil, fmt.Errorf("invalid converted_amount in db for %s: %w", t.ID.String(), err)
			}
			t.ConvertedAmountCents = &cc
		}

		if fromUserEmail.Valid {
			out.FromUserEmail = &fromUserEmail.String
		}
		if toUserEmail.Valid {
			out.ToUserEmail = &toUserEmail.String
		}

		transactions = append(transactions, out)
	}
	return transactions, rows.Err()
}

// GetByID loads a transaction by id and decodes DECIMAL amounts into cents.
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	transaction := &domain.Transaction{}
	query := `
		SELECT id, type, from_account_id, to_account_id, amount, currency, exchange_rate, converted_amount, description, created_at
		FROM transactions WHERE id = $1
	`

	var fromAccountID sql.NullString
	var amountStr string
	var convertedStr sql.NullString
	var exchangeRate sql.NullFloat64
	err := r.db.GetDB().QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.Type, &fromAccountID, &transaction.ToAccountID,
		&amountStr, &transaction.Currency, &exchangeRate,
		&convertedStr, &transaction.Description, &transaction.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperr.ErrTransactionNotFound
	}
	if err != nil {
		return nil, err
	}

	if fromAccountID.Valid {
		parsedID, _ := uuid.Parse(fromAccountID.String)
		transaction.FromAccountID = &parsedID
	}

	ac, err := domain.DecimalStringToCents(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction amount in db for %s: %w", transaction.ID.String(), err)
	}
	transaction.AmountCents = ac

	if exchangeRate.Valid {
		v := exchangeRate.Float64
		transaction.ExchangeRate = &v
	}
	if convertedStr.Valid {
		cc, err := domain.DecimalStringToCents(convertedStr.String)
		if err != nil {
			return nil, fmt.Errorf("invalid converted_amount in db for %s: %w", transaction.ID.String(), err)
		}
		transaction.ConvertedAmountCents = &cc
	}

	return transaction, nil
}
