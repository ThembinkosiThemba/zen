package zen

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func perform(h http.Handler, method, path string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func text(body string) HandlerFunc {
	return func(c *Context) { c.Text(http.StatusOK, "%s", body) }
}

func TestRouting(t *testing.T) {
	app := New()
	app.GET("/", text("root"))
	app.GET("/users", text("list"))
	app.GET("/users/new", text("new"))
	app.GET("/users/:id", func(c *Context) { c.Text(200, "user %s", c.GetParam("id")) })
	app.GET("/users/:id/posts/:post", func(c *Context) {
		c.Text(200, "%s/%s", c.GetParam("id"), c.GetParam("post"))
	})
	app.GET("/static/*filepath", func(c *Context) { c.Text(200, "file %s", c.GetParam("filepath")) })
	app.POST("/users", text("create"))

	tests := []struct {
		method, path string
		code         int
		body         string
	}{
		{"GET", "/", 200, "root"},
		{"GET", "/users", 200, "list"},
		{"GET", "/users/", 200, "list"},
		{"GET", "/users/new", 200, "new"}, // static wins over param
		{"GET", "/users/42", 200, "user 42"},
		{"GET", "/users/report.pdf", 200, "user report.pdf"},
		{"GET", "/users/42/posts/7", 200, "42/7"},
		{"GET", "/users/new/posts/7", 200, "new/7"}, // backtracks from the static branch
		{"GET", "/static/css/app.css", 200, "file /css/app.css"},
		{"GET", "/static/", 200, "file /"},
		{"POST", "/users", 200, "create"},
		{"GET", "/users/42/posts", 404, "404 NOT FOUND"},
		{"GET", "/missing", 404, "404 NOT FOUND"},
		{"DELETE", "/users", 404, "404 NOT FOUND"},
	}
	for _, tt := range tests {
		w := perform(app, tt.method, tt.path)
		if w.Code != tt.code || w.Body.String() != tt.body {
			t.Errorf("%s %s = %d %q, want %d %q", tt.method, tt.path, w.Code, w.Body.String(), tt.code, tt.body)
		}
	}
}

func TestRoutingPanics(t *testing.T) {
	tests := map[string]func(app *Engine){
		"duplicate":          func(app *Engine) { app.GET("/a", text("")); app.GET("/a", text("")) },
		"param name clash":   func(app *Engine) { app.GET("/u/:id", text("")); app.GET("/u/:name/x", text("")) },
		"catch-all not last": func(app *Engine) { app.GET("/a/*rest/b", text("")) },
		"unnamed param":      func(app *Engine) { app.GET("/a/:", text("")) },
		"no handlers":        func(app *Engine) { app.GET("/a") },
	}
	for name, register := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected panic")
				}
			}()
			register(New())
		})
	}
}

func TestMiddlewareOrder(t *testing.T) {
	var order []string
	mw := func(name string) HandlerFunc {
		return func(c *Context) {
			order = append(order, name)
			c.Next()
			order = append(order, "/"+name)
		}
	}

	app := New()
	app.Apply(mw("global"))
	api := app.Group("/api", mw("api"))
	v1 := api.Group("/v1")
	v1.Apply(mw("v1"))
	v1.GET("/x", mw("route"), func(c *Context) { order = append(order, "handler") })

	perform(app, "GET", "/api/v1/x")
	want := []string{"global", "api", "v1", "route", "handler", "/route", "/v1", "/api", "/global"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("order = %v, want %v", order, want)
	}
}

func TestGroupsDoNotShareMiddleware(t *testing.T) {
	app := New()
	api := app.Group("/api")
	a := api.Group("/a")
	b := api.Group("/b")
	a.Apply(func(c *Context) { c.SetHeader("X-Group", "a") })
	b.Apply(func(c *Context) { c.SetHeader("X-Group", "b") })
	a.GET("/x", text("a"))
	b.GET("/x", text("b"))

	if got := perform(app, "GET", "/api/a/x").Header().Get("X-Group"); got != "a" {
		t.Errorf("group a header = %q", got)
	}
	if got := perform(app, "GET", "/api/b/x").Header().Get("X-Group"); got != "b" {
		t.Errorf("group b header = %q", got)
	}
}

func TestNotFoundRunsGlobalMiddleware(t *testing.T) {
	app := New()
	app.Apply(func(c *Context) { c.SetHeader("X-Global", "1") })

	w := perform(app, "GET", "/missing")
	if w.Code != 404 || w.Header().Get("X-Global") != "1" {
		t.Errorf("got %d, X-Global=%q", w.Code, w.Header().Get("X-Global"))
	}

	app.NoRoute(func(c *Context) { c.JSON(404, M{"error": "not found"}) })
	w = perform(app, "GET", "/missing")
	if w.Body.String() != `{"error":"not found"}` || w.Header().Get("X-Global") != "1" {
		t.Errorf("custom NoRoute: %q, X-Global=%q", w.Body.String(), w.Header().Get("X-Global"))
	}
}

func TestQuitStopsChain(t *testing.T) {
	app := New()
	app.Apply(func(c *Context) { c.QuitWithStatus(http.StatusForbidden) })
	app.GET("/", func(c *Context) { t.Error("handler should not run") })

	if w := perform(app, "GET", "/"); w.Code != http.StatusForbidden || w.Body.Len() != 0 {
		t.Errorf("got %d %q", w.Code, w.Body.String())
	}
}

func TestRoutes(t *testing.T) {
	app := New()
	app.GET("/a", text(""))
	app.Group("/api").POST("/b/:id", text(""))

	want := []Route{{"GET", "/a"}, {"POST", "/api/b/:id"}}
	if got := app.Routes(); !reflect.DeepEqual(got, want) {
		t.Errorf("Routes() = %v, want %v", got, want)
	}
}

func TestJoinPaths(t *testing.T) {
	tests := [][3]string{
		{"/", "/users", "/users"},
		{"/", "", "/"},
		{"/api", "/", "/api"},
		{"/api/", "users/", "/api/users"},
		{"/", "api", "/api"},
	}
	for _, tt := range tests {
		if got := joinPaths(tt[0], tt[1]); got != tt[2] {
			t.Errorf("joinPaths(%q, %q) = %q, want %q", tt[0], tt[1], got, tt[2])
		}
	}
}

func TestSetTrustedProxies(t *testing.T) {
	app := New()
	if err := app.SetTrustedProxies("10.0.0.0/8", "192.168.1.1", "::1"); err != nil {
		t.Fatal(err)
	}
	if err := app.SetTrustedProxies("not-an-ip"); err == nil || !strings.Contains(err.Error(), "not-an-ip") {
		t.Errorf("expected error for invalid proxy, got %v", err)
	}
}
