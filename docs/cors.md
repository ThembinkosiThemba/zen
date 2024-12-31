# CORS Middleware Documentation

The CORS (Cross-Origin Resource Sharing) middleware for Zen framework provides a flexible way to handle cross-origin requests with customizable configurations.

## Table of Contents

- [Features](#Features)
- [Basic Usage](#basic-usage)
- [Configuration Options](#configuration-options)
- [Custom Configuration](#custom-configuration)
- [Security Considerations](#security-considerations)

## Features

- Configurable allowed origins
- Customizable HTTP methods
- Custom headers support
- Credentials handling
- Preflight request support
- Configurable max age for preflight responses

## Basic Usage

This simple uses default CORS configurations.

```go
func main() {
    app := zen.New()

    // Apply default CORS middleware
    app.Apply(middleware.DefaultCors())

    app.Serve(":8080")
}
```

## Configuration Options

| Option           | Type     | Description                 | Default                                                      |
| ---------------- | -------- | --------------------------- | ------------------------------------------------------------ |
| AllowOrigins     | []string | List of allowed origins     | ["*"]                                                        |
| AllowMethods     | []string | Allowed HTTP methods        | ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"] |
| AllowHeaders     | []string | Allowed HTTP headers        | []                                                           |
| AllowCredentials | bool     | Allow credentials           | false                                                        |
| ExposeHeaders    | []string | Headers that can be exposed | []                                                           |
| MaxAge           | int      | Preflight cache duration    | 0                                                            |

## Custom Configuration

You can use a more fine-grained control measure with CORS and set up your custom configurations:

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

    app.Apply(middleware.CORSWithConfig(config))

    app.Serve(":8080")
}
```

## Security Considerations

1. Avoid using `"*"` for `AllowOrigins` in production
2. Be specific with allowed methods and headers
3. Enable `AllowCredentials` only when necessary
4. Set appropriate `MaxAge` to reduce preflight requests

For additional details and updates, visit the [GitHub repository](https://github.com/ThembinkosiThemba/zen).
