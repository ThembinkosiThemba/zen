package zen

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_Basic(t *testing.T) {
	router := NewRouter()
	if router.handlers == nil {
		t.Error("router handlers should be initialised")
	}
}

func TestRouterGroup_AddRoute(t *testing.T) {
	engine := New()
	group := engine.GroupRoutes("/api")

	handlerCalled := false
	handler := func(c *Context) {
		handlerCalled = true
	}

	group.GET("/test", handler)

	if len(engine.router.handlers["GET"]) != 1 {
		t.Error("Router not properly registered")
	}

	path := "/api/test"
	if _, ok := engine.router.handlers["GET"][path]; !ok {
		t.Errorf("Expected route %s not found", path)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)
	engine.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestRouterGroup_Middleware(t *testing.T) {
	engine := New()
	group := engine.GroupRoutes("/api")

	middlewareCalled := false
	middleware := func(c *Context) {
		middlewareCalled = true
		c.Next()
	}

	handlerCalled := false
	handler := func(c *Context) {
		handlerCalled = true
	}

	group.Use(middleware)
	group.GET("/test", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)
	engine.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Middleware not called")
	}
	if !handlerCalled {
		t.Error("Handler not called")
	}
}

func TestRouter_HandleMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			engine := New()
			handlerCalled := false

			handler := func(c *Context) {
				handlerCalled = true
				if c.Request.Method != method {
					t.Errorf("Expected method %s, got %s", method, c.Request.Method)
				}
			}

			// Register the route using the appropriate method
			group := engine.GroupRoutes("")
			switch method {
			case "GET":
				group.GET("/test", handler)
			case "POST":
				group.POST("/test", handler)
			case "PUT":
				group.PUT("/test", handler)
			case "DELETE":
				group.DELETE("/test", handler)
			case "PATCH":
				group.PATCH("/test", handler)
			case "OPTIONS":
				group.OPTIONS("/test", handler)
			case "HEAD":
				group.HEAD("/test", handler)
			}

			// Test the route
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/test", nil)
			engine.ServeHTTP(w, req)

			if !handlerCalled {
				t.Errorf("%s handler not called", method)
			}
		})
	}
}

func TestRouter_404Handler(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent", nil)

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	expectedBody := "404 NOT FOUND"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}
func TestRouter_Handle(t *testing.T) {
	engine := New()
	group := engine.GroupRoutes("/api")

	handlerCalled := false
	handler := func(c *Context) {
		handlerCalled = true
	}

	group.GET("/test", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)
	engine.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestRouter_HandleOptions(t *testing.T) {
	engine := New()
	group := engine.GroupRoutes("/api")

	handlerCalled := false
	handler := func(c *Context) {
		handlerCalled = true
	}

	group.OPTIONS("/test", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	engine.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestRouterGroup_Use(t *testing.T) {
	engine := New()
	group := engine.GroupRoutes("/api")

	middlewareCalled := false
	middleware := func(c *Context) {
		middlewareCalled = true
		c.Next()
	}

	handlerCalled := false
	handler := func(c *Context) {
		handlerCalled = true
	}

	group.Use(middleware)
	group.GET("/test", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)
	engine.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestRouterGroup_GroupRoutes(t *testing.T) {
	engine := New()
	apiGroup := engine.GroupRoutes("/api")
	v1Group := apiGroup.GroupRoutes("/v1")

	handlerCalled := false
	handler := func(c *Context) {
		handlerCalled = true
	}

	v1Group.GET("/test", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	engine.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		path       string
		wantMatch  bool
		wantParams map[string]string
	}{
		{
			name:       "Exact match",
			pattern:    "/users",
			path:       "/users",
			wantMatch:  true,
			wantParams: map[string]string{},
		},
		{
			name:       "Parameter match",
			pattern:    "/users/:id",
			path:       "/users/123",
			wantMatch:  true,
			wantParams: map[string]string{"id": "123"},
		},
		{
			name:       "Multiple parameters",
			pattern:    "/users/:id/posts/:postId",
			path:       "/users/123/posts/456",
			wantMatch:  true,
			wantParams: map[string]string{"id": "123", "postId": "456"},
		},
		{
			name:       "No match",
			pattern:    "/users/:id",
			path:       "/posts/123",
			wantMatch:  false,
			wantParams: nil,
		},
		{
			name:       "Different lengths",
			pattern:    "/users/:id",
			path:       "/users/123/extra",
			wantMatch:  false,
			wantParams: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, match := matchPath(tt.pattern, tt.path)

			if match != tt.wantMatch {
				t.Errorf("matchPath() match = %v, want %v", match, tt.wantMatch)
			}

			if tt.wantMatch {
				if len(params) != len(tt.wantParams) {
					t.Errorf("matchPath() params = %v, want %v", params, tt.wantParams)
				}
				for k, v := range tt.wantParams {
					if params[k] != v {
						t.Errorf("matchPath() param[%s] = %v, want %v", k, params[k], v)
					}
				}
			}
		})
	}
}
