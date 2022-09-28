package utils

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func QueryAllTableNames(ctx context.Context, exec sqlx.QueryerContext) ([]string, error) {
	var tables []string
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
	`

	err := sqlx.SelectContext(ctx, exec, &tables, query)

	return tables, err
}
