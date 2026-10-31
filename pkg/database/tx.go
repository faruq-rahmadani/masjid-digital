package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

func txFromContext(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx
}

// WithTx runs fn in a transaction: committed when fn returns nil, rolled back
// when fn returns an error or panics. Repository calls made with the ctx passed
// to fn join the transaction. A nested WithTx joins the outer transaction.
func (d *DB) WithTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if txFromContext(ctx) != nil {
		return fn(ctx)
	}

	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", d.translate(err))
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, fmt.Errorf("error rolling back transaction: %w", rbErr))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", d.translate(err))
	}

	return nil
}
