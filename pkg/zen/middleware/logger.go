package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// logger configuration for the zen framework
type LoggerConfig struct {
	// skip logging for specific paths
	SkipPaths []string
	// Custom log format function
	Formatter func(*zen.Context, time.Duration) string
}

// Logger middleware logs the incoming HTTP request details
func Logger() zen.HandlerFunc {
	return func(c *zen.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// Get status code color
		statusColor := colorForStatus(c.Writer.Status())
		methodColor := colorForMethod(c.Request.Method)

		log.Printf("%s %3d %s| %13v | %15s | %-7s %s %s\n",
			statusColor, c.Writer.Status(), reset,
			latency,
			c.ClientIP(),
			methodColor+c.Request.Method+reset,
			gray, path,
		)
	}
}



func Logger2(config ...*LoggerConfig) zen.HandlerFunc {
	conf := &LoggerConfig{
		Formatter: defaultLogFormat, // default config for the formatter
	}

	if len(config) > 0 && config[0] != nil {
		conf = config[0]
	}

	return func(c *zen.Context) {
		start := time.Now()

		// processing the next request
		c.Next()

		path := c.Request.URL.Path
		for _, skipPath := range conf.SkipPaths {
			if skipPath == path {
				return
			}
		}

		// logging request details
		latency := time.Since(start)
		log.Println(conf.Formatter(c, latency))
	}
}

func defaultLogFormat(c *zen.Context, latency time.Duration) string {
	return fmt.Sprintf("[ZEN] %v | %s | %13v | %15s | %-7s %s",
		time.Now().Format("2006/01/02 - 15:04:05"),
		statusCodeColor(c.Writer.Status()),
		latency,
		c.ClientIP(),
		c.Request.Method,
		c.Request.URL.Path,
	)
}

// statusCodeColor returns colored status code
func statusCodeColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return fmt.Sprintf("\033[32m%d\033[0m", code) // Green
	case code >= 300 && code < 400:
		return fmt.Sprintf("\033[36m%d\033[0m", code) // Cyan
	case code >= 400 && code < 500:
		return fmt.Sprintf("\033[33m%d\033[0m", code) // Yellow
	default:
		return fmt.Sprintf("\033[31m%d\033[0m", code) // Red
	}
}

// Helper functions for colorizing output
func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return blue
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return green
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return cyan
	case "HEAD":
		return purple
	case "OPTIONS":
		return white
	default:
		return reset
	}
}
