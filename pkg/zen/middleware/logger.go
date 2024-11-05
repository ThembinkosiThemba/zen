package middleware

import (
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
		statusColor := zen.ColorForStatus(c.Writer.Status())
		methodColor := zen.GetMethodColor(c.Request.Method)

		log.Printf("%s %3d %s| %13v | %15s | %-7s %s %s\n",
			statusColor, c.Writer.Status(), reset,
			latency,
			c.ClientIP(),
			methodColor+c.Request.Method+reset,
			gray, path,
		)
	}
}
