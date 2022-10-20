package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vhakulinen/dino/db/fixtures"
	"github.com/vhakulinen/dino/db/tx"
)

func DatabaseCommands(v *viper.Viper, config *Config, dbdriver string) *cobra.Command {

	cmdDump := &cobra.Command{
		Use:   "dump",
		Short: "Dump fixture directly from database",
		RunE: func(cmd *cobra.Command, args []string) error {

			dump, err := fixtures.DumpFixture(config.DbConnParams)

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
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			contents, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			db, err := config.DbConnParams.Open(dbdriver)
			if err != nil {
				return err
			}

			err = tx.WithTransaction(cmd.Context(), db, func(tx *sqlx.Tx) error {
				return fixtures.LoadFixture(cmd.Context(), tx, string(contents))
			})

			return err
		},
	}

	truncateCmd := &cobra.Command{
		Use:   "truncate-all",
		Short: "Truncate all tables in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := config.DbConnParams.Open(dbdriver)
			if err != nil {
				return err
			}

			err = tx.WithTransaction(cmd.Context(), db, func(tx *sqlx.Tx) error {
				config.Logger.Printf("Truncating...")
				return fixtures.TruncateAll(cmd.Context(), tx)
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
