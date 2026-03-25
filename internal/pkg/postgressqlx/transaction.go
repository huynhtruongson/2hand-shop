package postgressqlx

import (
	"context"
	"database/sql"
	"fmt"
)

type TxFn func(ctx context.Context, tx TX) error

func (db *sqlxDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (TX, error) {
	return db.BeginTxx(ctx, opts)

}

func ExecTx(ctx context.Context, db TxStarter, fn TxFn) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw so the caller's recovery logic still fires
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("db: rollback failed (%v) after error: %w", rbErr, err)
			}
		}
	}()

	if err = fn(ctx, tx); err != nil {
		return err // defer handles rollback
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("db: commit transaction: %w", err)
	}

	return nil
}
