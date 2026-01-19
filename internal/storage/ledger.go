package storage

import (
	"database/sql"

	"banking-platform/internal/model"
	"github.com/google/uuid"
)

type LedgerRepository struct {
	db *DB
}

func NewLedgerRepository(db *DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) CreateEntry(tx *sql.Tx, entry *model.LedgerEntry) error {
	query := `
		INSERT INTO ledger (id, transaction_id, account_id, amount, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.Exec(
		query,
		entry.ID, entry.TransactionID, entry.AccountID, entry.Amount, entry.CreatedAt,
	)
	return err
}

func (r *LedgerRepository) GetByTransactionID(transactionID uuid.UUID) ([]*model.LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, account_id, amount, created_at
		FROM ledger WHERE transaction_id = $1 ORDER BY created_at
	`

	rows, err := r.db.GetDB().Query(query, transactionID)
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

func (r *LedgerRepository) VerifyTransactionBalance(transactionID uuid.UUID) (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM ledger WHERE transaction_id = $1`
	var sum float64
	err := r.db.GetDB().QueryRow(query, transactionID).Scan(&sum)
	return sum, err
}
