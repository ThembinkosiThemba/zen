package zen_test

import (
	"net/http/httptest"
	"testing"

	"github.com/ThembinkosiThemba/zen"
	"github.com/ThembinkosiThemba/zen/middleware"
)

func benchmark(b *testing.B, app *zen.Engine, method, path string, headers ...string) {
	req := httptest.NewRequest(method, path, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	b.ReportAllocs()
	for b.Loop() {
		app.ServeHTTP(w, req)
	}
}

func BenchmarkStaticRoute(b *testing.B) {
	app := zen.New()
	app.GET("/test", func(c *zen.Context) { c.Text(200, "hello") })
	benchmark(b, app, "GET", "/test")
}

func BenchmarkParamRoute(b *testing.B) {
	app := zen.New()
	app.GET("/users/:id/posts/:post_id", func(c *zen.Context) {
		c.JSON(200, zen.M{"id": c.GetParam("id"), "post_id": c.GetParam("post_id")})
	})
	benchmark(b, app, "GET", "/users/123/posts/456")
}

func BenchmarkManyRoutes(b *testing.B) {
	app := zen.New()
	for _, r := range []string{"/a", "/b", "/c/:id", "/d/e/f", "/users", "/users/:id", "/users/:id/posts"} {
		app.GET(r, func(c *zen.Context) {})
	}
	benchmark(b, app, "GET", "/users/42/posts")
}

func BenchmarkCORSPreflight(b *testing.B) {
	app := zen.New()
	app.Apply(middleware.CORS())
	app.GET("/test", func(c *zen.Context) { c.Text(200, "hello") })
	benchmark(b, app, "OPTIONS", "/test", "Origin", "http://localhost:3000", "Access-Control-Request-Method", "GET")
}
