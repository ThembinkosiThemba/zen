I've added the advanced rate limiting example with a custom configuration to the README usage guide:

---

# Rate Limiter Middleware Documentation

The Rate Limiter middleware for the Zen framework provides IP-based rate limiting to control API usage, prevent abuse, and ensure fair usage.

## Features

- Configurable request limits and time windows
- IP-based tracking by default, with customizable key functions (e.g., by API key or user ID)
- Optional burst handling for brief over-limit requests
- Route-specific limits and exclusions
- Customizable error responses

## Installation

The rate limiter is included in the middleware package. Import it as follows:

```go
import "github.com/ThembinkosiThemba/zen/pkg/middleware"
```

## Basic Usage

Here’s how to use the rate limiter with default settings:

```go
func main() {
    app := zen.New()
    app.Use(middleware.RateLimiter())
    app.Run(":8080")
}
```

The default configuration allows:
- 100 requests per minute
- 20-request burst capacity
- IP-based tracking

## Advanced Usage with Custom Configuration

Here’s an example of advanced rate limiting with custom configurations:

```go
func main() {
    app := zen.New()

    // Advanced rate limiting with custom configuration
    app.Use(middleware.RateLimiter(middleware.RateLimitConfig{
        Limit:      1000,           // Base limit
        Window:     time.Hour,      // Time window
        BurstLimit: 50,             // Allow bursts up to 50 extra requests
        ExlcudePaths: []string{     // Paths to exclude
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
    }))

    app.Run(":8080")
}
```

This configuration:
- Limits requests to 1000 per hour, with up to 50 additional burst requests
- Excludes specific paths (`/health` and `/metrics`) from rate limiting
- Uses `X-API-Key` as the unique identifier instead of IP
- Provides a custom JSON error response with a `retry_after` field

## Route-Specific Rate Limiting

Apply rate limits to specific routes or groups for fine-grained control:

```go
func main() {
    app := zen.New()
    
    authConfig := middleware.RateLimitConfig{
        Limit:  5,
        Window: time.Minute,
    }

    auth := app.Group("/auth")
    auth.Use(middleware.RateLimiter(authConfig))
    
    app.Use(middleware.RateLimiter())
    app.Run(":8080")
}
```

## Configuration Options

The `RateLimitConfig` struct provides the following configuration options:

| Option           | Type                      | Description                                      | Default                 |
|------------------|---------------------------|--------------------------------------------------|-------------------------|
| `Limit`          | `int`                     | Maximum requests allowed within the window       | `100`                   |
| `Window`         | `time.Duration`           | Time window for rate limiting                    | `1 minute`              |
| `BurstLimit`     | `int`                     | Extra requests allowed in a burst                | `20`                    |
| `CustomKeyFunc`  | `func(*zen.Context) string` | Custom function for rate limit keys             | Client IP function      |
| `ExlcudePaths`   | `[]string`                | Paths to exclude from rate limiting              | `nil`                   |
| `StatusCode`     | `int`                     | HTTP status for rate limit errors                | `429 (Too Many Requests)`|
| `CustomErrorFunc`| `func(*zen.Context, time.Duration)` | Custom error function | `nil`                   |

For additional details, visit the [GitHub repository](https://github.com/ThembinkosiThemba/zen).