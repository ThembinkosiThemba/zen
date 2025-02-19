package middleware

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/ThembinkosiThemba/zen"
)

/*
RateLimitStrategy defines different approaches to rate limiting:

1. IP-Based: Simple counter per IP address
2. Sliding Window: Smooth transition between time windows
3. Token Bucket: Token-based rate limiting with burst allowance
4. Leaky Bucket: Constant rate output with queue
5. Adaptive: Dynamic limits based on usage patterns
*/
type RateLimitStrategy string

/*
Strategy Details and Use Cases:

1. IP-Based Rate Limiting
   ----------------------
   Description:
   - Simplest form of rate limiting
   - Maintains a counter per IP address within a fixed time window
   - Counter resets completely when window expires

   How it works:
   - Each IP gets a counter and timestamp
   - If window expires, reset counter
   - If counter < limit, increment and allow
   - If counter >= limit, deny

   Best for:
   - Simple API protection
   - When exact precision isn't required
   - Basic DDoS protection

   Example config:
   ```go
   config := RateLimitConfig{
       Strategy: IPBased,
       Limit:    100,        // 100 requests
       Window:   time.Minute // per minute
   }
   ```

2. Sliding Window Rate Limiting
   ---------------------------
   Description:
   - More accurate than fixed window
   - Considers overlap between time windows
   - Prevents edge-case bursts

   How it works:
   - Maintains current window counter
   - When window slides, calculates overlap
   - New count = (old_count * overlap_ratio) + current_count

   Best for:
   - More precise rate limiting
   - Preventing window edge bursts
   - Smoother request distribution

   Example config:
   ```go
   config := RateLimitConfig{
       Strategy:   SlidingWindow,
       Limit:      100,
       Window:     time.Minute,
       BurstLimit: 20
   }
   ```

3. Token Bucket Rate Limiting
   -------------------------
   Description:
   - Tokens are added to bucket at fixed rate
   - Each request consumes one token
   - Bucket has maximum capacity
   - Allows temporary bursts

   How it works:
   - Tokens replenish at TokenRate
   - If bucket has tokens, allow request
   - If bucket empty, deny request
   - Bucket never exceeds BucketSize

   Best for:
   - API rate limiting
   - Allowing controlled bursts
   - Smooth request processing

   Example config:
   ```go
   config := RateLimitConfig{
       Strategy:    TokenBucket,
       TokenRate:   10,         // 10 tokens per second
       BucketSize:  100,        // Max 100 tokens
       BlockDuration: time.Minute
   }
   ```

4. Leaky Bucket Rate Limiting
   -------------------------
   Description:
   - Requests fill a bucket
   - Bucket leaks at constant rate
   - Fixed queue size
   - Smooths out traffic spikes

   How it works:
   - Requests enter bucket
   - Bucket processes at LeakRate
   - If bucket full, deny requests
   - Constant outflow rate

   Best for:
   - Traffic shaping
   - Network bandwidth control
   - Queue-based processing

   Example config:
   ```go
   config := RateLimitConfig{
       Strategy: Leaky,
       Limit:    100,      // Bucket size
       LeakRate: 1.0       // Leak 1 request per second
   }
   ```
*/

const (
	IPBased       RateLimitStrategy = "ip-based"
	SlidingWindow RateLimitStrategy = "sliding-window"
	TokenBucket   RateLimitStrategy = "token-bucket"
	Leaky         RateLimitStrategy = "leaky-bucket"
	Adaptive      RateLimitStrategy = "adaptive"
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

	IPWhitelist   []string       // IPs exempt from rate limiting
	IPBlacklist   []string       // IPs always blocked
	RateByPath    map[string]int // Different limits for different paths
	RateByMethod  map[string]int // Different limits for HTTP methods
	BlockDuration time.Duration  // How long to block after limit exceeded
	CooldownRate  float64        // Rate at which limits cool down

	// Token bucket specific
	TokenRate  float64 // Rate at which tokens are added
	BucketSize int     // Maximum tokens per bucket

	// Leaky bucket specific
	LeakRate float64 // Rate at which requests leak

	// Adaptive specific
	MinLimit      int     // Minimum rate limit
	MaxLimit      int     // Maximum rate limit
	ScalingFactor float64 // How quickly limits adjust
}

// DefaultRateLimiterConfig returns the default configs for RateLimiter which for now, we are going to stick to IP-based
func DefaultRateLimiterConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:      100,
		Window:     time.Minute,
		BurstLimit: 20,
		CustomKeyFunc: func(c *zen.Context) string {
			return c.GetClientIP()
		},
		StatusCode: http.StatusTooManyRequests,
		Strategy:   IPBased,
	}
}

