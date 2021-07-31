package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"github.com/vhakulinen/dino/commands"
)

func dbConnString() string {
	// TODO(ville): Read these from somewhere.
	return "user='postgres' password='password' dbname='dino' host='localhost' port='5432' sslmode='disable'"
}

func main() {
	entry := commands.Commands(
		commands.OptionMigrationsDir("./migrations"),
		commands.OptionOpenDB(func() (*sqlx.DB, error) {
			connstr := dbConnString()
			return sqlx.Open("postgres", connstr)
		}),
	)

	root := &cobra.Command{Use: "dino-dev"}
	root.AddCommand(entry...)
	root.Execute()
}
