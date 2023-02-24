package fixtures

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vhakulinen/dino/db/utils"
)

var cleanRegexps = []*regexp.Regexp{
	// pg_catalog. statements.
	regexp.MustCompile(`(?m)^SELECT pg_catalog\..*;$`),
	// SET statements.
	regexp.MustCompile(`(?m)^SET .*;$`),
	// Comments.
	regexp.MustCompile(`(?m)^--.*$`),
	// Empty lines.
	regexp.MustCompile(`(?m)^\n`),
}

func cleanDump(dump []byte) []byte {
	for _, re := range cleanRegexps {
		dump = re.ReplaceAll(dump, []byte(""))
	}

	return dump
}

type fixtureDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func LoadFixture(ctx context.Context, conn fixtureDB, fixture string) error {
	_, err := conn.Exec(ctx, fixture)
	if err != nil {
		return err
	}

	return FixSequences(ctx, conn)
}

func DumpFixture(opts *utils.ConnectionParams) ([]byte, error) {
	cmd := exec.Command(
		"pg_dump",
		"-h", opts.Host,
		"-p", strconv.Itoa(opts.Port),
		"-d", opts.Database,
		"-U", opts.Username,
		"--data-only",
		// Exlcude schema_version table.
		"--exclude-table", "schema_version",
		// Don't do each row in their own INSERT.
		"--rows-per-insert", "1000",
		"--column-inserts",
	)
	cmd.Env = []string{"PGPASSWORD=" + opts.Password}

	// Output errors to stderr.
	cmd.Stderr = os.Stderr

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return cleanDump(out.Bytes()), nil
}

func queryAllTableNames(ctx context.Context, conn pgx.Tx) ([]string, error) {
	query := `
		SELECT table_schema || '.' || table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		AND table_type = 'BASE TABLE'
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// Truncates all tables (e.g. removes all data!).
// TODO(ville): Move to the utils package?
func TruncateAll(ctx context.Context, conn pgx.Tx) error {
	tables, err := queryAllTableNames(ctx, conn)
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		return errors.New("No tables to truncate")
	}

	query := fmt.Sprintf(`TRUNCATE %s RESTART IDENTITY`, strings.Join(tables, ","))
	_, err = conn.Exec(ctx, query)

	return err
}

const sequencesQuery = `
SELECT 'SELECT SETVAL(' ||
       quote_literal(quote_ident(PGT.schemaname) || '.' || quote_ident(S.relname)) ||
       ', COALESCE(MAX(' ||quote_ident(C.attname)|| '), 1) ) FROM ' ||
       quote_ident(PGT.schemaname)|| '.'||quote_ident(T.relname)|| ';'
FROM pg_class AS S,
     pg_depend AS D,
     pg_class AS T,
     pg_attribute AS C,
     pg_tables AS PGT
WHERE S.relkind = 'S'
    AND S.oid = D.objid
    AND D.refobjid = T.oid
    AND D.refobjid = C.attrelid
    AND D.refobjsubid = C.attnum
    AND T.relname = PGT.tablename
ORDER BY S.relname;`

// Attemps to fix sequences based on the current values.
// See: https://wiki.postgresql.org/wiki/Fixing_Sequences
func FixSequences(ctx context.Context, conn fixtureDB) error {
	rows, err := conn.Query(ctx, sequencesQuery)
	if err != nil {
		return err
	}

	stmts, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, strings.Join(stmts, ""))
	return err
}
