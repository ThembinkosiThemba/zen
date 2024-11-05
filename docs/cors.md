# CORS Middleware Documentation

The CORS (Cross-Origin Resource Sharing) middleware for Zen framework provides a flexible way to handle cross-origin requests with customizable configurations.

## Features

- Configurable allowed origins
- Customizable HTTP methods
- Custom headers support
- Credentials handling
- Preflight request support
- Configurable max age for preflight responses

## Basic Usage

```go
func main() {
    app := zen.New()
    
    // Use default CORS middleware
    app.Use(middleware.DefaultCors())
    
    app.Run(":8080")
}
```

## Custom Configuration

```go
func main() {
    app := zen.New()
    
    config := middleware.CORSConfig{
        AllowOrigins:     []string{"https://example.com", "https://api.example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true,
        ExposeHeaders:    []string{"Content-Length"},
        MaxAge:          3600,
    }
    
    app.Use(middleware.CORSWithConfig(config))
    
    app.Run(":8080")
}
```

## Configuration Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| AllowOrigins | []string | List of allowed origins | ["*"] |
| AllowMethods | []string | Allowed HTTP methods | ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"] |
| AllowHeaders | []string | Allowed HTTP headers | [] |
| AllowCredentials | bool | Allow credentials | false |
| ExposeHeaders | []string | Headers that can be exposed | [] |
| MaxAge | int | Preflight cache duration | 0 |

## Security Considerations

1. Avoid using `"*"` for `AllowOrigins` in production
2. Be specific with allowed methods and headers
3. Enable `AllowCredentials` only when necessary
4. Set appropriate `MaxAge` to reduce preflight requests

## Examples

### Specific Origins

```go
config := middleware.CORSConfig{
    AllowOrigins: []string{
        "https://app.example.com",
        "https://admin.example.com",
    },
}
app.Use(middleware.CORSWithConfig(config))
```

### API Configuration

```go
config := middleware.CORSConfig{
    AllowOrigins: []string{"https://api.example.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{
        "Origin",
        "Content-Type",
        "Authorization",
        "X-API-Key",
    },
    MaxAge: 3600,
}
```