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

	// Print header
	fmt.Printf("\n%s%s%s\n", Green, "Registered Routes", Reset)

	// Print horizontal line with box characters
	fmt.Printf("╔═════════╦%s╗\n", strings.Repeat("═", 30))

	// Print routes
	for _, r := range routes {
		methodColor := GetMethodColor(r.Method)

		// Print each route with proper alignment
		fmt.Printf("║ %s%-7s%s ║ %-28s║\n",
			methodColor,
			r.Method,
			Reset,
			r.Path)
	}

	// Print bottom border
	fmt.Printf("╚═════════╩%s╝\n", strings.Repeat("═", 30))
}

// func containsString(slice []string, item string) bool {
// 	for _, s := range slice {
// 		if s == item {
// 			return true
// 		}
// 	}
// 	return false
// }

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
