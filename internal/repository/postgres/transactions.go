package postgres

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/repository/ent"
)

// TxFunc represents a function that executes within an ent transaction
type TxFunc func(ctx context.Context, tx *ent.Tx) error

// withTx executes a function within an ent transaction
func (r *postgresRepository) withTx(ctx context.Context, fn TxFunc) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic and re-panic
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("rollback failed after panic: %v (original panic: %v)", rollbackErr, p)
			}
			panic(p)
		} else if err != nil {
			// Rollback on error
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("rollback failed: %v (original error: %w)", rollbackErr, err)
			}
		} else {
			// Commit on success
			err = tx.Commit()
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
			}
		}
	}()

	err = fn(ctx, tx)
	return err
}