package migrations_test

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"

	"github.com/vhakulinen/dino/db/dbtest"
	"github.com/vhakulinen/dino/db/migrations"
	"github.com/vhakulinen/dino/db/tx"
	"github.com/vhakulinen/dino/db/utils"
)

const testmigrationsPath = "testdata/testmigrations"

func TestMigrationsFromFS(t *testing.T) {
	source := os.DirFS(testmigrationsPath)

	got, err := migrations.MigrationsFromFS(source)
	if err != nil {
		t.Fatal(err)
	}

	expected := migrations.MigrationSlice{{
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

	migrations, err := migrations.MigrationsFromFS(source)
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

	migrations, err := migrations.MigrationsFromFS(source)
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
		Run func(t *testing.T, db *sqlx.DB, migrations migrations.MigrationSlice)
	}

	tables := func(t *testing.T, db *sqlx.DB) []string {
		t.Helper()

		// TODO(ville): Context?
		tables, err := utils.QueryAllTableNames(context.TODO(), db)
		if err != nil {
			t.Fatal(err)
		}

		return tables
	}

	tests := map[string]Test{
		"empty database": {
			Run: func(t *testing.T, db *sqlx.DB, migs migrations.MigrationSlice) {
				if err := migs.ApplyAll(db, log.Default()); err != nil {
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
			Run: func(t *testing.T, db *sqlx.DB, migs migrations.MigrationSlice) {
				ctx := context.TODO()
				err := tx.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
					return migrations.EnsureSchema(tx)
				})
				if err != nil {
					t.Fatal(err)
				}

				if err := migs.ApplyAll(db, log.Default()); err != nil {
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
			Run: func(t *testing.T, db *sqlx.DB, migrations migrations.MigrationSlice) {
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

	connParams := dbtest.DefaultConnectionParams

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			source := os.DirFS(testmigrationsPath)
			migrations, err := migrations.MigrationsFromFS(source)
			if err != nil {
				t.Fatal(err)
			}

			dbname := strings.ToLower(t.Name())
			dbname = strings.ReplaceAll(dbname, "/", "_")
			db, drop := dbtest.WithCreateDB(t, &connParams, dbname)
			defer drop(t)

			tt.Run(t, db, migrations)
		})
	}
}
