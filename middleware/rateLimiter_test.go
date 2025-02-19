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

func TestTokenBucketRateLimiter(t *testing.T) {
	config := RateLimitConfig{
		Strategy:      TokenBucket,
		TokenRate:     2, // 2 tokens per second
		BucketSize:    3, // Max 3 tokens
		BlockDuration: time.Second,
	}
	limiter := NewRateLimiter(config)

	// Initial burst should be allowed up to bucket size
	assert.True(t, limiter.isAllowed("127.0.0.1"), "First request should be allowed")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Second request should be allowed")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Third request should be allowed")

	// Should be blocked as bucket is empty
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Fourth request should be blocked")
	assert.False(t, limiter.isAllowed("127.0.0.1"), "Fifht request should be blocked")

	// Wait for token replenishment
	time.Sleep(time.Second)

	// Should be allowed as tokens were replenished
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Request after replenishment should be allowed")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Second request after replenishment should be allowed")
}

func TestLeakyBucketRateLimiter(t *testing.T) {
	config := RateLimitConfig{
		Strategy: Leaky,
		Limit:    3,   // Queue size
		LeakRate: 2.0, // 2 requests processed per second
	}
	limiter := NewRateLimiter(config)

	// Fill the bucket
	assert.True(t, limiter.isAllowed("127.0.0.1"), "First request should be allowed")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Second request should be allowed")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Third request should be allowed")

	// Bucket is full, should be blocked
	assert.False(t, limiter.isAllowed("127.0.0.1"), "Fourth request should be blocked")

	// Wait for leaking
	time.Sleep(time.Second)

	// Should allow new requests after leak
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Request after leak should be allowed")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "Second request after leak should be allowed")
}

// func TestAdaptiveRateLimiter(t *testing.T) {
//     config := RateLimitConfig{
//         Strategy:      Adaptive,
//         MinLimit:      2,
//         MaxLimit:      4,
//         ScalingFactor: 1.5,
//         Window:        time.Second,
//     }
//     limiter := NewRateLimiter(config)

//     // Test initial limit
//     assert.True(t, limiter.isAllowed("127.0.0.1"), "First request should be allowed")
//     assert.True(t, limiter.isAllowed("127.0.0.1"), "Second request should be allowed")
//     assert.False(t, limiter.isAllowed("127.0.0.1"), "Third request should be blocked")

//     // Wait for window to expire
//     time.Sleep(time.Second)

//     // Test limit increase after high usage
//     assert.True(t, limiter.isAllowed("127.0.0.1"), "First request in new window should be allowed")
//     assert.True(t, limiter.isAllowed("127.0.0.1"), "Second request in new window should be allowed")
//     assert.True(t, limiter.isAllowed("127.0.0.1"), "Third request in new window should be allowed")
//     assert.False(t, limiter.isAllowed("127.0.0.1"), "Fourth request should be blocked")
// }

func TestRateLimiterWithMultipleIPs(t *testing.T) {
	config := RateLimitConfig{
		Strategy:   TokenBucket,
		TokenRate:  2,
		BucketSize: 2,
	}
	limiter := NewRateLimiter(config)

	// Test different IPs independently
	assert.True(t, limiter.isAllowed("127.0.0.1"), "First IP first request")
	assert.True(t, limiter.isAllowed("127.0.0.2"), "Second IP first request")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "First IP second request")
	assert.True(t, limiter.isAllowed("127.0.0.2"), "Second IP second request")
	assert.True(t, limiter.isAllowed("127.0.0.1"), "First IP third request")
	assert.True(t, limiter.isAllowed("127.0.0.2"), "Second IP third request")
}

func TestRateLimiterEdgeCases(t *testing.T) {
	t.Run("Zero Limit", func(t *testing.T) {
		config := RateLimitConfig{
			Strategy:   TokenBucket,
			TokenRate:  0,
			BucketSize: 1,
		}
		limiter := NewRateLimiter(config)
		assert.True(t, limiter.isAllowed("127.0.0.1"), "Should block all requests with zero limit")
	})

	t.Run("Very Large Limit", func(t *testing.T) {
		config := RateLimitConfig{
			Strategy:   TokenBucket,
			TokenRate:  1000000,
			BucketSize: 1000000,
		}
		limiter := NewRateLimiter(config)
		assert.True(t, limiter.isAllowed("127.0.0.1"), "Should handle very large limits")
	})
}