// windowEntry keeps track of requests in a time window
type windowEntry struct {
	count        int
	startTime    time.Time
	tokens       float64   // for token bucket
	lastUpdate   time.Time // for token bucket/leaky
	blocked      bool      // traking blocked status for IP address
	blockedUntil time.Time // when blocking expires
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
	case TokenBucket:
		return rl.isTokenBucketAllowed(key, now)
	case Leaky:
		return rl.isLeakyBucketAllowed(key, now)
	case Adaptive:
		return rl.isAdaptiveAllowed(key, now)
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

// TokenBucket implementations
func (rl *RateLimiter) isTokenBucketAllowed(key string, now time.Time) bool {
	val, exists := rl.windows.Load(key)
	if !exists {
		rl.windows.Store(key, &windowEntry{
			tokens:     float64(rl.config.BucketSize),
			lastUpdate: now,
		})
		return true
	}

	entry := val.(*windowEntry)

	// Check for blocked entries
	if entry.blocked && now.Before(entry.blockedUntil) {
		return false
	}

	// Calculation for token addition based on time passed
	elapsed := now.Sub(entry.lastUpdate).Seconds()
	newTokens := elapsed * rl.config.TokenRate

	entry.tokens = math.Min(
		float64(rl.config.BucketSize),
		entry.tokens+newTokens,
	)

	entry.lastUpdate = now

	if entry.tokens > 1 {
		entry.tokens--
		return true
	}

	// Check and implement blocking if configured
	if rl.config.BlockDuration > 0 {
		entry.blocked = true
		entry.blockedUntil = now.Add(rl.config.BlockDuration)
	}

	return false
}

func (rl *RateLimiter) isLeakyBucketAllowed(key string, now time.Time) bool {
	val, exists := rl.windows.Load(key)
	if !exists {
		rl.windows.Store(key, &windowEntry{
			count:      1,
			lastUpdate: now,
		})
		return true
	}

	entry := val.(*windowEntry)

	// Calculate leaked requests
	elapsed := now.Sub(entry.lastUpdate).Seconds()
	leaked := int(elapsed * rl.config.LeakRate)
	entry.count = max(0, entry.count-leaked)
	entry.lastUpdate = now

	if entry.count < rl.config.Limit {
		entry.count++
		return true
	}

	return false
}

// Adaptive rate limiting
func (rl *RateLimiter) isAdaptiveAllowed(key string, now time.Time) bool {
	val, exists := rl.windows.Load(key)
	if !exists {
		rl.windows.Store(key, &windowEntry{
			count:     1,
			startTime: now,
		})
		return true
	}

	entry := val.(*windowEntry)
	windowDuration := now.Sub(entry.startTime)

	if windowDuration > rl.config.Window {
		// Adjust limit based on previous window's usage
		currentRate := float64(entry.count) / windowDuration.Seconds()
		newLimit := int(float64(rl.config.Limit) *
			math.Pow(rl.config.ScalingFactor, currentRate/float64(rl.config.Limit)))

		// Ensure limit stays within bounds
		newLimit = max(rl.config.MinLimit, min(rl.config.MaxLimit, newLimit))

		entry.count = 1
		entry.startTime = now
		rl.config.Limit = newLimit
		return true
	}

	if entry.count < rl.config.Limit {
		entry.count++
		return true
	}
	return false
}

// Token Bucket Strategy
// config := middleware.RateLimitConfig{
//     Strategy:    middleware.TokenBucket,
//     TokenRate:   10,    // 10 tokens per second
//     BucketSize:  100,   // Max 100 tokens
//     BlockDuration: time.Minute,
// }

// // Leaky Bucket Strategy
// config := middleware.RateLimitConfig{
//     Strategy:    middleware.Leaky,
//     Limit:       100,
//     LeakRate:    1.0,  // 1 request per second leaks
// }

// // Adaptive Strategy
// config := middleware.RateLimitConfig{
//     Strategy:      middleware.Adaptive,
//     MinLimit:      50,
//     MaxLimit:      200,
//     ScalingFactor: 1.1,
// }

// // Distributed Strategy
// config := middleware.RateLimitConfig{
//     Strategy:    middleware.Distributed,
//     RedisClient: redisClient,
//     KeyPrefix:   "ratelimit:",
//     Limit:       100,
//     Window:      time.Minute,
// }
