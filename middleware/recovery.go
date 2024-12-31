package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/ThembinkosiThemba/zen"
)

// Recovery middleware recovers from panics and sends a 500 error
func Recovery() zen.HandlerFunc {
	return func(z *zen.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("%sPANIC RECOVERED%s\n%s\n%s%s\n",
					zen.Red, zen.Reset,
					err,
					string(debug.Stack()),
					zen.Reset)

				z.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error": "Internal Server Error",
				})
			}
		}()
		z.Next()
	}
}
