package httputil

import "net/http"

// StatusRecorder wraps http.StatusRecorder and records the status code
// written to the client.
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		// Default to 200 OK.
		Status: http.StatusOK,
	}
}

func (rr *StatusRecorder) WriteHeader(status int) {
	rr.Status = status
	rr.ResponseWriter.WriteHeader(status)
}
