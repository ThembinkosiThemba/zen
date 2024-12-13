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

	// using the logger and recovery middleware
	config := middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           3600,
	}

	log.Printf("Initializing middleware...")

	corsMiddleware := middleware.CORSWithConfig(config)
	app.Use(corsMiddleware)
	app.Use(middleware.Recovery())

	// Enable logging to file with default path (logs/zen.log)
	app.Use(middleware.Logger(middleware.LoggerConfig{
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
