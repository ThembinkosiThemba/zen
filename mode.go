package zen

import (
	"fmt"
	"os"
)

// Mode is the running mode of the framework.
type Mode int

const (
	// DevMode prints registered routes and startup information. It is the default.
	DevMode Mode = iota
	// Production disables debug output.
	Production
)

var currentMode = DevMode

// SetCurrentMode sets the running mode. Call it before creating routes.
func SetCurrentMode(m Mode) { currentMode = m }

// GetMode returns the running mode.
func GetMode() Mode { return currentMode }

// IsDevMode reports whether zen is running in DevMode.
func IsDevMode() bool { return currentMode == DevMode }

func debugPrint(format string, values ...any) {
	if IsDevMode() {
		fmt.Fprintf(os.Stderr, "[ZEN-debug] "+format+"\n", values...)
	}
}

func (engine *Engine) debugPrintStartup(addr string) {
	if !IsDevMode() {
		return
	}
	debugPrint("Running in DevMode. Switch to production with zen.SetCurrentMode(zen.Production)")
	for _, r := range engine.routes {
		debugPrint("%-7s %s", r.Method, r.Path)
	}
	debugPrint("Listening on %s", addr)
}
