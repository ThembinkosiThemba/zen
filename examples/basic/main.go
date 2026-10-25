package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThembinkosiThemba/zen"
	"github.com/ThembinkosiThemba/zen/middleware"
)

func main() {
	app := zen.Default() // Logger + Recovery
	app.Apply(middleware.CORS(), middleware.Secure())

	app.GET("/", func(c *zen.Context) {
		c.JSON(http.StatusOK, zen.M{"message": "Welcome to Zen!"})
	})

	api := app.Group("/api", middleware.RateLimit())
	api.GET("/users/:id", func(c *zen.Context) {
		c.Success(http.StatusOK, zen.M{"id": c.GetParam("id")}, "user retrieved")
	})

	// Graceful shutdown: use your own http.Server with the engine as the handler.
	srv := &http.Server{Addr: ":8080", Handler: app, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
