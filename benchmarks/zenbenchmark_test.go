package zen_test

import (
	"net/http/httptest"
	"testing"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
	"github.com/ThembinkosiThemba/zen/pkg/zen/middleware"
)

func BenchmarkBasicRouting(b *testing.B) {
	engine := zen.New()
	engine.GET("/test", func(c *zen.Context) {
		c.Text(200, "hello")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		engine.ServeHTTP(w, req)
	}
}

func BenchmarkParameterizedRouting(b *testing.B) {
	engine := zen.New()
	engine.GET("/users/:id/posts/:post_id", func(c *zen.Context) {
		id := c.GetParam("id")
		postID := c.GetParam("post_id")
		c.JSON(200, map[string]string{"id": id, "post_id": postID})
	})

	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		engine.ServeHTTP(w, req)
	}
}

func BenchmarkCORSMiddleware(b *testing.B) {
	engine := zen.New()
	engine.Use(middleware.DefaultCors())
	engine.GET("/test", func(c *zen.Context) {
		c.Text(200, "hello")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		engine.ServeHTTP(w, req)
	}
}
