package main

import (
	"log"
	"net/http"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
	"github.com/ThembinkosiThemba/zen/pkg/zen/middleware"
)

func main() {
	// create new zen instancee
	app := zen.New()
	zen.SetCurrentMode(zen.DevMode)

	app.Use(
		middleware.DefaultCors(),
		middleware.Recovery(),
	)

	// Enable logging to file with default path (logs/zen.log)
	app.Use(zen.Logger(zen.LoggerConfig{
		LogToFile:   true,
		LogFilePath: "logs/zen.log",
	}))

	// define simple routes
	app.GET("/", func(c *zen.Context) {
		c.JSON(http.StatusOK, "Welcome to Zen!")
	})

	app.POST("/users", func(c *zen.Context) {
		c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "User created successfully",
		})
	})

	if err := app.Serve(":8080"); err != nil {
		log.Fatal(err)
	}
}
