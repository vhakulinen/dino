package dbtestutils

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

// Creates a new database `dbname`, and drops it afterwards. Returned function
// will also close the returned database conn.
func WithCreateDB(t *testing.T, params *ConnectionParams, dbname string) (*sqlx.DB, func(t *testing.T)) {
	t.Helper()
	mainConn, err := sqlx.Open("postgres", params.ConnString())
	if err != nil {
		t.Fatalf("WithCreateDB: %v", err)
	}

	if _, err := mainConn.Exec(`CREATE DATABASE ` + dbname); err != nil {
		mainConn.Close()
		t.Fatalf("WithCreateDB: %v", err)
	}

	connParams := new(ConnectionParams)
	*connParams = *params
	connParams.Database = dbname

	db, err := sqlx.Open("postgres", connParams.ConnString())
	if err != nil {
		mainConn.Close()
		t.Fatalf("WithCreateDB: %v", err)
	}

	return db, func(t *testing.T) {
		if err := db.Close(); err != nil {
			t.Errorf("WithCreateDB: failed to close conn: %v", err)
		}

		if _, err := mainConn.Exec(`DROP DATABASE ` + dbname); err != nil {
			t.Errorf("WithCreateDB: failed to drop database: %v", err)
		}

		if err := mainConn.Close(); err != nil {
			t.Errorf("WithCreateDB: %v", err)
		}
	}
}
