package httputil

import (
	"context"
	"log"
	"net/http"
)

type contextKey string

const loggerContextKey = contextKey("logger")

type Logger func(error)

func WithLogger(fn Logger) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loggerContextKey, fn)
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}

// Log error using logger from `WithLogger` middleware. If no logger is
// provided, the default `log` package logger is used.
func Log(r *http.Request, err error) {
	fn := r.Context().Value(loggerContextKey)
	if fn == nil {
		log.Println(err)
	} else {
		fn.(Logger)(err)
	}
}
