package zen

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
)

// Recovery returns middleware that recovers from panics, logs the stack trace
// to os.Stderr, and responds with 500 Internal Server Error if nothing has
// been written yet.
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			// net/http uses this sentinel to abort a response silently.
			if err == http.ErrAbortHandler {
				panic(err)
			}
			fmt.Fprintf(os.Stderr, "[ZEN] panic recovered: %v\n%s\n", err, debug.Stack())
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, M{"error": http.StatusText(http.StatusInternalServerError)})
			}
			c.Quit()
		}()
		c.Next()
	}
}
