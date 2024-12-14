# Logger Middleware Documentation

The Logger middleware for Zen framework provides request logging capabilities with customizable formatting, colored output, and file logging support.

## Features

- Request timing
- Status code coloring
- Method coloring
- Path logging
- IP address logging
- Latency measurement
- Custom formatting
- Path skipping
- File logging with configurable path

## Basic Usage

```go
func main() {
    app := zen.New()

    // Use default logger (console only)
    app.Use(zen.Logger())

    // Use logger with file logging enabled
    app.Use(zen.Logger(zen.LoggerConfig{
        LogToFile: true,  // Logs will be written to logs/zen.log
    }))

    app.Serve(":8080")
}
```

## Custom Configuration

```go
func main() {
    app := zen.New()

    config := zen.LoggerConfig{
        SkipPaths: []string{"/health", "/metrics"},
        LogToFile: true,
        LogFilePath: "custom/path/api.log",
        Formatter: func(c *zen.Context, latency time.Duration) string {
            return fmt.Sprintf("Custom format: %s %s %v",
                c.Request.Method,
                c.Request.URL.Path,
                latency,
            )
        },
    }

    app.Use(zen.Logger(config))

    app.Serve(":8080")
}
```

## Configuration Options

| Option      | Type                                  | Description                 | Default        |
| ----------- | ------------------------------------- | --------------------------- | -------------- |
| SkipPaths   | []string                              | Paths to skip logging       | []             |
| Formatter   | func(\*Context, time.Duration) string | Custom log format function  | nil            |
| LogToFile   | bool                                  | Enable/disable file logging | false          |
| LogFilePath | string                                | Path to log file            | "logs/zen.log" |

## Log Output Format

Default log format includes:

- Timestamp
- Status code (colored in console)
- Latency
- Client IP
- HTTP method (colored in console)
- Request path

Console output example:

```
[ZEN] 2024/01/02 - 15:04:05 | 200 | 13.45ms | 192.168.1.1 | GET /api/users
```

File output example:

```
2024/01/02 15:04:05 200 | 13.45ms | 192.168.1.1 | GET /api/users
```

## Custom Formatting Example

```go
config := zen.LoggerConfig{
    LogToFile: true,
    LogFilePath: "api.log",
    Formatter: func(c *zen.Context, latency time.Duration) string {
        return fmt.Sprintf("[API] %s - %s - %v",
            c.ClientIP(),
            c.Request.URL.Path,
            latency,
        )
    },
}
app.Use(zen.Logger(config))
```
