package commands

import (
	"log"

	"github.com/spf13/viper"
	"github.com/vhakulinen/dino/dbutils"
)

type Config struct {
	// Viper instance used by dino.
	Viper *viper.Viper

	DbConnParams  *dbutils.ConnectionParams
	Logger        dbutils.Logger
	MigrationsDir string
}

func configFromViper(v *viper.Viper, opts ...option) Config {
	config := Config{
		Viper: v,
		DbConnParams: &dbutils.ConnectionParams{
			Host:     v.GetString("dino.db.host"),
			Port:     v.GetInt("dino.db.port"),
			Database: v.GetString("dino.db.database"),
			Username: v.GetString("dino.db.username"),
			Password: v.GetString("dino.db.password"),
			SSLMode:  v.GetString("dino.db.sslmod"),
		},
		MigrationsDir: v.GetString("dino.migrations.dir"),

		// Default logger value.
		Logger: log.Default(),
	}

	for _, opt := range opts {
		opt(&config)
	}

	return config
}
