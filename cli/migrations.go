package cli

import (
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vhakulinen/dino/db/migrations"
	"github.com/vhakulinen/dino/db/tx"
)

type Migration struct {
	Name string
	Up   string
	Down string
}

type MigrationSlice []*Migration

func MigrationsCommand(v *viper.Viper, config *Config, dbdriver string) *cobra.Command {
	cmdNew := &cobra.Command{
		Use:   "new [migration name]",
		Short: "Create new migration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations, err := migrations.MigrationsFromFS(os.DirFS(config.MigrationsDir))
			if err != nil {
				return err
			}

			m, err := migrations.CreateNext(config.MigrationsDir, strings.Join(args, "_"))
			if err != nil {
				return err
			}

			config.Logger.Printf("Created a new migration '%s'", m.Name)

			return nil
		},
	}

	cmdRevert := &cobra.Command{
		Use:   "revert",
		Short: "Revert the latest migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			migs, err := migrations.MigrationsFromFS(os.DirFS(config.MigrationsDir))
			if err != nil {
				return err
			}

			db, err := connParamsFromViper(v).Open(dbdriver)
			if err != nil {
				return err
			}

			return tx.WithTransaction(cmd.Context(), db, func(tx *sqlx.Tx) error {
				return migs.RevertCurrent(tx)
			})
		},
	}

	cmdApply := &cobra.Command{
		Use:   "apply",
		Short: "Apply all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations, err := migrations.MigrationsFromFS(os.DirFS(config.MigrationsDir))
			if err != nil {
				return err
			}

			db, err := connParamsFromViper(v).Open(dbdriver)
			if err != nil {
				return err
			}

			return migrations.ApplyAll(db, config.Logger)
		},
	}

	rootCmd := &cobra.Command{
		Use:   "migrations",
		Short: "Manage migrations",
	}

	rootCmd.AddCommand(cmdNew, cmdApply, cmdRevert)
	return rootCmd
}
