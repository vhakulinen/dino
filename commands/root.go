package commands

import (
	"fmt"
	"log"
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

func RootCommand(cmdname string, opts ...Option) *cobra.Command {
	options := Options{
		Logger: log.Default(),
	}

	for _, fn := range opts {
		fn(&options)
	}

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

	cobra.OnInitialize(func() {
		v.SetConfigFile(configFile)
		v.SetConfigType("toml")

		if err := v.ReadInConfig(); err != nil {
			fmt.Printf("Can't read config file: %v\n", err)
			os.Exit(1)
		}
	})

	rootCmd.AddCommand(
		MigrationsCommand(v, &options),
		DatabaseCommands(v, &options),
	)

	return rootCmd
}
