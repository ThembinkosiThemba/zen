# Zen Response Helpers Documentation

## Quick Start

Zen provides three main response helpers for consistent API responses:
- `zen.M` - Quick map creation
- `zen.R` - Standard data responses
- `zen.Err` - Error responses

## Basic Usage

### Using zen.M (Map)
```go
app.GET("/", func(c *zen.Context) {
    // Method 1: Direct creation
    c.JSON(200, zen.M{
        "message": "Welcome",
        "version": "1.0",
    })

    // Method 2: Using variadic function
    c.JSON(200, zen.M(
        "message", "Welcome",
        "version", "1.0",
    ))
})
```

### Using zen.R (Response)
```go
app.GET("/users", func(c *zen.Context) {
    users := []string{"John", "Jane"}
    c.JSON(200, zen.R(users))
})

// Response:
{
    "data": ["John", "Jane"],
    "success": true
}
```

### Using zen.Err (ApiError)
```go
app.POST("/users", func(c *zen.Context) {
    // Simple error
    c.JSON(400, zen.Err("Invalid request", 400))

    // Error with details
    c.JSON(400, zen.ErrWithDetails("Validation failed", 400, zen.M{
        "email": "invalid format",
        "name": "required",
    }))
})
```

## Complete Example

```go
package main

import (
    "log"
    "net/http"

    "github.com/ThembinkosiThemba/zen/pkg/zen"
    "github.com/ThembinkosiThemba/zen/pkg/zen/middleware"
)

func main() {
    app := zen.New()

    app.Use(
        middleware.Logger(),
        middleware.RateLimiter(),
        middleware.Recovery(),
    )

    // Basic response using Map
    app.GET("/status", func(c *zen.Context) {
        c.JSON(200, zen.M{
            "status": "healthy",
            "uptime": "2h",
        })
    })

    // Data response with metadata
    app.GET("/users", func(c *zen.Context) {
        users := []zen.M{
            {"id": 1, "name": "John"},
            {"id": 2, "name": "Jane"},
        }
        
        response := zen.R(users).WithMeta(zen.M{
            "total": 2,
            "page": 1,
        })
        
        c.JSON(200, response)
    })

    // Error handling
    app.GET("/users/:id", func(c *zen.Context) {
        id := c.ParamInt("id")
        
        if id <= 0 {
            c.JSON(400, zen.Err("Invalid ID", 400))
            return
        }

        // Not found error with details
        c.JSON(404, zen.ErrWithDetails("User not found", 404, zen.M{
            "id": id,
            "type": "user",
        }))
    })

    if err := app.Serve(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

## Response Examples

### Success with zen.M
```json
{
    "status": "healthy",
    "uptime": "2h"
}
```

### Data with zen.R
```json
{
    "data": [
        {"id": 1, "name": "John"},
        {"id": 2, "name": "Jane"}
    ],
    "success": true,
    "meta": {
        "total": 2,
        "page": 1
    }
}
```

### Error with zen.Err
```json
{
    "message": "Validation failed",
    "code": 400,
    "details": {
        "email": "invalid format",
        "name": "required"
    }
}
```

## Key Features

1. **Type Safety**: All helpers are properly typed with Go structs
2. **Chainable Methods**: Support for method chaining (WithMeta, WithDetails)
3. **Consistent Format**: Standardized response formats across your API
4. **Helper Methods**: Additional utility methods for type conversion and data access
5. **Clean Syntax**: More readable than raw map declarations

## Best Practices

1. Use `zen.M` for simple key-value responses
2. Use `zen.R` for returning data with success status
3. Use `zen.Err` for error responses with proper HTTP status codes
4. Include relevant details in error responses
5. Use consistent response formats across your API
6. Add metadata for paginated responses