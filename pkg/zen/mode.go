// This package is responsible for setting the mode for the zen framework.
// If you do not set the mode when running your application, by default, you will be running
// on DevMode. For production releases, you can switch to release mode.
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
		Warn("Running in development mode - switch to Production for deployment")
		Info("You can change the mode by: zen.SetCurrentMode(zen.Production)")
		fmt.Println()
	}
}

func GetMode() Mode {
	return currentMode
}

func IsDevMode() bool {
	return currentMode == DevMode
}

// TODO: find usage for Production mode checker function
// func isProdMode() bool {
// 	return currentMode == Production
// }
