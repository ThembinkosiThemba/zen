package zen

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// ResponseWriter wraps http.ResponseWriter and records the status code and
// body size. The status line is sent lazily, on the first Write or when the
// handler chain finishes, so middleware can still set headers after a handler
// has chosen a status code.
type ResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func newResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w, status: http.StatusOK, size: -1}
}

// WriteHeader sets the status code. It has no effect once the headers are sent.
func (w *ResponseWriter) WriteHeader(code int) {
	if code > 0 && !w.Written() {
		w.status = code
	}
}

func (w *ResponseWriter) writeHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

// Write sends the headers if needed and writes data to the body.
func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.writeHeaderNow()
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

// Status returns the response status code.
func (w *ResponseWriter) Status() int { return w.status }

// Size returns the number of body bytes written, or -1 if headers have not been sent.
func (w *ResponseWriter) Size() int { return w.size }

// Written reports whether the headers have been sent.
func (w *ResponseWriter) Written() bool { return w.size != -1 }

// Flush implements http.Flusher.
func (w *ResponseWriter) Flush() {
	w.writeHeaderNow()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker, which WebSocket libraries need.
func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("zen: underlying ResponseWriter does not implement http.Hijacker")
	}
	if w.size < 0 {
		w.size = 0
	}
	return h.Hijack()
}

// Unwrap returns the underlying http.ResponseWriter, for http.ResponseController.
func (w *ResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
