package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// RateLimitConfig holds configuration for the RateLimiter middleware
type RateLimitConfig struct {
	// Request per window
	Limit int

	// Window duration
	Window time.Duration
}

// DefaultRateLimiterConfig returns the default configs for RateLimiter
func DefaultRateLimiterConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:  100,
		Window: time.Minute,
	}
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// RateLimiter creates a rate limiting middleware with configurable limits
func RateLimiter(config ...RateLimitConfig) zen.HandlerFunc {
	cfg := DefaultRateLimiterConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	var mu sync.RWMutex
	visitors := make(map[string]*visitor)

	// go routing for cleaning up
	go func() {
		for {
			time.Sleep(cfg.Window)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > cfg.Window {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *zen.Context) {
		ip := c.ClientIP()
		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}

		if time.Since(v.lastSeen) > cfg.Window {
			v.count = 1
			v.lastSeen = time.Now()
		} else {
			v.count++
		}

		if v.count > cfg.Limit {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error": fmt.Sprintf("Rate limit exceeded. Try again in %v", cfg.Window),
			})
			return
		}

		v.lastSeen = time.Now()
		mu.Unlock()
		c.Next()
	}
}


