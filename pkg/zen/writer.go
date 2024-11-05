package zen

import "net/http"

// ResponseWriter wraps http.ResponseWriter to capture the status code
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader captures the status code and calls the underlying ResponseWriter
func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Status returns the HTTP status code
func (w *ResponseWriter) Status() int {
	if w.StatusCode == 0 {
		return http.StatusOK
	}
	return w.StatusCode
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w, StatusCode: 0}
}
