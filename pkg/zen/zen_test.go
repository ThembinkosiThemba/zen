package zen

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	engine := New()

	if engine.router == nil {
		t.Error("Router should be initilized")
	}
	if engine.RouterGroup == nil {
		t.Error("RouterGroup should be initialized")
	}
	if len(engine.groups) != 1 {
		t.Error("Groups should be initialized with root group")
	}

	if engine.RouterGroup.prefix != "" {
		t.Error("Root group should have empty prefix")
	}
}

func TestEngine_ServeHTTP(t *testing.T) {
	engine := New()

	tests := []struct {
		name, method, path, handlerPath string
		expectedCode                    int
		expectedBody                    string
	}{
		{
			name:         "Existing route",
			method:       "GET",
			path:         "/test",
			handlerPath:  "/test",
			expectedCode: http.StatusOK,
			expectedBody: "success",
		},
		{
			name:         "Non-existent route",
			method:       "GET",
			path:         "/notfound",
			handlerPath:  "/test",
			expectedCode: http.StatusNotFound,
			expectedBody: "404 NOT FOUND",
		},
		{
			name:         "Wrong method",
			method:       "POST",
			path:         "/test",
			handlerPath:  "/test",
			expectedCode: http.StatusNotFound,
			expectedBody: "404 NOT FOUND",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register test handler
			engine.GET(tt.handlerPath, func(c *Context) {
				c.Text(http.StatusOK, "success")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			engine.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if body := strings.TrimSpace(w.Body.String()); body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestEngine_Use(t *testing.T) {
	engine := New()
	middlewareOrder := []string{}

	engine.Use(func(c *Context) {
		middlewareOrder = append(middlewareOrder, "global1")
		c.Next()
	})
	engine.Use(func(c *Context) {
		middlewareOrder = append(middlewareOrder, "global2")
		c.Next()
	})

	engine.GET("/test", func(c *Context) {
		middlewareOrder = append(middlewareOrder, "handler")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	expected := []string{"global1", "global2", "handler"}
	if len(middlewareOrder) != len(expected) {
		t.Errorf("Expected %d middleware calls, got %d", len(expected), len(middlewareOrder))
	}

	for i, v := range expected {
		if i >= len(middlewareOrder) || middlewareOrder[i] != v {
			t.Errorf("Expected middleware order %v, got %v", expected, middlewareOrder)
			break
		}
	}
}

func TestEngine_Routes(t *testing.T) {
	engine := New()

	// Register various routes
	routes := []Route{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/posts"},
		{"DELETE", "/posts/:id"},
	}

	for _, r := range routes {
		switch r.Method {
		case "GET":
			engine.GET(r.Path, func(c *Context) {})
		case "POST":
			engine.POST(r.Path, func(c *Context) {})
		case "DELETE":
			engine.DELETE(r.Path, func(c *Context) {})
		}
	}

	// Get registered routes
	registeredRoutes := engine.Routes()

	if len(registeredRoutes) != len(routes) {
		t.Errorf("Expected %d routes, got %d", len(routes), len(registeredRoutes))
	}

	// Sort both slices for comparison
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Method != routes[j].Method {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	sort.Slice(registeredRoutes, func(i, j int) bool {
		if registeredRoutes[i].Method != registeredRoutes[j].Method {
			return registeredRoutes[i].Method < registeredRoutes[j].Method
		}
		return registeredRoutes[i].Path < registeredRoutes[j].Path
	})

	// Compare routes
	for i, route := range routes {
		if registeredRoutes[i].Method != route.Method || registeredRoutes[i].Path != route.Path {
			t.Errorf("Route mismatch at index %d: expected %v, got %v",
				i, route, registeredRoutes[i])
		}
	}
}

func TestEngine_GroupRoutes(t *testing.T) {
	engine := New()

	// Create groups with different prefixes
	apiV1 := engine.Group("/api/v1")
	apiV2 := engine.Group("/api/v2")

	// Add routes to groups
	apiV1.GET("/users", func(c *Context) {})
	apiV2.GET("/users", func(c *Context) {})

	routes := engine.Routes()

	expectedPaths := map[string]bool{
		"/api/v1/users": false,
		"/api/v2/users": false,
	}

	if len(routes) != len(expectedPaths) {
		t.Errorf("Expected %d routes, got %d", len(expectedPaths), len(routes))
	}

	for _, route := range routes {
		if route.Method != "GET" {
			t.Errorf("Unexpected method: %s", route.Method)
		}
		if _, exists := expectedPaths[route.Path]; !exists {
			t.Errorf("Unexpected path: %s", route.Path)
		}
		expectedPaths[route.Path] = true
	}

	// Verify all expected paths were found
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected path not found: %s", path)
		}
	}
}

func TestEngine_NestedGroups(t *testing.T) {
	engine := New()

	// Create nested groups
	api := engine.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")

	handlerCalled := false
	users.GET("/:id", func(c *Context) {
		handlerCalled = true
		if id := c.GetParam("id"); id != "123" {
			t.Errorf("Expected id parameter '123', got '%s'", id)
		}
	})

	// Test the nested route
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	engine.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}
