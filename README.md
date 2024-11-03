# Zen Web Framework

Zen is a lightweight and fast HTTP framework for Go, focusing on simplicity and performance while providing essential features for modern web applications.

<p align="start">
    <img src="./docs/assets/zen.png" alt="zen" />
</p>

## Features

- 🚀 Lightweight and Fast
- 🛡️ Built-in Middleware Support
- 🎯 Simple Routing
- 🔒 Authentication Support
- 🌐 CORS Handling
- ⚡  Rate Limiting
- 📝 Request Logging
- 🔄 Hot Reloading

## Quick Start

### Installation

```bash
go get github.com/ThembinkosiThemba/zen
```

### Basic Example

```go
package main

import (
    "github.com/ThembinkosiThemba/zen/pkg/zen"
    "github.com/ThembinkosiThemba/zen/pkg/middleware"
)

func main() {
    // Create new Zen app
    app := zen.New()

    // Global middleware
    app.Use(
        middleware.Logger(),      // Zen logger
        middleware.RateLimiter(), // Rate limiting of APIS
        middleware.DefaultCors(), // Default CORS configuration
    )
    
    // Basic routes
    app.GET("/", func(c *zen.Context) {
        c.JSON(http.StatusOK, map[string]interface{}{
            "message": "Welcome to Zen!",
        })
    })

    // Start server
    app.Run(":8080")
}
```

### Complete Example with All Features

```go
package main

import (
    "github.com/ThembinkosiThemba/zen/pkg/zen"
    "github.com/ThembinkosiThemba/zen/pkg/middleware"
    "time"
)

func main() {
    // Create new Zen app
    app := zen.New()

    // Configure middleware
    corsConfig := middleware.CORSConfig{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:          3600,
    }

    rateConfig := middleware.RateLimitConfig{
        Limit:  100,
        Window: time.Minute,
    }

    // Apply middleware
    app.Use(middleware.Logger())                    // Logging
    app.Use(middleware.CORSWithConfig(corsConfig))  // CORS
    app.Use(middleware.RateLimiter(rateConfig))     // Rate Limiting

    // Public routes
    app.GET("/", func(c *zen.Context) {
        c.JSON(http.StatusOK, map[string]interface{}{
            "message": "Welcome to Zen!",
        })
    })

    // Protected routes group
    api := app.Group("/api")
    api.Use(middleware.Auth("your-secret-key")) // Authentication middleware

    api.GET("/users", func(c *zen.Context) {
        // Example of binding JSON request
        var req struct {
            Page  int `json:"page"`
            Limit int `json:"limit"`
        }

        if !c.BindJSONWithError(&req) {
            return
        }

        c.JSON(http.StatusOK, map[string]interface{}{
            "users": []string{"user1", "user2"},
        })
    })

    api.POST("/users", func(c *zen.Context) {
        var user struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }

        if !c.BindJSONWithError(&user) {
            return
        }

        c.JSON(http.StatusCreated, user)
    })

    // Start server with hot reload
    zen.HotReloadEnabled = true // Enable hot reloading
    app.Run(":8080")
}
```

## Middleware Documentation

Detailed documentation for each middleware:

- [Authentication](docs/middleware/auth.md)
- [CORS](docs/middleware/cors.md)
- [Rate Limiter](docs/middleware/rate_limiter.md)
- [Logger](docs/middleware/logger.md)

## Configuration

### Hot Reload

Enable hot reloading during development:

```go
zen.HotReloadEnabled = true
```

### Custom Middleware Stack

Create custom middleware combinations:

```go
func CustomMiddlewareStack() []zen.HandlerFunc {
    return []zen.HandlerFunc{
        middleware.Logger(),
        middleware.DefaultCors(),
        middleware.RateLimiter(),
    }
}
```

## Best Practices

1. **Middleware Order**:
   - Logger (first to log all requests)
   - CORS (early to handle preflight)
   - Rate Limiter
   - Authentication (after rate limiting)

2. **Error Handling**:
   ```go
   app.Use(func(c *zen.Context) {
       defer func() {
           if err := recover(); err != nil {
               c.JSON(http.StatusInternalServerError, map[string]interface{}{
                   "error": "Internal Server Error",
               })
           }
       }()
       c.Next()
   })
   ```

3. **Route Groups**:
   ```go
   v1 := app.Group("/v1")
   v1.Use(middleware.Auth("secret"))
   
   admin := v1.Group("/admin")
   admin.Use(adminAuthMiddleware)
   ```

## Performance Tips

1. Use appropriate rate limits
2. Enable hot reload only in development
3. Configure CORS specifically for your needs
4. Use route groups for better organization

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details