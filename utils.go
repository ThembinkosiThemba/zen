package zen

import (
	"fmt"
	"strings"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

func (engine *Engine) zenAsciiArt(port string) string {
	return fmt.Sprintf(`
%s    ███████╗███████╗███╗   ██╗
    ╚══███╔╝██╔════╝████╗  ██║
      ███╔╝ █████╗  ██╔██╗ ██║
     ███╔╝  ██╔══╝  ██║╚██╗██║
    ███████╗███████╗██║ ╚████║
    ╚══════╝╚══════╝╚═╝  ╚═══╝%s
    
    %s🎋 Lightweight, Secure & Fast HTTP Framework for Modern Apps%s
    %s⚡ Running on port %s%s
    %s✨ %s%s
    `,
		Cyan, Reset,
		Green, Reset,
		Yellow, port, Reset,
		Purple, time.Now().Format("2006-01-02 15:04:05"), Reset)
}

// getMethodColor returns the color for HTTP methods
func GetMethodColor(method string) string {
	switch method {
	case "GET":
		return Blue
	case "POST":
		return Green
	case "PUT":
		return Yellow
	case "DELETE":
		return Red
	case "PATCH":
		return Cyan
	case "HEAD":
		return Purple
	case "OPTIONS":
		return White
	default:
		return Reset
	}
}

// print routes to terminal
func (engine *Engine) printRoutes() {
	routes := engine.Routes()
	maxPathLength := 0
	methodWidth := 7

	for _, r := range routes {
		if len(r.Path) > maxPathLength {
			maxPathLength = len(r.Path)
		}
	}

	maxPathLength += 2
	if maxPathLength < 30 {
		maxPathLength = 30
	}

	// Print header for routes
	fmt.Printf("%s\nRegistered Routes%s\n", Green, Reset)

	// Print the top border
	fmt.Printf("╔═%s═╦═%s═╗\n",
		strings.Repeat("═", methodWidth),
		strings.Repeat("═", maxPathLength))

	for _, r := range routes {
		methodColor := GetMethodColor(r.Method)
		method := fmt.Sprintf("%-"+fmt.Sprint(methodWidth)+"s", r.Method)
		path := fmt.Sprintf("%-"+fmt.Sprint(maxPathLength)+"s", r.Path)

		fmt.Printf("║ %s%s%s ║ %s ║\n",
			methodColor,
			method,
			Reset,
			path)
	}

	fmt.Printf("╚═%s═╩═%s═╝\n",
		strings.Repeat("═", methodWidth),
		strings.Repeat("═", maxPathLength))
}

// Helper functions for colorizing output
func ColorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return Green
	case code >= 300 && code < 400:
		return Blue
	case code >= 400 && code < 500:
		return Yellow
	default:
		return Red
	}
}
