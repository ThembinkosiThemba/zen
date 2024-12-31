package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.Equal(t, []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	}, config.AllowMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Accept"}, config.AllowHeaders)
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, []string{}, config.ExposeHeaders)
	assert.Equal(t, int(12*time.Hour.Seconds()), config.MaxAge)
}

