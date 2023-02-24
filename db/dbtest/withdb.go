package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vhakulinen/dino/db/utils"
)

// Creates a new database `dbname`, and drops it on after the test is done.
// Holds connection open based on `params` until test cleanup.
//
// Returns new ConnParmas pointed to the new database.
func WithCreateDB(t *testing.T, ctx context.Context, params *utils.ConnectionParams, dbname string) *utils.ConnectionParams {
	t.Helper()

	mainConn := OpenDB(t, ctx, params)

	if _, err := mainConn.Exec(ctx, `CREATE DATABASE `+dbname); err != nil {
		t.Fatalf("WithCreateDB: %v", err)
	}

	t.Cleanup(func() {
		if _, err := mainConn.Exec(ctx, `DROP DATABASE `+dbname); err != nil {
			t.Errorf("WithCreateDB: failed to drop database: %v", err)
		}
	})

	connParams := new(utils.ConnectionParams)
	*connParams = *params
	connParams.Database = dbname

	return connParams
}

// Opens a database pool for the duration of the test.
func OpenDB(t *testing.T, ctx context.Context, connParams *utils.ConnectionParams) *pgxpool.Pool {
	db, err := pgxpool.New(ctx, connParams.ConnString())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	return db
}
