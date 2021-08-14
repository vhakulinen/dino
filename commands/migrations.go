package commands

import (
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vhakulinen/dino/dbutils"
)

type Migration struct {
	Name string
	Up   string
	Down string
}

type MigrationSlice []*Migration

func MigrationsCommand(v *viper.Viper, opts *Options) *cobra.Command {
	src := func(v *viper.Viper) string {
		return v.GetString("dino.migrations.dir")
	}

	getFS := func(v *viper.Viper) fs.FS {
		return os.DirFS(src(v))
	}

	cmdNew := &cobra.Command{
		Use:   "new [migration name]",
		Short: "Create new migration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			migrations, err := dbutils.MigrationsFromFS(getFS(v))
			if err != nil {
				return err
			}

			m, err := migrations.CreateNext(src(v), strings.Join(args, "_"))
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
			migrations, err := dbutils.MigrationsFromFS(getFS(v))
			if err != nil {
				return err
			}

			db, err := connParamsFromViper(v).Open()
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

	rootCmd.AddCommand(cmdNew, cmdApply)
	return rootCmd
}
