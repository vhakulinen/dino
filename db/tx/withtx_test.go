package tx_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/vhakulinen/dino/db/dbtest"
	"github.com/vhakulinen/dino/db/tx"
)

func TestBeginFn_commit(t *testing.T) {
	connParams := dbtest.DefaultConnectionParams
	dbname := strings.ToLower(t.Name())

	db := dbtest.OpenDB(t, "pgx", dbtest.WithCreateDB(t, "pgx", &connParams, dbname))

	if _, err := db.Exec(`CREATE TABLE foobar (value TEXT NOT NULL);`); err != nil {
		t.Fatal(err)
	}

	err := tx.BeginFn(context.Background(), db, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO foobar VALUES ('foobar')`)
		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	var count int
	if err := db.QueryRowx(`SELECT COUNT(*) FROM foobar`).Scan(&count); err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Invalid count: %d", err)
	}
}

func TestBeginFn_rollback(t *testing.T) {
	myerr := errors.New("My error")

	connParams := dbtest.DefaultConnectionParams
	dbname := strings.ToLower(t.Name())

	db := dbtest.OpenDB(t, "pgx", dbtest.WithCreateDB(t, "pgx", &connParams, dbname))

	if _, err := db.Exec(`CREATE TABLE foobar (value TEXT NOT NULL);`); err != nil {
		t.Fatal(err)
	}

	err := tx.BeginFn(context.Background(), db, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO foobar VALUES ('foobar')`)
		if err != nil {
			t.Errorf("Exec failed: %v", err)
			return err
		}

		return myerr
	})

	if err != myerr {
		t.Fatalf("Expected myerr, got %v", err)
	}

	var count int
	if err := db.QueryRowx(`SELECT COUNT(*) FROM foobar`).Scan(&count); err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("Invalid count: %d", err)
	}
}
