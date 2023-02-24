package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5"
)

//go:embed schema.sql
var schema string

// Initializes database for tracking migrations.
func EnsureSchema(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, schema)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `SELECT COUNT(*) FROM schema_version`)
	if err != nil {
		return err
	}

	count, err := pgx.CollectOneRow(rows, pgx.RowTo[int])
	if err != nil {
		return err
	}

	// If no rows, insert the default value.
	if count == 0 {
		_, err = tx.Exec(ctx, `INSERT INTO schema_version VALUES (0)`)
		return err
	}

	if count != 1 {
		return fmt.Errorf("Expected 1 row in schema_version, got %d", count)
	}

	return nil
}

// Returns the current schema version in the database.
func QuerySchemaVersion(ctx context.Context, tx pgx.Tx) (int, error) {
	rows, err := tx.Query(ctx, `SELECT version FROM schema_version LIMIT 1`)
	if err != nil {
		return 0, err
	}

	return pgx.CollectOneRow(rows, pgx.RowTo[int])
}

func setSchemaVersion(ctx context.Context, tx pgx.Tx, v int) error {
	_, err := tx.Exec(ctx, `UPDATE schema_version SET version = $1`, v)
	return err
}
