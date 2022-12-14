package cli

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// New gives entry point for dino's cli.
func New(opts ...option) (*cobra.Command, *Config) {
	// TODO(ville): Add tests for reading and binding the config.

	var configFile string
	c := &Config{viper.New(), newOptions(opts...)}

	rootCmd := &cobra.Command{
		Use:          c.opts.cmdName,
		SilenceUsage: true,
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "dino.toml", "Config file")

	rootCmd.PersistentFlags().StringP("db-host", "", "localhost", "Database host")
	rootCmd.PersistentFlags().IntP("db-port", "", 5432, "Database port")
	rootCmd.PersistentFlags().StringP("db-username", "", "postgres", "Database username")
	rootCmd.PersistentFlags().StringP("db-password", "", "password", "Database password")
	rootCmd.PersistentFlags().StringP("db-sslmode", "", "disable", "Database sslmode")
	rootCmd.PersistentFlags().StringP("db-database", "", "postgres", "Database name")

	rootCmd.PersistentFlags().StringP("migrations-dir", "", "migrations", "Directory where migrations are placed")

	// Bind all the flags to viper and env.
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		// Bind the flags.
		key := "dino." + strings.ReplaceAll(flag.Name, "-", ".")
		c.BindPFlag(key, flag)

		// Bind the env. Replace the dots with underscore for more usable
		// ENV_VARS. E.g. dino.db.host => DINO_DB_HOST.
		c.BindEnv(key, strings.ReplaceAll(strings.ToUpper(key), ".", "_"))
	})

	cobra.OnInitialize(func() {
		c.SetConfigFile(configFile)
		c.SetConfigType("toml")

		if err := c.ReadInConfig(); err != nil {
			if _, ok := err.(*fs.PathError); !ok {
				fmt.Printf("Can't read config file: %v %v\n", ok, err)
				os.Exit(1)
			}
		}
	})

	rootCmd.AddCommand(
		migrationsCommand(c),
		databaseCommands(c),
	)

	return rootCmd, c
}
