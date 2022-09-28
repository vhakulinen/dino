package cli

import (
	"github.com/vhakulinen/dino/db/migrations"
)

type option func(*Config)

func OptionMigrationsLogger(logger migrations.Logger) option {
	return func(opts *Config) {
		opts.Logger = logger
	}
}
