package middleware

import (
	"net/http"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// RateLimitConfig holds configuration for the RateLimiter middleware
type RateLimitConfig struct {
	Limit           int           // Maximum requests per window
	Window          time.Duration // Window duration
	BurstLimit      int           // Temporary burst allowance above the limit
	CustomKeyFunc   func(*zen.Context) string
	ExlcudePaths    []string
	StatusCode      int
	CustomErrorFunc func(*zen.Context, time.Duration)
}

// client holds information for each client tracked by the rate limiter
// type client struct {
// 	requestCount int
// 	windowStart  time.Time
// 	burstUsed    int
// 	blocked      bool
// }

// DefaultRateLimiterConfig returns the default configs for RateLimiter
func DefaultRateLimiterConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:      100,
		Window:     time.Minute,
		BurstLimit: 20,
		CustomKeyFunc: func(c *zen.Context) string {
			return c.ClientIP()
		},
		StatusCode: http.StatusTooManyRequests,
	}
}
