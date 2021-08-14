package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vhakulinen/dino/dbutils"
)

func DatabaseCommands(v *viper.Viper, config *Config) *cobra.Command {

	cmdDump := &cobra.Command{
		Use:   "dump",
		Short: "Dump fixture directly from database",
		RunE: func(cmd *cobra.Command, args []string) error {

			dump, err := dbutils.DumpFixture(connParamsFromViper(v))

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

			db, err := connParamsFromViper(v).Open()
			if err != nil {
				return err
			}

			err = dbutils.WithTransaction(cmd.Context(), db, func(tx *sqlx.Tx) error {
				return dbutils.LoadFixture(cmd.Context(), tx, string(contents))
			})

			return err
		},
	}

	truncateCmd := &cobra.Command{
		Use:   "truncate-all",
		Short: "Truncate all tables in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := connParamsFromViper(v).Open()
			if err != nil {
				return err
			}

			err = dbutils.WithTransaction(cmd.Context(), db, func(tx *sqlx.Tx) error {
				config.Logger.Printf("Truncating...")
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
