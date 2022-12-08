package httputil

import (
	"fmt"
	"net/http"
)

type ErrorResponder interface {
	ErrorRespond(http.ResponseWriter, *http.Request) error
}

type StatusError struct {
	err    error
	status int
	msg    string
}

func NewStatusError(status int, msg string, err error) *StatusError {
	return &StatusError{
		err:    err,
		status: status,
		msg:    msg,
	}
}

func (e *StatusError) Unwrap() error {
	return e.err
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("StatusError(%d - %s): %s", e.status, e.msg, e.err)
}

func (e *StatusError) ErrorRespond(w http.ResponseWriter, r *http.Request) error {
	msg := e.msg
	if msg == "" {
		msg = http.StatusText(e.status)
	}

	http.Error(w, msg, e.status)

	return nil
}
