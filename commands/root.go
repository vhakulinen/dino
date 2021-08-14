package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vhakulinen/dino/dbutils"
)

func connParamsFromViper(v *viper.Viper) *dbutils.ConnectionParams {
	return &dbutils.ConnectionParams{
		Host:     v.GetString("dino.db.host"),
		Port:     v.GetInt("dino.db.port"),
		Database: v.GetString("dino.db.database"),
		Username: v.GetString("dino.db.username"),
		Password: v.GetString("dino.db.password"),
		SSLMode:  v.GetString("dino.db.sslmod"),
	}
}

// RootCommand gives entry point for dino's cli. Config will be non-nil, mostly
// empty value which will be populated by the time any cobra commands are executed.
func RootCommand(cmdname string, opts ...option) (*cobra.Command, *Config) {
	var configFile string
	v := viper.New()

	rootCmd := &cobra.Command{
		Use: cmdname,
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "dino.toml", "Config file")

	rootCmd.PersistentFlags().StringP("db-host", "", "localhost", "Database host")
	rootCmd.PersistentFlags().IntP("db-port", "", 5432, "Database port")
	rootCmd.PersistentFlags().StringP("db-username", "", "postgres", "Database username")
	rootCmd.PersistentFlags().StringP("db-password", "", "password", "Database password")
	rootCmd.PersistentFlags().StringP("db-sslmode", "", "disable", "Database sslmode")
	rootCmd.PersistentFlags().StringP("db-database", "", "postgres", "Database name")

	rootCmd.PersistentFlags().StringP("migrations-dir", "", "migrations", "Directory where migrations are placed")

	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		key := "dino." + strings.ReplaceAll(flag.Name, "-", ".")
		// Bind config with prefix.
		v.BindPFlag(key, flag)
		// Bind env with prefix. Remember to replace the dots with underscore
		// for more usable ENV_VARS. E.g. dino.db.host => DINO_DB_HOST.
		v.BindEnv(key, strings.ReplaceAll(strings.ToUpper(key), ".", "_"))
	})

	config := new(Config)
	cobra.OnInitialize(func() {
		v.SetConfigFile(configFile)
		v.SetConfigType("toml")

		if err := v.ReadInConfig(); err != nil {
			fmt.Printf("Can't read config file: %v\n", err)
			os.Exit(1)
		}

		*config = configFromViper(v, opts...)
	})

	rootCmd.AddCommand(
		MigrationsCommand(v, config),
		DatabaseCommands(v, config),
	)

	return rootCmd, config
}
