package zen

import (
	"net/http"
	"path"
	"strings"
)

// RouterGroup groups routes under a shared path prefix and middleware chain.
type RouterGroup struct {
	handlers []HandlerFunc
	basePath string
	engine   *Engine
}

// Apply attaches middleware to the group. It only applies to routes
// registered on the group (or its subgroups) after the call.
func (group *RouterGroup) Apply(middleware ...HandlerFunc) {
	group.handlers = append(group.handlers, middleware...)
}

// Group creates a subgroup with the given path prefix. The subgroup
// inherits the middleware of its parent at the time of the call.
//
//	api := app.Group("/api/v1")
//	api.GET("/users", listUsers) // GET /api/v1/users
func (group *RouterGroup) Group(prefix string, middleware ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		handlers: group.combineHandlers(middleware),
		basePath: joinPaths(group.basePath, prefix),
		engine:   group.engine,
	}
}

// Handle registers handlers for the given method and path. Paths may contain
// named parameters (":id") and a trailing catch-all ("*filepath").
// The last handler is the route handler; any before it are route middleware.
func (group *RouterGroup) Handle(method, relativePath string, handlers ...HandlerFunc) {
	group.engine.addRoute(method, joinPaths(group.basePath, relativePath), group.combineHandlers(handlers))
}

// GET registers a GET route.
func (group *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	group.Handle(http.MethodGet, path, handlers...)
}

// POST registers a POST route.
func (group *RouterGroup) POST(path string, handlers ...HandlerFunc) {
	group.Handle(http.MethodPost, path, handlers...)
}

// PUT registers a PUT route.
func (group *RouterGroup) PUT(path string, handlers ...HandlerFunc) {
	group.Handle(http.MethodPut, path, handlers...)
}

// PATCH registers a PATCH route.
func (group *RouterGroup) PATCH(path string, handlers ...HandlerFunc) {
	group.Handle(http.MethodPatch, path, handlers...)
}

// DELETE registers a DELETE route.
func (group *RouterGroup) DELETE(path string, handlers ...HandlerFunc) {
	group.Handle(http.MethodDelete, path, handlers...)
}

// OPTIONS registers an OPTIONS route.
func (group *RouterGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	group.Handle(http.MethodOptions, path, handlers...)
}

// HEAD registers a HEAD route.
func (group *RouterGroup) HEAD(path string, handlers ...HandlerFunc) {
	group.Handle(http.MethodHead, path, handlers...)
}

// combineHandlers returns a new slice of the group's middleware followed by handlers.
// A fresh slice is always allocated so groups never share a backing array.
func (group *RouterGroup) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	merged := make([]HandlerFunc, 0, len(group.handlers)+len(handlers))
	merged = append(merged, group.handlers...)
	return append(merged, handlers...)
}

func joinPaths(base, relative string) string {
	joined := path.Join(base, relative)
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}
