# Rate Limiter Middleware Documentation

The Rate Limiter middleware for Zen framework provides IP-based rate limiting capabilities to protect your API endpoints from abuse and ensure fair usage.

## Features

- IP-based rate limiting
- Configurable request limits and time windows
- Automatic cleanup of expired records
- Thread-safe implementation
- Customizable rate limit configuration

## Installation

The rate limiter is included in the middleware package. Import it in your project:

```go
import "github.com/ThembinkosiThemba/zen/pkg/middleware"
```

## Basic Usage

Here's a simple example of how to use the rate limiter with default settings:

```go
func main() {
    app := zen.New()
    
    // Apply rate limiter middleware globally
    app.Use(middleware.RateLimiter())
    
    // ... your routes ...
    
    app.Run(":8080")
}
```

The default configuration allows:
- 100 requests per minute
- Per IP address tracking
- Automatic cleanup of expired records

## Custom Configuration

You can customize the rate limiter behavior by providing a custom configuration:

```go
func main() {
    app := zen.New()
    
    // Custom rate limit configuration
    config := middleware.RateLimitConfig{
        Limit:  50,              // 50 requests
        Window: time.Minute * 5, // per 5 minutes
    }
    
    // Apply rate limiter middleware with custom config
    app.Use(middleware.RateLimiter(config))
    
    app.Run(":8080")
}
```

## Route-Specific Rate Limiting

You can apply different rate limits to specific routes or groups:

```go
func main() {
    app := zen.New()
    
    // Strict rate limit for authentication endpoints
    authConfig := middleware.RateLimitConfig{
        Limit:  5,
        Window: time.Minute,
    }
    
    // Apply to specific routes
    auth := app.Group("/auth")
    auth.Use(middleware.RateLimiter(authConfig))
    
    // Regular rate limit for other endpoints
    app.Use(middleware.RateLimiter())
    
    app.Run(":8080")
}
```

## Configuration Options

The `RateLimitConfig` struct provides the following configuration options:

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| Limit | int | Maximum number of requests allowed within the window | 100 |
| Window | time.Duration | Time window for rate limiting | 1 minute |

## Response Format

When rate limit is exceeded, the middleware returns a 429 (Too Many Requests) status code with a JSON response:

```json
{
    "error": "Rate limit exceeded. Try again in 1m0s"
}
```

## Best Practices

1. **Choose Appropriate Limits**: Set rate limits based on your API's capacity and expected usage patterns.

2. **Different Limits for Different Routes**: Apply stricter limits to sensitive endpoints (like authentication) and more lenient limits to public endpoints.

3. **Monitor Rate Limiting**: Keep track of how often your rate limits are being hit to adjust them if needed.

4. **Inform Users**: Consider adding rate limit information in response headers:
   - X-RateLimit-Limit
   - X-RateLimit-Remaining
   - X-RateLimit-Reset

## Memory Considerations

The rate limiter stores visitor information in memory. The cleanup routine runs periodically to remove expired records, but consider your server's memory capacity when setting up rate limits for high-traffic applications.

## Error Handling

The rate limiter automatically handles error responses when limits are exceeded. You don't need to add any additional error handling in your route handlers.

## Thread Safety

The implementation is thread-safe and can be used in concurrent applications without additional synchronization.

## Limitations

1. In-memory storage means rate limit data is not shared across multiple server instances
2. Rate limits are reset when the server restarts
3. IP-based limiting might not be accurate behind a proxy unless properly configured

## Future Improvements

Consider implementing these features for more robust rate limiting:

1. TempDB-backed storage for distributed systems
2. Custom response headers for rate limit information
3. Dynamic rate limiting based on user roles
4. Token bucket algorithm implementation
5. Configurable response messages

## Example Implementation

Here's a complete example showing various rate limiting configurations:

```go
package main

import (
    "time"
    "github.com/ThembinkosiThemba/zen/pkg/zen"
    "github.com/ThembinkosiThemba/zen/pkg/middleware"
)

func main() {
    app := zen.New()

    // Global rate limit
    app.Use(middleware.RateLimiter())

    // Strict rate limit for auth endpoints
    authConfig := middleware.RateLimitConfig{
        Limit:  5,
        Window: time.Minute,
    }

    // API endpoints with custom rate limits
    auth := app.Group("/auth")
    auth.Use(middleware.RateLimiter(authConfig))
    
    auth.POST("/login", func(c *zen.Context) {
        // Login handler
    })

    // Public endpoints with default rate limit
    app.GET("/public", func(c *zen.Context) {
        // Public handler
    })

    app.Run(":8080")
}
```

For more information or support, please visit the [GitHub repository](https://github.com/ThembinkosiThemba/zen) or open an issue.