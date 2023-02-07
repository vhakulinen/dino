package cli

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/viper"

	"github.com/vhakulinen/dino/db/utils"
)

// Config embeds viper and adds utility functions to read dino's settings.
type Config struct {
	*viper.Viper
	opts *options
}

func (c *Config) ReadConfigFile() {
	c.SetConfigFile(c.opts.configFile)
	c.SetConfigType("toml")

	if err := c.ReadInConfig(); err != nil {
		if _, ok := err.(*fs.PathError); !ok {
			fmt.Printf("Can't read config file: %v %v\n", ok, err)
			os.Exit(1)
		}
	}
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
