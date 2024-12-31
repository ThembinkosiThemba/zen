// This package is responsible for setting the mode for the zen framework.
// If you do not set the mode when running your application, by default, you will be running
// on DevMode. For production releases, you can switch to release mode.
// Realease mode will remove all useful development functions for a much cleaner
// out on your console/terminal
package zen

import (
	"fmt"
)

// Mode represents the running mode of the Zen framework
type Mode int

const (
	// DevMode enables detailed logging and debugging features
	DevMode Mode = iota

	// Production disables features for better performance and organization
	Production
)

var currentMode = DevMode

// SetCurrentMode sets the current mode that zen will be running on
func SetCurrentMode(m Mode) {
	currentMode = m
	if currentMode == DevMode {
		Debug("Running in development mode - switch to Production for deployment")
		Info("You can change the mode by: zen.SetCurrentMode(zen.Production)")
		fmt.Println()
	}
}
// Helper function used to get the mode for zen
func GetMode() Mode {
	return currentMode
}

func IsDevMode() bool {
	return currentMode == DevMode
}
