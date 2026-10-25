package zen

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LoggerConfig configures the Logger middleware.
type LoggerConfig struct {
	// Output is where log lines are written. Defaults to os.Stdout.
	// To also log to a file, use io.MultiWriter(os.Stdout, file).
	Output io.Writer

	// SkipPaths are request paths that are not logged, such as health checks.
	SkipPaths []string
}

// Logger returns middleware that logs each request to os.Stdout.
func Logger() HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}

// LoggerWithConfig returns request logging middleware with the given config.
// Colors are used only when writing to a terminal.
func LoggerWithConfig(cfg LoggerConfig) HandlerFunc {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}
	color := isTerminal(out)

	skip := make(map[string]struct{}, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skip[p] = struct{}{}
	}

	var mu sync.Mutex
	return func(c *Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if _, ok := skip[path]; ok {
			return
		}
		if raw != "" {
			path += "?" + raw
		}

		status, method := c.Writer.Status(), c.Request.Method
		statusStr, methodStr := fmt.Sprintf(" %3d ", status), fmt.Sprintf(" %-7s", method)
		if color {
			statusStr = statusColor(status) + statusStr + reset
			methodStr = methodColor(method) + methodStr + reset
		}

		mu.Lock()
		defer mu.Unlock()
		fmt.Fprintf(out, "[ZEN] %s |%s| %13v | %15s |%s %q\n",
			start.Format("2006/01/02 - 15:04:05"),
			statusStr,
			time.Since(start),
			c.GetClientIP(),
			methodStr,
			path,
		)
	}
}

const (
	reset   = "\033[0m"
	red     = "\033[97;41m"
	green   = "\033[97;42m"
	yellow  = "\033[90;43m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	white   = "\033[90;47m"
)

func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func methodColor(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	default:
		return white
	}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
