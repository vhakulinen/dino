package commands

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	"github.com/vhakulinen/dino/dbutils"
)

func DatabaseCommands(opts *Options) *cobra.Command {

	cmdDump := &cobra.Command{
		Use:   "dump",
		Short: "Dump fixture directly from database",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(ville): Read these from somewhere.
			dump, err := dbutils.DumpFixture(&dbutils.DumpFixtureOpts{
				Host:     "localhost",
				Port:     "5432",
				Username: "postgres",
				Password: "password",
				Database: "dino",
			})

			if err != nil {
				return err
			}

			// TODO(ville): Better way to output?
			fmt.Printf("%s\n", dump)

			return nil
		},
	}

	fixtureLoadCmd := &cobra.Command{
		Use:   "fixture-load",
		Short: "Load fixture into the database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.OpenDB == nil {
				return errors.New("OpenDB option missing")
			}

			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			contents, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			db, err := opts.OpenDB()
			if err != nil {
				return err
			}

			err = dbutils.WithTransaction(db, cmd.Context(), func(tx *sqlx.Tx) error {
				return dbutils.LoadFixture(cmd.Context(), tx, string(contents))
			})

			return err
		},
	}

	truncateCmd := &cobra.Command{
		Use:   "truncate-all",
		Short: "Truncate all tables in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.OpenDB == nil {
				return errors.New("OpenDB option missing")
			}

			db, err := opts.OpenDB()
			if err != nil {
				return err
			}

			err = dbutils.WithTransaction(db, cmd.Context(), func(tx *sqlx.Tx) error {
				opts.Logger.Printf("Truncating...")
				return dbutils.TruncateAll(cmd.Context(), tx)
			})

			return err
		},
	}

	rootCmd := &cobra.Command{
		Use:   "db",
		Short: "Database utilities",
	}

	rootCmd.AddCommand(cmdDump, fixtureLoadCmd, truncateCmd)

	return rootCmd
}
