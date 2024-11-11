# Security Middleware Documentation

The Security middleware for the Zen framework provides a comprehensive security solution that combines multiple security features into a single, configurable middleware.

## Features

- Security Headers (HSTS, CSP, X-Frame-Options, etc.)
- Request Sanitization (Size limits, file validation, XSS protection)
- IP Security (Allow/blocklisting, geo-blocking)
- Session Security (Timeout, rotation, secure cookies)

## Installation

The security middleware is included in the middleware package:

```go
import "github.com/ThembinkosiThemba/zen/pkg/middleware"
```

## Basic Usage

Use with default configuration:

```go
func main() {
    app := zen.New()
    app.Use(middleware.SecurityMiddleware())
    app.Serve(":8080")
}
```

## Security Strategies

The middleware supports multiple security strategies that can be enabled/disabled:

```go
const (
    HeaderSecurity      // Security headers
    RequestSanitization // Request validation and sanitization
    IPSecurity         // IP-based security
    SessionSecurity    // Session management security
)
```

## Advanced Usage

Custom configuration example:

```go
func main() {
    app := zen.New()

    config := middleware.SecurityConfig{
        Strategies: middleware.HeaderSecurity | 
                   middleware.RequestSanitization,
        
        // Security Headers Configuration
        HSTS: true,
        HSTSMaxAge: 63072000, // 2 years
        CSPDirectives: &middleware.ContentSecurityPolicyDirective{
            DefaultSrc: []string{"'self'"},
            ScriptSrc:  []string{"'self'", "trusted-scripts.com"},
        },
        
        // Request Sanitization
        MaxRequestSize: 5 * 1024 * 1024, // 5MB
        AllowedFileTypes: []string{".pdf", ".docx"},
        SanitizeHTML: true,
        
        // Custom error handling
        CustomErrorHandler: func(c *zen.Context, err error) {
            c.JSON(http.StatusBadRequest, map[string]string{
                "error": err.Error(),
            })
        },
    }

    app.Use(middleware.SecurityMiddleware(config))
    app.Serve(":8080")
}
```

## Configuration Options

### Security Headers Configuration

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| HSTS | bool | Enable HSTS | true |
| HSTSMaxAge | int | HSTS max age in seconds | 31536000 |
| HSTSIncludeSubdomains | bool | Include subdomains in HSTS | true |
| CSPDirectives | *ContentSecurityPolicyDirective | CSP configuration | See defaults |
| FrameOptions | string | X-Frame-Options value | "SAMEORIGIN" |

### Request Sanitization Configuration

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| MaxRequestSize | int64 | Maximum request size | 10MB |
| AllowedFileTypes | []string | Allowed file extensions | [".jpg", ".pdf", ...] |
| SanitizeHTML | bool | Sanitize HTML in requests | true |
| SQLInjectionCheck | bool | Check for SQL injection | true |

### Security Best Practices

1. Always enable HTTPS in production
2. Use strict CSP policies
3. Enable all security checks in production
4. Customize file upload limits based on your needs
5. Implement proper logging for security violations

## Examples

### API Security Configuration

```go
config := middleware.SecurityConfig{
    Strategies: middleware.HeaderSecurity | 
                middleware.RequestSanitization,
    CSPDirectives: &middleware.ContentSecurityPolicyDirective{
        DefaultSrc: []string{"'self'"},
        ConnectSrc: []string{"'self'", "api.yourdomain.com"},
    },
    MaxRequestSize: 1 * 1024 * 1024, // 1MB limit for API requests
}
```

### File Upload Security

```go
config := middleware.SecurityConfig{
    Strategies: middleware.RequestSanitization,
    MaxRequestSize: 50 * 1024 * 1024, // 50MB
    AllowedFileTypes: []string{".jpg", ".png", ".pdf", ".docx"},
}
```

For additional details and updates, visit the [GitHub repository](https://github.com/ThembinkosiThemba/zen).