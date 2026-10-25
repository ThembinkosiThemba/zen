package middleware

import (
	"maps"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ThembinkosiThemba/zen"
)

// RateLimitConfig configures the rate limit middleware.
type RateLimitConfig struct {
	// Limit is the number of requests allowed per key per Window. Defaults to 100.
	Limit int

	// Window is the rate limit window. Defaults to one minute.
	Window time.Duration

	// KeyFunc returns the key requests are counted under. Defaults to the client IP.
	KeyFunc func(*zen.Context) string

	// LimitReached writes the response when a request is rejected.
	// Defaults to 429 Too Many Requests.
	LimitReached func(*zen.Context)
}

// DefaultRateLimitConfig returns a config allowing 100 requests per minute per client IP.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:   100,
		Window:  time.Minute,
		KeyFunc: func(c *zen.Context) string { return c.GetClientIP() },
		LimitReached: func(c *zen.Context) {
			c.Text(http.StatusTooManyRequests, "rate limit exceeded")
		},
	}
}

// RateLimit returns rate limit middleware using DefaultRateLimitConfig.
func RateLimit() zen.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig())
}

// RateLimitWithConfig returns rate limit middleware with the given config.
// It uses a sliding window counter held in memory, so limits are per process.
// Rejected requests get a Retry-After header and the chain is stopped.
func RateLimitWithConfig(config RateLimitConfig) zen.HandlerFunc {
	defaults := DefaultRateLimitConfig()
	if config.Limit <= 0 {
		config.Limit = defaults.Limit
	}
	if config.Window <= 0 {
		config.Window = defaults.Window
	}
	if config.KeyFunc == nil {
		config.KeyFunc = defaults.KeyFunc
	}
	if config.LimitReached == nil {
		config.LimitReached = defaults.LimitReached
	}

	limiter := newRateLimiter(config.Limit, config.Window)

	return func(c *zen.Context) {
		ok, retryAfter := limiter.allow(config.KeyFunc(c), time.Now())
		if ok {
			return
		}
		c.SetHeader("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
		config.LimitReached(c)
		c.Quit()
	}
}

// slidingWindow estimates the request rate by weighting the previous window's
// count by how much of it still overlaps the sliding window.
type slidingWindow struct {
	start      time.Time
	prev, curr int
}

type rateLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	entries   map[string]*slidingWindow
	lastSweep time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, entries: make(map[string]*slidingWindow)}
}

// allow records a request for key and reports whether it is within the limit.
// If not, it also returns how long until a request is likely to be allowed.
func (rl *rateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.sweep(now)

	w := rl.entries[key]
	if w == nil {
		w = &slidingWindow{start: now}
		rl.entries[key] = w
	}

	if elapsed := now.Sub(w.start); elapsed >= rl.window {
		if elapsed < 2*rl.window {
			w.prev = w.curr
		} else {
			w.prev = 0
		}
		w.curr = 0
		w.start = w.start.Add(elapsed.Truncate(rl.window))
	}

	elapsed := now.Sub(w.start)
	weight := 1 - float64(elapsed)/float64(rl.window)
	if float64(w.prev)*weight+float64(w.curr) >= float64(rl.limit) {
		return false, rl.window - elapsed
	}
	w.curr++
	return true, 0
}

// sweep drops keys that have been idle for two windows, so memory does not
// grow without bound. It runs at most once per window.
func (rl *rateLimiter) sweep(now time.Time) {
	if now.Sub(rl.lastSweep) < rl.window {
		return
	}
	rl.lastSweep = now
	maps.DeleteFunc(rl.entries, func(_ string, w *slidingWindow) bool {
		return now.Sub(w.start) >= 2*rl.window
	})
}
