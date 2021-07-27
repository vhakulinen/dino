package logger

type Logger interface {
	Printf(template string, args ...interface{})
}
