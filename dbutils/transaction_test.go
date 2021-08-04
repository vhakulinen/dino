package dbutils_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/vhakulinen/dino/dbtestutils"
	"github.com/vhakulinen/dino/dbutils"
)

func TestWithTransaction_commit(t *testing.T) {
	connParams := defaultConnectionParams
	dbname := strings.ToLower(t.Name())

	db, drop := dbtestutils.WithCreateDB(t, &connParams, dbname)
	defer drop(t)

	if _, err := db.Exec(`CREATE TABLE foobar (value TEXT NOT NULL);`); err != nil {
		t.Fatal(err)
	}

	err := dbutils.WithTransaction(db, context.Background(), func(tx *sqlx.Tx) error {
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

func TestWithTransaction_rollback(t *testing.T) {
	myerr := errors.New("My error")

	connParams := defaultConnectionParams
	dbname := strings.ToLower(t.Name())

	db, drop := dbtestutils.WithCreateDB(t, &connParams, dbname)
	defer drop(t)

	if _, err := db.Exec(`CREATE TABLE foobar (value TEXT NOT NULL);`); err != nil {
		t.Fatal(err)
	}

	err := dbutils.WithTransaction(db, context.Background(), func(tx *sqlx.Tx) error {
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
