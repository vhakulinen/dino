package commands

func OptionMigrationsDir(dir string) Option {
	return func(opts *Options) {
		opts.MigrationsDir = dir
	}
}

func OptionOpenDB(fn OpenDBFn) Option {
	return func(opts *Options) {
		opts.OpenDB = fn
	}
}
