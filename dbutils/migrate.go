package dbutils

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const format = "20060102_1504"

type Logger interface {
	Printf(template string, args ...interface{})
}

type Migration struct {
	Name string
	Num  int
	Up   string
	Down string
}

type MigrationSlice []*Migration

func MigrationsFromFS(source fs.FS) (MigrationSlice, error) {
	files, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, file := range files {
		if file.Type().IsDir() {
			dirs = append(dirs, file.Name())
		}
	}

	migrations := make(MigrationSlice, len(dirs))
	for i, dirname := range dirs {
		up, err := readFile(source, fmt.Sprintf("%s/up.sql", dirname))
		if err != nil {
			return nil, err
		}

		down, err := readFile(source, fmt.Sprintf("%s/down.sql", dirname))
		if err != nil {
			return nil, err
		}

		parts := strings.Split(dirname, "_")
		if len(parts) < 4 {
			return nil, fmt.Errorf("Malformed migration name: %q", dirname)
		}

		num, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("Failed to decode migration number: %v", err)
		}

		migrations[i] = &Migration{
			Name: dirname,
			Num:  num,
			Up:   string(up),
			Down: string(down),
		}
	}

	return migrations, nil
}

// Return next migration number.
func (slice MigrationSlice) NextNum() int {
	return len(slice) + 1
}

// Create a new migration. baseDir should point to the directory where all the
// migrations live.
func (slice MigrationSlice) CreateNext(baseDir, migrationName string) (*Migration, error) {
	num := slice.NextNum()

	d := time.Now().Format(format)
	name := fmt.Sprintf("%04d_%s_%s", num, d, migrationName)

	dir := baseDir + "/" + name
	if err := os.Mkdir(dir, 0755); err != nil {
		return nil, err
	}

	for _, fname := range []string{"/up.sql", "/down.sql"} {
		f, err := os.Create(dir + fname)
		if err != nil {
			return nil, err
		}

		f.Close()
	}

	return &Migration{
		Name: name,
	}, nil
}

func (slice MigrationSlice) find(num int) *Migration {
	for _, m := range slice {
		if m.Num == num {
			return m
		}
	}

	return nil
}

// Applies all pending migrations to the database.
func (slice MigrationSlice) ApplyAll(db *sqlx.DB, logger Logger) error {
	ctx := context.TODO()

	err := WithTransaction(db, ctx, func(tx *sqlx.Tx) error {
		err := InitDB(tx)
		if err != nil {
			return err
		}

		current, err := QuerySchemaVersion(tx)
		if err != nil {
			return err
		}

		for m := slice.find(current + 1); m != nil; m = slice.find(current + 1) {
			logger.Printf("Applying '%s'...", m.Name)

			_, err := tx.Exec(m.Up)
			if err != nil {
				return err
			}

			if err := setSchemaVersion(tx, m.Num); err != nil {
				return err
			}

			current = m.Num
		}

		return nil
	})

	return err
}

func readFile(fs fs.FS, fname string) ([]byte, error) {
	f, err := fs.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}
