package cli

import (
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"

	"github.com/vhakulinen/dino/db/migrations"
)

func migrationsCommand(config *Config) *cobra.Command {
	cmdNew := &cobra.Command{
		Use:   "new [migration name]",
		Short: "Create new migration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations, err := migrations.MigrationsFromFS(os.DirFS(config.MigrationsDir()))
			if err != nil {
				return err
			}

			m, err := migrations.CreateNext(config.MigrationsDir(), strings.Join(args, "_"))
			if err != nil {
				return err
			}

			config.opts.logger.Printf("Created a new migration '%s'", m.Name)

			return nil
		},
	}

	cmdRevert := &cobra.Command{
		Use:   "revert",
		Short: "Revert the latest migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			migs, err := migrations.MigrationsFromFS(os.DirFS(config.MigrationsDir()))
			if err != nil {
				return err
			}

			db, err := pgx.Connect(cmd.Context(), config.ConnParams().ConnString())
			if err != nil {
				return err
			}

			return pgx.BeginFunc(cmd.Context(), db, func(tx pgx.Tx) error {
				return migs.RevertCurrent(cmd.Context(), tx)
			})
		},
	}

	cmdApply := &cobra.Command{
		Use:   "apply",
		Short: "Apply all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations, err := migrations.MigrationsFromFS(os.DirFS(config.MigrationsDir()))
			if err != nil {
				return err
			}

			db, err := pgx.Connect(cmd.Context(), config.ConnParams().ConnString())
			if err != nil {
				return err
			}

			return migrations.ApplyAll(db, config.opts.logger)
		},
	}

	rootCmd := &cobra.Command{
		Use:   "migrations",
		Short: "Manage migrations",
	}

	rootCmd.AddCommand(cmdNew, cmdApply, cmdRevert)
	return rootCmd
}
