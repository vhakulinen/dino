package commands

import "github.com/vhakulinen/dino/dbutils"

type Options struct {
	Logger dbutils.Logger
}

type Option func(*Options)

func OptionLogger(logger dbutils.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}
