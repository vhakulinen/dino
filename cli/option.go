package cli

import (
	"log"

	"github.com/vhakulinen/dino/db/migrations"
)

type options struct {
	logger   migrations.Logger
	cmdName  string
	dbDriver string
}

func newOptions(opts ...option) *options {
	// Initialize with default values.
	options := &options{
		logger:   log.Default(),
		cmdName:  "dino",
		dbDriver: "psql",
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type option func(*options)

// Set the logger for the database migrations.
func OptionMigrationsLogger(logger migrations.Logger) option {
	return func(opts *options) {
		opts.logger = logger
	}
}

// Set the cmd name.
func OptionCmdName(name string) option {
	return func(opts *options) {
		opts.cmdName = name
	}
}

// Set the database driver.
func OptionDbDriver(driver string) option {
	return func(opts *options) {
		opts.dbDriver = driver
	}
}
