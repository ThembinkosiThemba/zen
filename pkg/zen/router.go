package zen

import (
	"net/http"
	"strings"
)

// type HandlerFunc defines the handler used by zen
type HandlerFunc func(*Context)

// Router is responsible for routing HTTP requests
type Router struct {
	handlers map[string]map[string][]HandlerFunc
}

// RouterGroup is used for grouping routes
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	engine      *Engine
}

// newRouter creates a new router instance
func newRouter() *Router {
	return &Router{
		handlers: make(map[string]map[string][]HandlerFunc),
	}
}

// Group creates a new router group
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// Use adds middleware to the group
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleware...)
}

// addRoute registers a new router
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	handlers := group.combineHandlers(handler)

	if group.engine.router.handlers[method] == nil {
		group.engine.router.handlers[method] = make(map[string][]HandlerFunc)
	}
	group.engine.router.handlers[method][pattern] = handlers
}

// combineHandlers combines group middleware with route handler
func (group *RouterGroup) combineHandlers(handlers ...HandlerFunc) []HandlerFunc {
	finalSize := len(group.middlewares) + len(handlers)
	mergedHandlers := make([]HandlerFunc, finalSize)
	copy(mergedHandlers, group.middlewares)
	copy(mergedHandlers[len(group.middlewares):], handlers)
	return mergedHandlers
}

// GET registers a new GET route
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST registers a new POST route
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// PUT registers a new PUT route
func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRoute("PUT", pattern, handler)
}

// DELETE registers a new DELETE route
func (group *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	group.addRoute("DELETE", pattern, handler)
}

// PATCH registers a new PATCH route
func (group *RouterGroup) PATCH(pattern string, handler HandlerFunc) {
	group.addRoute("PATCH", pattern, handler)
}

// OPTIONS registers a new OPTIONS route
func (group *RouterGroup) OPTIONS(pattern string, handler HandlerFunc) {
	group.addRoute("OPTIONS", pattern, handler)
}

// HEAD registers a new HEAD route
func (group *RouterGroup) HEAD(pattern string, handler HandlerFunc) {
	group.addRoute("HEAD", pattern, handler)
}

// handle processes the request and finds the matching route
func (r *Router) handle(c *Context) {
	method := c.Request.Method
	path := c.Request.URL.Path

	if methodHandlers := r.handlers[method]; methodHandlers != nil {
		for pattern, handlers := range methodHandlers {
			if params, ok := matchPath(pattern, path); ok {
				c.Params = params
				c.Handlers = handlers
				c.Next()
				return
			}
		}
	}

	// if not found
	c.Text(http.StatusNotFound, "404 NOT FOUND")
}

// matchPath checks if a path matches a pattern
func matchPath(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i := 0; i < len(patternParts); i++ {
		if strings.HasPrefix(patternParts[i], ":") {
			paramName := strings.TrimPrefix(patternParts[i], ":")
			params[paramName] = pathParts[i]
		} else if patternParts[i] != pathParts[i] {
			return nil, false
		}
	}

	return params, true
}
