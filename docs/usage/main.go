package main

import (
	"net/http"

	"github.com/ThembinkosiThemba/zen"
	"github.com/ThembinkosiThemba/zen/middleware"
)

func main() {
	// create new zen instancee
	app := zen.New()
	zen.SetCurrentMode(zen.DevMode)

	app.Apply(
		middleware.DefaultCors(),
		middleware.Recovery(),
	)

	// Enable logging to file with default path (logs/zen.log)
	app.Apply(zen.Logger(zen.LoggerConfig{
		LogToFile:   true,
		LogFilePath: "logs/zen.log",
	}))

	// define simple routes
	app.GET("/", func(c *zen.Context) {
		c.JSON(http.StatusOK, "Welcome to Zen!")
	})

	app.GET("/users", func(c *zen.Context) {
		user := zen.M{"name": "Thembinkosi", "surname": "Mkhonta"}
		c.Success(http.StatusOK, user, "User retrieved successfully")
	})

	if err := app.Serve(":8080"); err != nil {
		zen.Fatal(err)
	}
}
