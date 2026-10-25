package main

import (
	"net/http"

	"github.com/ThembinkosiThemba/zen"
)

func main() {
	app := zen.Default() // Logger and Recovery middleware

	app.GET("/hello/:name", func(c *zen.Context) {
		c.JSON(http.StatusOK, zen.M{"message": "Hello, " + c.GetParam("name")})
	})

	files := http.StripPrefix("/static", http.FileServer(http.Dir("./public")))
	app.GET("/static/*filepath", func(c *zen.Context) {
		files.ServeHTTP(c.Writer, c.Request)
	})

	app.Serve("127.0.0.1:18080")
}
