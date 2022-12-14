package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Handler that can return a error.
//
// If returned error is a ErrorResponder, it is used to respond to the client.
//
// If error is not a ErrorResponder or the ErrorResponder returns a new error,
// 500 Internal Server Error is written to the client and WithHandlerErrorLogger
// is used to log the error. Is the middleware is not present, the log package
// is used.
//
// No guarantees are made about the logged error, except that it wraps the
// original error.
type Handler func(http.ResponseWriter, *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		// All good.
		return
	}

	if resp, ok := err.(ErrorResponder); ok {
		err = resp.ErrorRespond(w, r)
		if err == nil {
			// ErrorResponder did its job, we're done.
			return
		}
	}

	handlerLog(r, fmt.Errorf("Error handling a request: %w", err))

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type MethodHandler struct {
	handlers map[string]http.Handler
}

func NewMethodHandler() *MethodHandler {
	return &MethodHandler{handlers: make(map[string]http.Handler)}
}

func (mh *MethodHandler) Add(m string, h http.Handler) *MethodHandler {
	mh.handlers[m] = h
	return mh
}

func (mh *MethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, ok := mh.handlers[r.Method]
	if !ok {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	h.ServeHTTP(w, r)
}

// NewJSONHandler wraps Handler and tries to decode the request body into T.
// If decoding fails, bad request status is written to the client.
func NewJSONHandler[T any](fn func(w http.ResponseWriter, r *http.Request, req *T) error) Handler {
	BadRequest := func(err error) error {
		return NewStatusError(http.StatusBadRequest, "", err)
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var body T
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return BadRequest(err)
		}

		return fn(w, r, &body)
	}
}
