package commands

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vhakulinen/dino/migrate"
)

type Migration struct {
	Name string
	Up   string
	Down string
}

type MigrationSlice []*Migration

func MigrationsCommand(opts *Options) *cobra.Command {
	src := opts.MigrationsDir

	getFS := func(opts *Options) fs.FS {
		return os.DirFS(src)
	}

	cmdNew := &cobra.Command{
		Use:   "new [migration name]",
		Short: "Create new migration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations, err := migrate.MigrationsFromFS(getFS(opts))
			if err != nil {
				return err
			}

			m, err := migrations.CreateNext(src, strings.Join(args, "_"))
			if err != nil {
				return err
			}

			opts.Logger.Printf("Created a new migration '%s'", m.Name)

			return nil
		},
	}

	cmdApply := &cobra.Command{
		Use:   "apply",
		Short: "Apply all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations, err := migrate.MigrationsFromFS(getFS(opts))
			if err != nil {
				return err
			}

			if opts.OpenDB == nil {
				return errors.New("OpenDB option missing")
			}

			db, err := opts.OpenDB()
			if err != nil {
				return err
			}

			return migrations.ApplyAll(db, opts.Logger)
		},
	}

	rootCmd := &cobra.Command{
		Use:   "migrations",
		Short: "Manage migrations",
	}

	rootCmd.PersistentFlags().StringVarP(&src, "migrations-dir", "m", src, "Directory where migrations live")

	rootCmd.AddCommand(cmdNew, cmdApply)
	return rootCmd
}
