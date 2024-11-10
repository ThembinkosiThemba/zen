package middleware

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
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

	// These should be allowed
	assert.True(t, limiter.isAllowed("127.0.0.1"))
	assert.True(t, limiter.isAllowed("127.0.0.1"))

	// third request should be blocked
	assert.False(t, limiter.isAllowed("127.0.0.1"))

	time.Sleep(config.Window)

	// first request in new window should be allowed
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

	// First two requests should be allowed
	assert.True(t, limiter.isAllowed("127.0.0.1"))
	assert.True(t, limiter.isAllowed("127.0.0.1"))

	// Third request should be blocked
	assert.False(t, limiter.isAllowed("127.0.0.1"))

	// Wait for some time within the window
	time.Sleep(time.Millisecond * 500)

	// Fourth request should be allowed (due to sliding window)
	assert.True(t, limiter.isAllowed("127.0.0.1"))
}

func TestCustomKeyFunction(t *testing.T) {
	config := RateLimitConfig{
		Limit:      2,
		Window:     time.Second,
		BurstLimit: 1,
		CustomKeyFunc: func(c *zen.Context) string {
			return c.GetQueryParam("user_id")
		},
		Strategy: IPBased,
	}
	limiter := NewRateLimiter(config)

	c1 := &zen.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: "user_id=123",
			},
		},
	}
	c2 := &zen.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: "user_id=456",
			},
		},
	}

	// Requests for different users should be limited separately
	assert.True(t, limiter.isAllowed(config.CustomKeyFunc(c1)))
	assert.True(t, limiter.isAllowed(config.CustomKeyFunc(c2)))
	assert.False(t, limiter.isAllowed(config.CustomKeyFunc(c1)))
	assert.False(t, limiter.isAllowed(config.CustomKeyFunc(c2)))
}

func TestExcludedPaths(t *testing.T) {
	config := RateLimitConfig{
		Limit:        2,
		Window:       time.Second,
		BurstLimit:   1,
		ExcludePaths: []string{"/healthz", "/metrics"},
		Strategy:     IPBased,
	}
	limiter := NewRateLimiter(config)

	c1 := &zen.Context{
		Request: &http.Request{
			URL: &url.URL{
				Path: "/healthz",
			},
		},
	}
	c2 := &zen.Context{
		Request: &http.Request{
			URL: &url.URL{
				Path: "/metrics",
			},
		},
	}
	c3 := &zen.Context{
		Request: &http.Request{
			URL: &url.URL{
				Path: "/api/v1/users",
			},
		},
	}

	// Requests to excluded paths should not be rate limited
	assert.True(t, limiter.isAllowed(config.CustomKeyFunc(c1)))
	assert.True(t, limiter.isAllowed(config.CustomKeyFunc(c2)))

	// Requests to non-excluded paths should be rate limited
	assert.True(t, limiter.isAllowed(config.CustomKeyFunc(c3)))
	assert.True(t, limiter.isAllowed(config.CustomKeyFunc(c3)))
	assert.False(t, limiter.isAllowed(config.CustomKeyFunc(c3)))
}
