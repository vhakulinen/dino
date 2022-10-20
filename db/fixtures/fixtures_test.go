package fixtures_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "github.com/jackc/pgx/v5"

	"github.com/vhakulinen/dino/db/dbtest"
	"github.com/vhakulinen/dino/db/fixtures"
	"github.com/vhakulinen/dino/db/utils"
)

func TestDumpFixture(t *testing.T) {
	connParams := dbtest.DefaultConnectionParams
	dbname := strings.ToLower(t.Name())

	db, drop := dbtest.WithCreateDB(t, &connParams, dbname)
	defer drop(t)

	_, err := db.Exec(`
	CREATE TABLE foo (
		id SERIAL PRIMARY KEY,
		name TEXT
	);

	CREATE TABLE bar (
		id SERIAL PRIMARY KEY,
		num INTEGER NOT NULL
	);

	INSERT INTO foo (name) VALUES ('hey there'), ('well hello');
	INSERT INTO bar (num) VALUES (4), (9);
	`)
	if err != nil {
		t.Fatal(err)
	}

	dump, err := fixtures.DumpFixture(&utils.ConnectionParams{
		Host:     connParams.Host,
		Port:     connParams.Port,
		Username: connParams.Username,
		Password: connParams.Password,
		Database: dbname,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := `INSERT INTO public.bar (id, num) VALUES
	(1, 4),
	(2, 9);
INSERT INTO public.foo (id, name) VALUES
	(1, 'hey there'),
	(2, 'well hello');
`

	if diff := cmp.Diff(string(dump), expected); diff != "" {
		t.Error(diff)
	}
}

func TestLoadFixture(t *testing.T) {
	connParams := dbtest.DefaultConnectionParams
	dbname := strings.ToLower(t.Name())

	db, drop := dbtest.WithCreateDB(t, &connParams, dbname)
	defer drop(t)

	_, err := db.Exec(`
	CREATE TABLE foo (
		id SERIAL PRIMARY KEY,
		name TEXT
	);

	CREATE TABLE bar (
		id SERIAL PRIMARY KEY,
		num INTEGER NOT NULL
	);
	`)
	if err != nil {
		t.Fatal(err)
	}

	err = fixtures.LoadFixture(context.TODO(), db, `
	INSERT INTO bar (id, num) VALUES
		(1, 4),
		(2, 9),
		(3, 10);
	INSERT INTO foo (id, name) VALUES
		(1, 'hey there'),
		(2, 'well hello');
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Test that the sequences were reset.
	var i int
	if err := db.QueryRowx(`SELECT last_value FROM bar_id_seq`).Scan(&i); err != nil {
		t.Fatal(err)
	}

	if i != 3 {
		t.Errorf("Unexpected sequence value: %d", i)
	}

	if err := db.QueryRowx(`SELECT last_value FROM foo_id_seq`).Scan(&i); err != nil {
		t.Fatal(err)
	}
	if i != 2 {
		t.Errorf("Unexpected sequence value: %d", i)
	}

}