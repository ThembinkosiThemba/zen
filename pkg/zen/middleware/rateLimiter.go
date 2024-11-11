package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// RateLimitStrategy defines the type of rate limiting to use
type RateLimitStrategy string

const (
	IPBased       RateLimitStrategy = "ip-based"
	SlidingWindow RateLimitStrategy = "sliding-window"
)

// RateLimitConfig holds configuration for the RateLimiter middleware
type RateLimitConfig struct {
	Limit           int                               // Maximum requests per window
	Window          time.Duration                     // Window duration
	BurstLimit      int                               // Temporary burst allowance above the limit
	CustomKeyFunc   func(*zen.Context) string         // function to generate key for rate limiting
	ExcludePaths    []string                          // paths to exclude from rate limiting
	StatusCode      int                               // HTTP status code for rate limit exceeded
	CustomErrorFunc func(*zen.Context, time.Duration) // Custom error handling
	Strategy        RateLimitStrategy                 // rate limiting strategy to use
}

// DefaultRateLimiterConfig returns the default configs for RateLimiter which for now, we are going to stick to IP-based
func DefaultRateLimiterConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:      100,
		Window:     time.Minute,
		BurstLimit: 20,
		CustomKeyFunc: func(c *zen.Context) string {
			return c.ClientIP()
		},
		StatusCode: http.StatusTooManyRequests,
		Strategy:   IPBased,
	}
}

// windowEntry keeps track of requests in a time window
type windowEntry struct {
	count     int
	startTime time.Time
}

// RateLimiter implements the ratelimiting logic
type RateLimiter struct {
	config  RateLimitConfig
	windows sync.Map // maps keys to windowEntry
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config: config,
	}
}

// RateLimiterMiddleware creates a new rate limiting middleware
func RateLimiterMiddleware(config ...RateLimitConfig) zen.HandlerFunc {
	var cfg RateLimitConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultRateLimiterConfig()
	}

	limiter := NewRateLimiter(cfg)

	return func(c *zen.Context) {
		for _, path := range cfg.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		key := cfg.CustomKeyFunc(c)
		allowed := limiter.isAllowed(key)

		if !allowed {
			if cfg.CustomErrorFunc != nil {
				cfg.CustomErrorFunc(c, cfg.Window)
			} else {
				c.Text(cfg.StatusCode, "Rate limit exceeded. Try again in %v", cfg.Window)
			}
			c.Quit()
			return
		}
		c.Next()
	}

}

// isAllowed checks if the request should be allowed based on the rate limiting strategy
func (rl *RateLimiter) isAllowed(key string) bool {
	now := time.Now()

	switch rl.config.Strategy {
	case SlidingWindow:
		return rl.isSlidingWindowAllowed(key, now)
	default: // IP-based
		return rl.isIPBasedAllowed(key, now)
	}
}

// isIPBasedAllowed implements simple IP-based rate limiting
func (rl *RateLimiter) isIPBasedAllowed(key string, now time.Time) bool {
	val, exists := rl.windows.Load(key)
	if !exists {
		rl.windows.Store(key, &windowEntry{count: 1, startTime: now})
		return true
	}

	entry := val.(*windowEntry)
	if now.Sub(entry.startTime) > rl.config.Window {
		// window has expired, reset counter
		entry.count = 1
		entry.startTime = now
		return true
	}

	// Allow requests within base limit + burst limit
	if entry.count < rl.config.Limit+rl.config.BurstLimit {
		entry.count++
		return true
	}

	return false
}

// isSlidingWindowAllowed implements sliding window rate limiting
func (rl *RateLimiter) isSlidingWindowAllowed(key string, now time.Time) bool {
	val, exists := rl.windows.Load(key)
    if !exists {
        rl.windows.Store(key, &windowEntry{count: 1, startTime: now})
        return true
    }

	entry := val.(*windowEntry)
	windowDuration := now.Sub(entry.startTime)

	if windowDuration > rl.config.Window {
		// get the overlap with the previous window
		// We are calculating how many requests from the previous window should be counted
		overlap := rl.config.Window - (windowDuration - rl.config.Window)
		if overlap < 0 {
			overlap = 0
		}

		weight := float64(overlap) / float64(rl.config.Window)
		previousCount := int(float64(entry.count) * weight)

		// get the requests in current window considering the overlap
		entry.count = previousCount + 1
		entry.startTime = now

		return entry.count < rl.config.Limit+rl.config.BurstLimit
	}

	if entry.count < rl.config.Limit+rl.config.BurstLimit {
		entry.count++
		return true
	}
	return false
}
