package main

import (
	"log"
	"net/http"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
	"github.com/ThembinkosiThemba/zen/pkg/zen/middleware"
)

func main() {

	// enable host reload
	zen.HotReloadEnabled = false

	// create new zen instancee
	app := zen.New()

	// using the logger and recovery middleware
	app.Use(
		middleware.Logger(),
		middleware.RateLimiter(),
		middleware.Recovery(),
	)

	// define simple routes
	app.GET("/", func(c *zen.Context) {
		c.Text(http.StatusOK, "Welcome to Zen!")
	})

	app.POST("/users", func(c *zen.Context) {
		c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "User created successfully",
		})
	})

	if err := app.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
