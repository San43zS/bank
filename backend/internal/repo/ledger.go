package repo

import (
	"context"
	"fmt"

	"banking-platform/internal/domain"
	"banking-platform/internal/service"
	"github.com/google/uuid"
)

type LedgerRepository struct {
	db *DB
}

func NewLedgerRepository(db *DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

// CreateEntry inserts a ledger entry. Amount is stored in DB as DECIMAL(15,2).
func (r *LedgerRepository) CreateEntry(ctx context.Context, tx service.Tx, entry *domain.LedgerEntry) error {
	query := `
		INSERT INTO ledger (id, transaction_id, account_id, amount, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(
		ctx,
		query,
		entry.ID, entry.TransactionID, entry.AccountID, domain.CentsToDecimalString(entry.AmountCents), entry.CreatedAt,
	)
	return err
}

// GetByTransactionID loads all ledger entries for a transaction (ordered by creation time).
func (r *LedgerRepository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*domain.LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, account_id, amount, created_at
		FROM ledger WHERE transaction_id = $1 ORDER BY created_at
	`

	rows, err := r.db.GetDB().QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.LedgerEntry
	for rows.Next() {
		entry := &domain.LedgerEntry{}
		var amountStr string
		if err := rows.Scan(
			&entry.ID, &entry.TransactionID, &entry.AccountID, &amountStr, &entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		ac, err := domain.DecimalStringToCents(amountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ledger amount in db for entry %s: %w", entry.ID.String(), err)
		}
		entry.AmountCents = ac
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// VerifyTransactionBalanceTx ensures the ledger is balanced for a given transaction.
func (r *LedgerRepository) VerifyTransactionBalanceTx(ctx context.Context, tx service.Tx, transactionID uuid.UUID) error {
	query := `SELECT COALESCE(SUM((amount * 100)::bigint), 0) FROM ledger WHERE transaction_id = $1`
	var sumCents int64
	if err := tx.QueryRowContext(ctx, query, transactionID).Scan(&sumCents); err != nil {
		return err
	}

	if sumCents != 0 {
		return fmt.Errorf("ledger not balanced for transaction %s: sum_cents=%d", transactionID.String(), sumCents)
	}
	return nil
}

// FindUnbalancedTransactionIDs finds transactions for which the ledger sum (in cents) is not zero.
func (r *LedgerRepository) FindUnbalancedTransactionIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT transaction_id
		FROM ledger
		GROUP BY transaction_id
		HAVING COALESCE(SUM((amount * 100)::bigint), 0) <> 0
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

// FindAccountBalanceMismatches compares account balance vs ledger sum (both in cents).
func (r *LedgerRepository) FindAccountBalanceMismatches(ctx context.Context, limit int) ([]*domain.AccountBalanceMismatch, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT
			a.id,
			a.user_id,
			a.currency,
			(a.balance * 100)::bigint AS balance_cents,
			COALESCE(SUM((l.amount * 100)::bigint), 0) AS ledger_sum_cents,
			((a.balance * 100)::bigint - COALESCE(SUM((l.amount * 100)::bigint), 0)) AS diff_cents
		FROM accounts a
		LEFT JOIN ledger l ON l.account_id = a.id
		GROUP BY a.id, a.user_id, a.currency, a.balance
		HAVING ((a.balance * 100)::bigint - COALESCE(SUM((l.amount * 100)::bigint), 0)) <> 0
		ORDER BY ABS(((a.balance * 100)::bigint - COALESCE(SUM((l.amount * 100)::bigint), 0))) DESC
		LIMIT $1
	`
	rows, err := r.db.GetDB().QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.AccountBalanceMismatch
	for rows.Next() {
		m := &domain.AccountBalanceMismatch{}
		if err := rows.Scan(&m.AccountID, &m.UserID, &m.Currency, &m.BalanceCents, &m.LedgerSumCents, &m.DiffCents); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
