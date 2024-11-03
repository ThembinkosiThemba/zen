package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// Colors for terminal output
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
	white  = "\033[97m"
)

// Recovery middleware recovers from panics and sends a 500 error
func Recovery() zen.HandlerFunc {
	return func(c *zen.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("%sPANIC RECOVERED%s\n%s\n%s%s\n",
					red, reset,
					err,
					string(debug.Stack()),
					reset)

				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}
