package dbtest

import (
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/vhakulinen/dino/db/utils"
)

// Creates a new database `dbname`, and drops it on after the test is done.
// Holds connection open based on `params` until test cleanup.
//
// Returns new ConnParmas pointed to the new database.
func WithCreateDB(t *testing.T, driver string, params *utils.ConnectionParams, dbname string) *utils.ConnectionParams {
	t.Helper()

	mainConn, err := sqlx.Open(driver, params.ConnString())
	if err != nil {
		t.Fatalf("WithCreateDB: %v", err)
	}

	t.Cleanup(func() {
		mainConn.Close()
	})

	if _, err := mainConn.Exec(`CREATE DATABASE ` + dbname); err != nil {
		t.Fatalf("WithCreateDB: %v", err)
	}

	t.Cleanup(func() {
		if _, err := mainConn.Exec(`DROP DATABASE ` + dbname); err != nil {
			t.Errorf("WithCreateDB: failed to drop database: %v", err)
		}

		if err := mainConn.Close(); err != nil {
			t.Errorf("WithCreateDB: %v", err)
		}
	})

	connParams := new(utils.ConnectionParams)
	*connParams = *params
	connParams.Database = dbname

	return connParams
}

// Opens `sqlx.DB` for the duration of the test.
func OpenDB(t *testing.T, driver string, connParams *utils.ConnectionParams) *sqlx.DB {
	db, err := sqlx.Open(driver, connParams.ConnString())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	return db
}
