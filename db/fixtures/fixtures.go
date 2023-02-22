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

	"github.com/jmoiron/sqlx"
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

func LoadFixture(ctx context.Context, exec sqlx.ExtContext, fixture string) error {
	_, err := exec.ExecContext(ctx, fixture)
	if err != nil {
		return err
	}

	return FixSequences(ctx, exec)
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

// TODO(ville): Move to the utils package?
func queryAllTableNames(ctx context.Context, exec sqlx.QueryerContext) ([]string, error) {
	var tables []string
	query := `
		SELECT table_schema || '.' || table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		AND table_type = 'BASE TABLE'
	`

	err := sqlx.SelectContext(ctx, exec, &tables, query)

	return tables, err
}

// Truncates all tables (e.g. removes all data!).
// TODO(ville): Move to the utils package?
func TruncateAll(ctx context.Context, exec sqlx.ExtContext) error {
	tables, err := queryAllTableNames(ctx, exec)
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		return errors.New("No tables to truncate")
	}

	query := fmt.Sprintf(`TRUNCATE %s RESTART IDENTITY`, strings.Join(tables, ","))
	_, err = exec.ExecContext(ctx, query)

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
func FixSequences(ctx context.Context, exec sqlx.ExtContext) error {
	var stmts []string
	if err := sqlx.SelectContext(ctx, exec, &stmts, sequencesQuery); err != nil {
		return err
	}

	_, err := exec.ExecContext(ctx, strings.Join(stmts, ""))
	return err
}
