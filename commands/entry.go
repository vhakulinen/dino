package commands

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	"github.com/vhakulinen/dino/logger"
)

type OpenDBFn func() (*sqlx.DB, error)

type Options struct {
	//Run func() error
	MigrationsDir string
	OpenDB        OpenDBFn
	Logger        logger.Logger
}

type Option func(*Options)

func Commands(opts ...Option) []*cobra.Command {
	options := Options{
		MigrationsDir: "migrations",
		Logger:        log.Default(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return []*cobra.Command{
		MigrationsCommand(&options),
	}
}
