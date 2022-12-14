package cli

import (
	"github.com/spf13/viper"

	"github.com/vhakulinen/dino/db/utils"
)

// Config embeds viper and adds utility functions to read dino's settings.
type Config struct {
	*viper.Viper
	opts *options
}

func (c *Config) ConnParams() *utils.ConnectionParams {
	return &utils.ConnectionParams{
		Host:     c.GetString("dino.db.host"),
		Port:     c.GetInt("dino.db.port"),
		Database: c.GetString("dino.db.database"),
		Username: c.GetString("dino.db.username"),
		Password: c.GetString("dino.db.password"),
		SSLMode:  c.GetString("dino.db.sslmod"),
	}
}

func (c *Config) MigrationsDir() string {
	return c.GetString("dino.migrations.dir")
}
