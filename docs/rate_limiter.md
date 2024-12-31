# Rate Limiter Middleware Documentation

The Rate Limiter middleware for the Zen framework provides flexible rate limiting strategies to control API usage, prevent abuse, and ensure fair usage.

## Table of Contents

- [Features](#Features)
- [Basic Usage](#basic-usage)
- [Rate Limiting Strategies](#rate-limiting-strategies)
  - [IP-Based Strategy](#ip-based-strategy)
  - [Sliding Window Strategy](#sliding-window-strategy)
- [Advanced Usage With Custom Configuration](#advanced-usage-with-custom-configuration)
- [Route Specific Ratelimiting](#route-specific-rate-limiting)
- [Configuration Options](#configuration-options)

## Features

- Multiple rate limiting strategies (IP-based and Sliding Window)
- Configurable request limits and time windows
- IP-based tracking by default, with customizable key functions
- Optional burst handling for brief over-limit requests
- Route-specific limits and exclusions
- Customizable error responses
- Thread-safe implementation

The rate limiter is included in the middleware package. Import it as follows:

## Basic Usage

Here's how to use the rate limiter with default settings:

```go
import "github.com/ThembinkosiThemba/zen/middleware"

func main() {
    app := zen.New()
    app.Use(middleware.RateLimiterMiddleware())
    app.Serve(":8080")
}
```

The default configuration allows:

- `100 requests` per minute
- `20-request` burst capacity
- `IP-based` rate limiting strategy

## Rate Limiting Strategies

### IP-Based Strategy

Simple time-window based limiting per IP address. When the window expires, the counter resets completely.

```go
func main() {
    app := zen.New()

    config := middleware.RateLimitConfig{
        Strategy: middleware.IPBased,
        Limit:    100,
        Window:   time.Minute,
    }

    app.Use(middleware.RateLimiterMiddleware(config))
    app.Serve(":8080")
}
```

Implementation details are as follows:

- Uses a simple time window approach
- When a window expires, the counter resets completely
- Good for simple use cases and when exact precision isn't required

### Sliding Window Strategy

More accurate rate limiting that considers the overlap between windows, providing smoother rate limiting behavior.

Implementation details are as follows:

- More sophisticated approach that considers window overlaps
- Provides smoother rate limiting by weighing requests from the previous window
- Better for scenarios requiring more precise rate limiting
- Helps prevent sudden traffic spikes at window boundaries

```go
func main() {
    app := zen.New()

    config := middleware.RateLimitConfig{
        Strategy: middleware.SlidingWindow,
        Limit:    100,
        Window:   time.Minute,
    }

    app.Use(middleware.RateLimiterMiddleware(config))
    app.Serve(":8080")
}
```

## Advanced Usage with Custom Configuration

Here's an example of advanced rate limiting with custom configurations:

```go
func main() {
    app := zen.New()

    config := middleware.RateLimitConfig{
        Strategy:   middleware.SlidingWindow,  // Use sliding window strategy
        Limit:      1000,                      // Base limit
        Window:     time.Hour,                 // Time window
        BurstLimit: 50,                        // Allow bursts up to 50 extra requests
        ExcludePaths: []string{                // Paths to exclude
            "/health",
            "/metrics",
        },
        CustomKeyFunc: func(c *zen.Context) string {
            // Use API key for rate limiting instead of IP
            return c.Request.Header.Get("X-API-Key")
        },
        CustomErrorFunc: func(c *zen.Context, window time.Duration) {
            // Custom error response
            c.JSON(429, map[string]interface{}{
                "status": "error",
                "message": "Too many requests",
                "retry_after": window.Seconds(),
            })
        },
    }

    app.Use(middleware.RateLimiterMiddleware(config))
    app.Serve(":8080")
}
```

## Route-Specific Rate Limiting

Apply different rate limiting strategies and configurations to specific routes:

```go
func main() {
    app := zen.New()

    // Strict rate limiting for authentication endpoints
    authConfig := middleware.RateLimitConfig{
        Strategy: middleware.SlidingWindow,
        Limit:    5,
        Window:   time.Minute,
    }

    // More lenient rate limiting for API endpoints
    apiConfig := middleware.RateLimitConfig{
        Strategy: middleware.IPBased,
        Limit:    1000,
        Window:   time.Hour,
    }

    // Apply different strategies to different routes
    auth := app.Group("/auth")
    auth.Use(middleware.RateLimiterMiddleware(authConfig))

    api := app.Group("/api")
    api.Use(middleware.RateLimiterMiddleware(apiConfig))

    app.Serve(":8080")
}
```

## Configuration Options

The `RateLimitConfig` struct provides the following configuration options:

| Option            | Type                                | Description                                | Default                   |
| ----------------- | ----------------------------------- | ------------------------------------------ | ------------------------- |
| `Strategy`        | `RateLimitStrategy`                 | Rate limiting strategy to use              | `IPBased`                 |
| `Limit`           | `int`                               | Maximum requests allowed within the window | `100`                     |
| `Window`          | `time.Duration`                     | Time window for rate limiting              | `1 minute`                |
| `BurstLimit`      | `int`                               | Extra requests allowed in a burst          | `20`                      |
| `CustomKeyFunc`   | `func(*zen.Context) string`         | Custom function for rate limit keys        | Client IP function        |
| `ExcludePaths`    | `[]string`                          | Paths to exclude from rate limiting        | `nil`                     |
| `StatusCode`      | `int`                               | HTTP status for rate limit errors          | `429 (Too Many Requests)` |
| `CustomErrorFunc` | `func(*zen.Context, time.Duration)` | Custom error function                      | `nil`                     |

For additional details and updates, visit the [GitHub repository](https://github.com/ThembinkosiThemba/zen).
