package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.Contains(t, config.AllowMethods, http.MethodGet)
	assert.Contains(t, config.AllowMethods, http.MethodPost)
	assert.Empty(t, config.AllowHeaders)
	assert.False(t, config.AllowCredentials)
	assert.Empty(t, config.ExposeHeaders)
	assert.Zero(t, config.MaxAge)
}

func TestCORSWithConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         CORSConfig
		method         string
		origin         string
		expectedHeader map[string]string
		expectedStatus int
	}{
		{
			name:           "Simple request with wildcard origin",
			config:         CORSConfig{AllowOrigins: []string{"*"}},
			method:         http.MethodGet,
			origin:         "http://example.com",
			expectedHeader: map[string]string{"Access-Control-Allow-Origin": "*"},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Preflight request with specific origin",
			config: CORSConfig{
				AllowOrigins:     []string{"http://example.com"},
				AllowMethods:     []string{http.MethodGet, http.MethodPost},
				AllowHeaders:     []string{"Content-Type"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			method: http.MethodOptions,
			origin: "http://example.com",
			expectedHeader: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Methods":     "GET,POST",
				"Access-Control-Allow-Headers":     "Content-Type",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "3600",
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Request with unauthorized origin",
			config: CORSConfig{
				AllowOrigins: []string{"http://allowed.com"},
			},
			method: http.MethodGet,
			origin: "http://unauthorized.com",
			expectedHeader: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Request without origin",
			config: CORSConfig{
				AllowOrigins: []string{"*"},
			},
			method:         http.MethodGet,
			origin:         "",
			expectedHeader: map[string]string{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}

			c := zen.NewContext(w, r)
			handler := CORSWithConfig(tt.config)
			handler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDefaultCors(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "http://example.com")

	c := zen.NewContext(w, r)
	handler := DefaultCors()
	handler(c)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}
