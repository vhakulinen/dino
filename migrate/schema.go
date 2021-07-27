package migrate

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const SchemaTable = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL DEFAULT 0
);
`

// Initializes database for tracking migrations.
func InitDB(tx *sqlx.Tx) error {
	_, err := tx.Exec(SchemaTable)
	if err != nil {
		return err
	}

	var count int
	err = tx.QueryRowx(`SELECT COUNT(*) FROM schema_version`).Scan(&count)
	if err != nil {
		return err
	}

	// If no rows, insert the default value.
	if count == 0 {
		_, err = tx.Exec(`INSERT INTO schema_version VALUES (0)`)
		return err
	}

	if count != 1 {
		return fmt.Errorf("Expected 1 row in schema_version, got %d", count)
	}

	return nil
}

// Returns the current schema version in the database.
func QuerySchemaVersion(tx *sqlx.Tx) (int, error) {
	var version int
	err := tx.QueryRowx(`SELECT version FROM schema_version LIMIT 1`).Scan(&version)
	return version, err
}

func setSchemaVersion(tx *sqlx.Tx, v int) error {
	_, err := tx.Exec(`UPDATE schema_version SET version = $1`, v)
	return err
}
