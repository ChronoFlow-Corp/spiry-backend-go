package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type txCtxKey struct{}

func injectTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func txFromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txCtxKey{}).(*sql.Tx)
	return tx
}

func (p Postgres) WithTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// finalize transaction on panic, etc.
	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, txErr)
				return
			}
		}

		txErr := tx.Commit()
		if txErr != nil {
			err = fmt.Errorf("commit err: %w", txErr)
		}
	}()

	// run callback
	err = tFunc(injectTx(ctx, tx))

	return nil
}