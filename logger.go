package dino

type Logger interface {
	Printf(template string, args ...interface{})
}
