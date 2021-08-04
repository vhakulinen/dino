package dbtestutils

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vhakulinen/dino/dbutils"
)

func TestWithCreateDB(t *testing.T) {
	// TODO(ville): Read these variables from env or something.
	connParams := &dbutils.ConnectionParams{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",

		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
	}

	conn, err := sqlx.Open("postgres", connParams.ConnString())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	query := `SELECT datname FROM pg_database WHERE datname = $1`
	dbname := strings.ToLower(t.Name())

	defer func() {
		var got string
		err := conn.QueryRowx(query, dbname).Scan(&got)

		// There should be no rows, since the database was dropped already.
		if err != sql.ErrNoRows {
			t.Fatalf("Expected sql.ErrNoRows, got %v", err)
		}
	}()

	_, drop := WithCreateDB(t, connParams, dbname)
	defer drop(t)

	var got string
	if err := conn.QueryRowx(query, dbname).Scan(&got); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(got, dbname); diff != "" {
		t.Fatal(diff)
	}
}
