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
	assert.Equal(t, IPBased, config.Strategy)
}

func TestIPBasedRateLimiting(t *testing.T) {
	config := RateLimitConfig{
		Limit:      2,
		Window:     time.Second,
		BurstLimit: 1,
		Strategy:   IPBased,
	}
	limiter := NewRateLimiter(config)

	// These should be allowed (within base limit)
	assert.True(t, limiter.isAllowed("127.0.0.1"))
	assert.True(t, limiter.isAllowed("127.0.0.1"))

	// Third request should be allowed (uses burst limit)
	assert.True(t, limiter.isAllowed("127.0.0.1"))

	// Fourth request should be blocked
	assert.False(t, limiter.isAllowed("127.0.0.1"))

	time.Sleep(config.Window)

	// First request in new window should be allowed
	assert.True(t, limiter.isAllowed("127.0.0.1"))
}

func TestSlidingWindowRateLimiter(t *testing.T) {
	config := RateLimitConfig{
		Limit:      2,
		Window:     time.Second,
		BurstLimit: 1,
		Strategy:   SlidingWindow,
	}
	limiter := NewRateLimiter(config)

	// These should be allowed
	assert.True(t, limiter.isAllowed("127.0.0.1"))
	assert.True(t, limiter.isAllowed("127.0.0.1"))
	assert.True(t, limiter.isAllowed("127.0.0.1")) // should be allowed due to burst

	// This should be blocked
	assert.False(t, limiter.isAllowed("127.0.0.1"))

	// Sleep for half the window
	time.Sleep(500 * time.Millisecond)

	// Should still be blocked
	assert.False(t, limiter.isAllowed("127.0.0.1"))

	// Sleep for the remaining window
	time.Sleep(time.Second)

	// Should be allowed again
	assert.True(t, limiter.isAllowed("127.0.0.1"))
}
