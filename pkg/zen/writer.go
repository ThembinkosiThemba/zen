package zen

import "net/http"

// ResponseWriter wraps http.ResponseWriter to capture the status code
type ResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

// WriteHeader captures the status code and calls the underlying ResponseWriter
func (w *ResponseWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

// Status returns the HTTP status code
func (w *ResponseWriter) Status() int {
    if w.statusCode == 0 {
        return http.StatusOK
    }
    return w.statusCode
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
    return &ResponseWriter{ResponseWriter: w, statusCode: 0}
}