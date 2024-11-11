# JWT Authentication Middleware for Zen Framework

A flexible and customizable JWT authentication middleware for the Zen web framework. This middleware provides secure authentication with support for custom claims, multiple token sources, and configurable unauthorized responses.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration Options](#configuration-options)
- [Token Sources](#token-sources)
- [Custom Claims](#custom-claims)
- [Skip Authentication](#skip-authentication)
- [Error Handling](#error-handling)
- [Complete Examples](#complete-examples)
- [API Reference](#api-reference)

## Installation

```bash
go get github.com/ThembinkosiThemba/zen
```

## Quick Start

Here's a minimal example to get started with the auth middleware:

```go
package main

import (
    "github.com/ThembinkosiThemba/zen/pkg/zen"
    "github.com/ThembinkosiThemba/zen/pkg/middleware"
)

func main() {
    r := zen.New()
    
    // Add auth middleware with a secret key
    r.Use(middleware.Auth("your-secret-key"))
    
    r.GET("/protected", func(c *zen.Context) {
        // Get the default claims
        claims, ok := middleware.GetClaims[*middleware.BaseClaims](c)
        if !ok {
            c.Status(http.StatusUnauthorized)
            return
        }
        
        c.JSON(200, map[string]string{
            "user_id": claims.UserID,
            "role":    claims.Role,
        })
    })
    
    r.Serve(":8080")
}
```

## Configuration Options

The middleware can be configured using `AuthConfig`:

```go
config := middleware.AuthConfig{
    SecretKey:     "your-secret-key",     // Required
    TokenLookup:   "header:Authorization", // Optional (default)
    TokenHeadName: "Bearer",              // Optional (default)
    SkipPaths:     []string{"/public"},   // Optional
    ClaimsFactory: func() jwt.Claims {    // Optional
        return &CustomClaims{}
    },
    Unauthorized: func(c *zen.Context, err error) { // Optional
        c.JSON(401, map[string]string{"error": err.Error()})
    },
}

r.Use(middleware.AuthWithConfig(config))
```

## Token Sources

The middleware supports multiple token sources:

### 1. Header (default)
```go
// Config for Authorization header
config := middleware.AuthConfig{
    TokenLookup: "header:Authorization",
    TokenHeadName: "Bearer",
}

// Token format in request
// Authorization: Bearer <token>
```

### 2. Query Parameter
```go
// Config for query parameter
config := middleware.AuthConfig{
    TokenLookup: "query:token",
}

// URL format
// /api/resource?token=<token>
```

### 3. Cookie
```go
// Config for cookie
config := middleware.AuthConfig{
    TokenLookup: "cookie:jwt",
}

// Cookie name will be "jwt"
```

## Custom Claims

You can extend the default claims with your own fields:

```go
// Define custom claims
type CustomClaims struct {
    UserID      string `json:"user_id"`
    Role        string `json:"role"`
    Name        string `json:"name"`
    Email       string `json:"email"`
    Department  string `json:"department"`
    jwt.RegisteredClaims
}

// Configure middleware with custom claims
config := middleware.AuthConfig{
    SecretKey: "your-secret-key",
    ClaimsFactory: func() jwt.Claims {
        return &CustomClaims{}
    },
}

// Generate token with custom claims
func generateCustomToken(userID, role, name, email, dept string) (string, error) {
    claims := &CustomClaims{
        UserID:      userID,
        Role:        role,
        Name:        name,
        Email:       email,
        Department:  dept,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    return middleware.GenerateToken(claims, "your-secret-key")
}

// Use custom claims in handler
r.GET("/profile", func(c *zen.Context) {
    claims, ok := middleware.GetClaims[*CustomClaims](c)
    if !ok {
        c.Status(http.StatusUnauthorized)
        return
    }
    
    c.JSON(200, map[string]interface{}{
        "user_id": claims.UserID,
        "name":    claims.Name,
        "email":   claims.Email,
        "dept":    claims.Department,
    })
})
```

## Skip Authentication

You can skip authentication for specific paths:

```go
config := middleware.AuthConfig{
    SecretKey: "your-secret-key",
    SkipPaths: []string{
        "/public",
        "/auth/login",
        "/health",
    },
}
```

## Error Handling

Customize unauthorized responses:

```go
config := middleware.AuthConfig{
    SecretKey: "your-secret-key",
    Unauthorized: func(c *zen.Context, err error) {
        switch err {
        case middleware.ErrMissingToken:
            c.JSON(401, map[string]interface{}{
                "error":   "Authentication required",
                "details": "No token provided",
            })
        case middleware.ErrInvalidToken:
            c.JSON(401, map[string]interface{}{
                "error":   "Authentication failed",
                "details": "Invalid or expired token",
            })
        default:
            c.JSON(500, map[string]interface{}{
                "error": "Internal server error",
            })
        }
    },
}
```

## Complete Examples

### 1. Basic Authentication Server

```go
package main

import (
    "github.com/ThembinkosiThemba/zen/pkg/zen"
    "github.com/ThembinkosiThemba/zen/pkg/middleware"
    "net/http"
    "time"
)

func main() {
    r := zen.New()
    
    // Configure auth middleware
    authConfig := middleware.AuthConfig{
        SecretKey: "your-secret-key",
        SkipPaths: []string{"/login"},
    }
    
    // Add middleware
    r.Use(middleware.AuthWithConfig(authConfig))
    
    // Login endpoint (skipped from auth)
    r.POST("/login", func(c *zen.Context) {
        // Simple login logic (replace with your own)
        var login struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        
        if err := c.BindJSON(&login); err != nil {
            c.Status(http.StatusBadRequest)
            return
        }
        
        // Create claims
        claims := middleware.NewBaseClaims(
            login.Username,
            "user",
            24*time.Hour,
        )
        
        // Generate token
        token, err := middleware.GenerateToken(claims, "your-secret-key")
        if err != nil {
            c.Status(http.StatusInternalServerError)
            return
        }
        
        c.JSON(200, map[string]string{
            "token": token,
        })
    })
    
    // Protected endpoint
    r.GET("/protected", func(c *zen.Context) {
        claims, ok := middleware.GetClaims[*middleware.BaseClaims](c)
        if !ok {
            c.Status(http.StatusUnauthorized)
            return
        }
        
        c.JSON(200, map[string]interface{}{
            "message": "Welcome to protected resource",
            "user_id": claims.UserID,
            "role":    claims.Role,
        })
    })
    
    r.Serve(":8080")
}
```

### 2. Advanced Usage with Custom Claims and Role-Based Access

```go
package main

import (
    "github.com/ThembinkosiThemba/zen/pkg/zen"
    "github.com/ThembinkosiThemba/zen/pkg/middleware"
    "github.com/golang-jwt/jwt/v5"
    "net/http"
    "time"
)

// Custom claims with extra fields
type UserClaims struct {
    UserID      string   `json:"user_id"`
    Role        string   `json:"role"`
    Name        string   `json:"name"`
    Email       string   `json:"email"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}

// Role-based middleware
func RequireRole(roles ...string) zen.HandlerFunc {
    return func(c *zen.Context) {
        claims, ok := middleware.GetClaims[*UserClaims](c)
        if !ok {
            c.Status(http.StatusUnauthorized)
            return
        }
        
        for _, role := range roles {
            if claims.Role == role {
                c.Next()
                return
            }
        }
        
        c.JSON(http.StatusForbidden, map[string]string{
            "error": "Insufficient permissions",
        })
    }
}

func main() {
    r := zen.New()
    
    // Configure auth with custom claims
    config := middleware.AuthConfig{
        SecretKey: "your-secret-key",
        SkipPaths: []string{"/login"},
        ClaimsFactory: func() jwt.Claims {
            return &UserClaims{}
        },
    }
    
    r.Use(middleware.AuthWithConfig(config))
    
    // Admin-only endpoint
    r.GET("/admin", RequireRole("admin"), func(c *zen.Context) {
        claims, _ := middleware.GetClaims[*UserClaims](c)
        c.JSON(200, map[string]interface{}{
            "message": "Welcome to admin panel",
            "user":    claims.Name,
            "email":   claims.Email,
        })
    })
    
    // User or admin endpoint
    r.GET("/dashboard", RequireRole("user", "admin"), func(c *zen.Context) {
        claims, _ := middleware.GetClaims[*UserClaims](c)
        c.JSON(200, map[string]interface{}{
            "message": "Welcome to dashboard",
            "user":    claims.Name,
        })
    })
    
    r.Serve(":8080")
}
```

## API Reference

### Types

```go
type AuthConfig struct {
    SecretKey     string
    TokenLookup   string
    TokenHeadName string
    SkipPaths     []string
    Unauthorized  func(*zen.Context, error)
    ClaimsFactory func() jwt.Claims
}

type BaseClaims struct {
    UserID string
    Role   string
    jwt.RegisteredClaims
}
```

### Functions

```go
// Create middleware with default config
func Auth(secretKey string) zen.HandlerFunc

// Create middleware with custom config
func AuthWithConfig(config AuthConfig) zen.HandlerFunc

// Generate new token
func GenerateToken(claims jwt.Claims, secretKey string) (string, error)

// Get claims from context
func GetClaims[T jwt.Claims](c *zen.Context) (T, bool)

// Create new base claims
func NewBaseClaims(userID, role string, expiry time.Duration) *BaseClaims
```

### Error Types

```go
var (
    ErrMissingToken = errors.New("missing authorization token")
    ErrInvalidToken = errors.New("invalid authorization token")
)
```

For more information or support, please visit the [GitHub repository](https://github.com/ThembinkosiThemba/zen) or open an issue.