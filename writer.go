package zen

import "net/http"

// ResponseWriter wraps the http.ResponseWriter to capture and manage the HTTP status code
// along with ensuring the status code is written only once during the request-response cycle.
type ResponseWriter struct {
	http.ResponseWriter      // The embedded ResponseWriter used for writing the response.
	StatusCode          int  // Captures the HTTP status code set for the response.
	headerWritten       bool // A flag to indicate if the header has already been written.
}

// WriteHeader captures the status code and calls the underlying ResponseWriter's WriteHeader method.
// It ensures that the status code is written only once and prevents overwriting.
func (w *ResponseWriter) WriteHeader(statusCode int) {
	// Prevents writing the header more than once.
	if w.headerWritten {
		return
	}
	w.StatusCode = statusCode                // Set the status code for the response.
	w.ResponseWriter.WriteHeader(statusCode) // Calls the underlying ResponseWriter's WriteHeader.
	w.headerWritten = true                   // Marks the header as written to prevent further writes.
}

// Status returns the current HTTP status code. If no status code is set, it returns http.StatusOK.
func (w *ResponseWriter) Status() int {
	// If no status code has been set, return HTTP Status OK (200).
	if w.StatusCode == 0 {
		return http.StatusOK
	}
	return w.StatusCode
}

// NewResponseWriter creates and returns a new ResponseWriter instance.
// It wraps the provided http.ResponseWriter and initializes the StatusCode to 0 and headerWritten flag to false.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,     // The original ResponseWriter that will be wrapped.
		StatusCode:     0,     // Default status code (0 means not yet set).
		headerWritten:  false, // Flag to track if the header has been written.
	}
}
