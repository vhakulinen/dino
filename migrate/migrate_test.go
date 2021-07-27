package migrate

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vhakulinen/dino/dbutils"
)

const testmigrationsPath = "testdata/testmigrations"

// Creates a empty database for testing. Returns a connection to said database
// and a close function to close the connection, and drop the database. Do not
// manually close the returned db.
func createTestDB(t *testing.T, dbname string) (*sqlx.DB, func(t *testing.T)) {
	t.Helper()

	connstr := "user='postgres' password='password' host='localhost' port='5432' sslmode='disable'"
	conn, err := sqlx.Open("postgres", connstr)
	if err != nil {
		t.Fatalf("openTestDB: %v", err)
	}
	if _, err := conn.Exec(`CREATE DATABASE ` + dbname); err != nil {
		t.Fatalf("openTestDB: %v", err)
	}

	connstr = fmt.Sprintf("user='postgres' password='password' dbname='%s' host='localhost' port='5432' sslmode='disable'", dbname)
	db, err := sqlx.Open("postgres", connstr)
	if err != nil {
		t.Fatalf("openTestDB: %v", err)
	}

	return db, func(t *testing.T) {
		if err := db.Close(); err != nil {
			t.Errorf("openTestDB: failed to close conn: %v", err)
		}

		if _, err := conn.Exec(`DROP DATABASE ` + dbname); err != nil {
			t.Errorf("openTestDB: failed to drop database: %v", err)
		}

		conn.Close()
	}
}

func TestMigrationsFromFS(t *testing.T) {
	source := os.DirFS(testmigrationsPath)

	got, err := MigrationsFromFS(source)
	if err != nil {
		t.Fatal(err)
	}

	expected := MigrationSlice{{
		Name: "0001_20210726_2134_first",
		Num:  1,
		Up:   "CREATE TABLE one (\n    id SERIAL PRIMARY KEY\n);\n",
		Down: "DROP TABLE one;\n",
	}, {
		Name: "0002_20210726_2134_second",
		Num:  2,
		Up:   "CREATE TABLE second (\n    id SERIAL PRIMARY KEY\n);\n",
		Down: "DROP TABLE second;\n",
	}, {
		Name: "0003_20210726_2134_third",
		Num:  3,
		Up:   "CREATE TABLE third (\n    id SERIAL PRIMARY KEY\n);\n",
		Down: "DROP TABLE third;\n",
	}}

	if diff := cmp.Diff(got, expected); diff != "" {
		t.Fatal(diff)
	}
}

func TestMigraitonSlice_NextNum(t *testing.T) {
	source := os.DirFS(testmigrationsPath)

	migrations, err := MigrationsFromFS(source)
	if err != nil {
		t.Fatal(err)
	}

	got := migrations.NextNum()
	want := 4
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatal(diff)
	}
}

func TestMigrationSlice_CreateNext(t *testing.T) {
	source := os.DirFS(testmigrationsPath)

	migrations, err := MigrationsFromFS(source)
	if err != nil {
		t.Fatal(err)
	}

	tmp, err := ioutil.TempDir("", "test-migrations")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	m, err := migrations.CreateNext(tmp, "foobar")
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(m.Name, "_")
	if len(parts) != 4 {
		t.Fatalf("Unexpected amount of rts: %d", len(parts))
	}

	if parts[0] != "0004" {
		t.Errorf("Invalid number part: %s", parts[0])
	}

	if parts[3] != "foobar" {
		t.Errorf("Invalud name part: %s", parts[3])
	}
}

func TestMigrationSlice_ApplyAll(t *testing.T) {
	type Test struct {
		Run func(t *testing.T, db *sqlx.DB, migrations MigrationSlice)
	}

	tables := func(t *testing.T, db *sqlx.DB) []string {
		t.Helper()

		var tables []string
		query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
		if err := db.Select(&tables, query); err != nil {
			t.Fatal(err)
		}

		return tables
	}

	tests := map[string]Test{
		"empty database": {
			Run: func(t *testing.T, db *sqlx.DB, migrations MigrationSlice) {
				if err := migrations.ApplyAll(db, log.Default()); err != nil {
					t.Fatal(err)
				}

				got := tables(t, db)
				expected := []string{
					"schema_version",
					"one",
					"second",
					"third",
				}

				if diff := cmp.Diff(got, expected); diff != "" {
					t.Fatal(diff)
				}
			},
		},
		"initialized database": {
			Run: func(t *testing.T, db *sqlx.DB, migrations MigrationSlice) {
				ctx := context.TODO()
				err := dbutils.WithTransaction(db, ctx, func(tx *sqlx.Tx) error {
					return InitDB(tx)
				})
				if err != nil {
					t.Fatal(err)
				}

				if err := migrations.ApplyAll(db, log.Default()); err != nil {
					t.Fatal(err)
				}

				got := tables(t, db)
				expected := []string{
					"schema_version",
					"one",
					"second",
					"third",
				}

				if diff := cmp.Diff(got, expected); diff != "" {
					t.Fatal(diff)
				}
			},
		},
		"partially migrated": {
			Run: func(t *testing.T, db *sqlx.DB, migrations MigrationSlice) {
				partial := migrations[:1]

				if err := partial.ApplyAll(db, log.Default()); err != nil {
					t.Fatal(err)
				}

				partialGot := tables(t, db)
				partialExpected := []string{
					"schema_version",
					"one",
				}
				if diff := cmp.Diff(partialGot, partialExpected); diff != "" {
					t.Fatal(diff)
				}

				if err := migrations.ApplyAll(db, log.Default()); err != nil {
					t.Fatal(err)
				}

				got := tables(t, db)
				expected := []string{
					"schema_version",
					"one",
					"second",
					"third",
				}

				if diff := cmp.Diff(got, expected); diff != "" {
					t.Fatal(diff)
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			source := os.DirFS(testmigrationsPath)
			migrations, err := MigrationsFromFS(source)
			if err != nil {
				t.Fatal(err)
			}

			dbname := strings.ToLower(t.Name())
			dbname = strings.ReplaceAll(dbname, "/", "_")
			db, drop := createTestDB(t, dbname)
			defer drop(t)

			tt.Run(t, db, migrations)
		})
	}
}
