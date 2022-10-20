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

// RootCommand gives entry point for dino's cli. Config will be non-nil, mostly
// empty value which will be populated by the time any cobra commands are executed.
func RootCommand(cmdname string, dbdriver string, opts ...option) (*cobra.Command, *Config) {
	// TODO(ville): Add tests for reading and binding the config.

	var configFile string
	v := viper.New()

	rootCmd := &cobra.Command{
		Use:          cmdname,
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
		v.BindPFlag(key, flag)

		// Bind the env. Replace the dots with underscore for more usable
		// ENV_VARS. E.g. dino.db.host => DINO_DB_HOST.
		v.BindEnv(key, strings.ReplaceAll(strings.ToUpper(key), ".", "_"))
	})

	config := new(Config)
	cobra.OnInitialize(func() {
		v.SetConfigFile(configFile)
		v.SetConfigType("toml")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(*fs.PathError); !ok {
				fmt.Printf("Can't read config file: %v %v\n", ok, err)
				os.Exit(1)
			}
		}

		*config = configFromViper(v, opts...)
	})

	rootCmd.AddCommand(
		MigrationsCommand(v, config, dbdriver),
		DatabaseCommands(v, config, dbdriver),
	)

	return rootCmd, config
}
