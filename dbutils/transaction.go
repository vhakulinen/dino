package dbutils

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type TxFunc func(tx *sqlx.Tx) error

func WithTransaction(db *sqlx.DB, ctx context.Context, fn TxFunc) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			rberr := tx.Rollback()
			if rberr != nil {
				log.Printf("Rollback failed. Wrapping original error (orignal was: %v)", err)
				err = fmt.Errorf("Rollback failed: %w (original error was: %v)", rberr, err)
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}
