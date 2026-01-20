package storage

import (
	"context"
	"fmt"

	"banking-platform/internal/model"
	"github.com/google/uuid"
)

type LedgerRepository struct {
	db *DB
}

func NewLedgerRepository(db *DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) CreateEntry(ctx context.Context, tx Tx, entry *model.LedgerEntry) error {
	query := `
		INSERT INTO ledger (id, transaction_id, account_id, amount, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(
		ctx,
		query,
		entry.ID, entry.TransactionID, entry.AccountID, entry.Amount, entry.CreatedAt,
	)
	return err
}

func (r *LedgerRepository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*model.LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, account_id, amount, created_at
		FROM ledger WHERE transaction_id = $1 ORDER BY created_at
	`

	rows, err := r.db.GetDB().QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.LedgerEntry
	for rows.Next() {
		entry := &model.LedgerEntry{}
		if err := rows.Scan(
			&entry.ID, &entry.TransactionID, &entry.AccountID, &entry.Amount, &entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *LedgerRepository) VerifyTransactionBalanceTx(ctx context.Context, tx Tx, transactionID uuid.UUID) error {
	query := `SELECT COALESCE(SUM(amount), 0) FROM ledger WHERE transaction_id = $1`
	var sum float64
	if err := tx.QueryRowContext(ctx, query, transactionID).Scan(&sum); err != nil {
		return err
	}

	if sum > 0.000001 || sum < -0.000001 {
		return fmt.Errorf("ledger not balanced for transaction %s: sum=%0.6f", transactionID.String(), sum)
	}
	return nil
}

func (r *LedgerRepository) FindUnbalancedTransactionIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT transaction_id
		FROM ledger
		GROUP BY transaction_id
		HAVING COALESCE(SUM(amount), 0) <> 0
		ORDER BY transaction_id
		LIMIT $1
	`
	rows, err := r.db.GetDB().QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *LedgerRepository) FindAccountBalanceMismatches(ctx context.Context, limit int) ([]*model.AccountBalanceMismatch, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT
			a.id,
			a.user_id,
			a.currency,
			a.balance,
			COALESCE(SUM(l.amount), 0) AS ledger_sum,
			(a.balance - COALESCE(SUM(l.amount), 0)) AS diff
		FROM accounts a
		LEFT JOIN ledger l ON l.account_id = a.id
		GROUP BY a.id, a.user_id, a.currency, a.balance
		HAVING (a.balance - COALESCE(SUM(l.amount), 0)) <> 0
		ORDER BY ABS(a.balance - COALESCE(SUM(l.amount), 0)) DESC
		LIMIT $1
	`
	rows, err := r.db.GetDB().QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*model.AccountBalanceMismatch
	for rows.Next() {
		m := &model.AccountBalanceMismatch{}
		if err := rows.Scan(&m.AccountID, &m.UserID, &m.Currency, &m.Balance, &m.LedgerSum, &m.Diff); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
