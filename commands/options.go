package commands

import "github.com/vhakulinen/dino/dbutils"

type option func(*Config)

func OptionLogger(logger dbutils.Logger) option {
	return func(opts *Config) {
		opts.Logger = logger
	}
}
