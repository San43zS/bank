package repo

import (
	"context"
	"fmt"

	"banking-platform/internal/service"
)

func (d *DB) WithTx(ctx context.Context, fn func(tx service.Tx) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
