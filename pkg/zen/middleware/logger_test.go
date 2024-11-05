package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		expectedStatus int
		checkLog       func(t *testing.T, log string)
	}{
		{
			name:           "Simple GET request",
			method:         http.MethodGet,
			path:           "/test",
			expectedStatus: http.StatusOK,
			checkLog: func(t *testing.T, log string) {
				assert.Contains(t, log, "GET")
				assert.Contains(t, log, "/test")
				assert.Contains(t, log, "200")
			},
		},
		{
			name:           "Request with query parameters",
			method:         http.MethodPost,
			path:           "/test",
			query:          "key=value",
			expectedStatus: http.StatusOK,
			checkLog: func(t *testing.T, log string) {
				assert.Contains(t, log, "POST")
				assert.Contains(t, log, "/test?key=value")
				assert.Contains(t, log, "200")
			},
		},
		{
			name:           "Error request",
			method:         http.MethodGet,
			path:           "/error",
			expectedStatus: http.StatusInternalServerError,
			checkLog: func(t *testing.T, log string) {
				assert.Contains(t, log, "GET")
				assert.Contains(t, log, "/error")
				assert.Contains(t, log, "500")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			buf.Reset()

			// Create test request
			w := httptest.NewRecorder()
			url := tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}
			r := httptest.NewRequest(tt.method, url, nil)

			// Create context and run middleware
			c := zen.NewContext(w, r)
			handler := Logger()

			if strings.Contains(tt.path, "error") {
				c.Status(http.StatusInternalServerError)
			}

			handler(c)

			// Verify log output
			logOutput := buf.String()
			tt.checkLog(t, logOutput)
		})
	}
}

func TestLoggerWithCustomIP(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Real-IP", "1.2.3.4")

	c := zen.NewContext(w, r)
	handler := Logger()
	handler(c)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "1.2.3.4")
}
