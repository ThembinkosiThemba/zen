# Logger Middleware Documentation

The Logger middleware for Zen framework provides request logging capabilities with customizable formatting and colored output.

## Features

- Request timing
- Status code coloring
- Method coloring
- Path logging
- IP address logging
- Latency measurement
- Custom formatting
- Path skipping

## Basic Usage

```go
func main() {
    app := zen.New()
    
    // Use default logger
    app.Use(middleware.Logger())
    
    app.Run(":8080")
}
```

## Custom Configuration

```go
func main() {
    app := zen.New()
    
    config := &middleware.LoggerConfig{
        SkipPaths: []string{"/health", "/metrics"},
        Formatter: func(c *zen.Context, latency time.Duration) string {
            return fmt.Sprintf("Custom format: %s %s %v",
                c.Request.Method,
                c.Request.URL.Path,
                latency,
            )
        },
    }
    
    app.Use(middleware.Logger(config))
    
    app.Run(":8080")
}
```

## Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| SkipPaths | []string | Paths to skip logging |
| Formatter | func(*Context, time.Duration) string | Custom log format function |

## Log Output Format

Default log format includes:
- Timestamp
- Status code (colored)
- Latency
- Client IP
- HTTP method (colored)
- Request path

Example output:
```
[ZEN] 2024/01/02 - 15:04:05 | 200 | 13.45ms | 192.168.1.1 | GET /api/users
```

## Custom Formatting Example

```go
config := &middleware.LoggerConfig{
    Formatter: func(c *zen.Context, latency time.Duration) string {
        return fmt.Sprintf("[API] %s - %s - %v",
            c.ClientIP(),
            c.Request.URL.Path,
            latency,
        )
    },
}
app.Use(middleware.Logger(config))
```