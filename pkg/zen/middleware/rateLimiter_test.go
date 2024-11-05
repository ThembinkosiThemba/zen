package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.Equal(t, 100, config.Limit)
	assert.Equal(t, time.Minute, config.Window)
	assert.Equal(t, 20, config.BurstLimit)
	assert.NotNil(t, config.CustomKeyFunc)
	assert.Equal(t, http.StatusTooManyRequests, config.StatusCode)
}
