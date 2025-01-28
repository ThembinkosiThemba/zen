package zen

import (
	"net/http"
	"regexp"
	"strings"
)

// type HandlerFunc defines the function signature for HTTP request handlers in the Zen Framework.
// It recieves a pointer to Context which contains the request and response information.
type HandlerFunc func(*Context)

// Router is the main routing component responsible for HTTP request routing and middleware management.
// It maintains a mapping of HTTP methods and paths to their corresponding handlers and middleware.
type Router struct {
	// handlers stores route patterns and their handler chains indexed by HTTP method
	handlers map[string]map[string][]HandlerFunc
	// globalMiddleware stores middleware that applies to all routes.
	globalMiddleware []HandlerFunc

	validationConfig ValidationConfig
}

// ValidationConfig holds configuration for path parameter validation
type ValidationConfig struct {
	maxLenght         int
	allowedPattern    *regexp.Regexp
	dangerousPatterns []string
}

// RouterGroup represents a logical grouping of routes with shared prefix and middleware.
// It enables modular organization of routes and middleware scoping.
type RouterGroup struct {
	prefix      string        // prefix is the URL prefix for all routes in this router group
	middlewares []HandlerFunc // middleware stores middleware specific to this router group
	engine      *Engine       // engine points to the main Engine instance for the zen framework
}

// NewRouter initializes and returns a new Router instance with empty handler maps
// and middleware slices.
func NewRouter() *Router {
	defaultValidation := ValidationConfig{
		maxLenght:      256,
		allowedPattern: regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`),
		dangerousPatterns: []string{
			"../", "..\\", // Path traversal
			"<", ">", // HTML/XML injection
			";",       // Command injection
			"'", "\"", // SQL injection
			"\x00",     // Null byte
			"\n", "\r", // CRLF injection
		},
	}
	return &Router{
		handlers:         make(map[string]map[string][]HandlerFunc),
		globalMiddleware: make([]HandlerFunc, 0, 10), // keeping the middleware that can be applied to 10
		validationConfig: defaultValidation,
	}
}

// Apply[Router] applies middleware functions to the global middleware stack.
// These middlewares will be executed for all routes in the application.
// Middleware functions are executed in the order they are added.
func (r *Router) Apply(middleware ...HandlerFunc) {
	r.globalMiddleware = append(r.globalMiddleware, middleware...)
}

// Apply[RouterGroup] applies middleware functions to the current RouterGroup.
// These middlewares will only be executed for routes defined in this group
// and its subgroups.
func (group *RouterGroup) Apply(middleware ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleware...)
}

// GroupRoutes creates a new RouterGroup with the given URL prefix.
// The new group inherits middleware from its parent group.
// Groups can be nested to create hierarchical route structures.
//
// Example:
//
//	api := router.Group("/api/v1")
//	api.GET("/users", GetUsers)  // matches /api/v1/users
func (group *RouterGroup) GroupRoutes(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix:      group.prefix + prefix,
		engine:      engine,
		middlewares: make([]HandlerFunc, 0, len(group.middlewares)),
	}
	copy(newGroup.middlewares, group.middlewares)
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// addRoute registers a new route with the given HTTP method, path pattern, and handler.
// It combines global middleware, group middleware, and the route handler into a single
// handler chain.
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	handlers := group.combineHandlers(handler)

	if group.engine.router.handlers[method] == nil {
		group.engine.router.handlers[method] = make(map[string][]HandlerFunc)
	}
	group.engine.router.handlers[method][pattern] = handlers
}

// combineHandlers merges global middleware, group middleware, and route handlers
// into a single slice while maintaining the correct execution order.
func (group *RouterGroup) combineHandlers(handlers ...HandlerFunc) []HandlerFunc {
	// Calculate final size including global middleware
	finalSize := len(group.engine.router.globalMiddleware) + len(group.middlewares) + len(handlers)
	mergedHandlers := make([]HandlerFunc, finalSize)

	// Copy global middleware first
	n := copy(mergedHandlers, group.engine.router.globalMiddleware)
	// Copy group middleware next
	n += copy(mergedHandlers[n:], group.middlewares)
	// Copy route handlers last
	copy(mergedHandlers[n:], handlers)

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

func (r *Router) writeNotFound(c *Context) {
	c.Status(http.StatusNotFound)
	c.Writer.Write([]byte("404 NOT FOUND"))
	// c.Quit()
}

// handle processes incoming HTTP requests by matching the request path
// to registered routes and executing the corresponding handler chain.
func (r *Router) handle(c *Context) {
	method := c.GetMethod()
	path := c.GetURLPath()

	if method == http.MethodOptions {
		r.handleOptions(c)
		return
	}

	if methodHandlers := r.handlers[method]; methodHandlers != nil {
		for pattern, handlers := range methodHandlers {
			if params, ok := r.matchPath(pattern, path); ok {
				c.Params = params
				c.Handlers = handlers
				c.Next()
				return
			}
		}
	}

	// If no route matches, we will still execute global middleware
	if len(r.globalMiddleware) > 0 {
		c.Handlers = r.globalMiddleware
		r.writeNotFound(c)
		c.Next()
	} else {
		r.writeNotFound(c)
	}
}

// matchPath determines if a request path matches a route pattern and extracts
// any URL parameters. It supports dynamic path segments with ":" prefix.
//
// Example:
//
//	pattern: "/users/:id"
//	path:    "/users/123"
//	result:  params["id"] = "123"
func (r *Router) matchPath(pattern, path string) (map[string]string, bool) {
	splitTrim := func(c rune) bool { return c == '/' || c == ' ' }
	patternParts := strings.FieldsFunc(pattern, splitTrim)
	pathParts := strings.FieldsFunc(path, splitTrim)

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i := 0; i < len(patternParts); i++ {
		if strings.HasPrefix(patternParts[i], ":") {
			paramName := strings.TrimPrefix(patternParts[i], ":")
			paramValue := pathParts[i]

			if !r.validatePathParam(paramValue) {
				return nil, false
			}

			params[paramName] = pathParts[i]
		} else if patternParts[i] != pathParts[i] {
			return nil, false
		}
	}

	return params, true
}

func (r *Router) handleOptions(c *Context) {
	methodHandlers := r.handlers[http.MethodOptions]
	for pattern, handlers := range methodHandlers {
		if params, ok := r.matchPath(pattern, c.Request.URL.Path); ok {
			c.Params = params
			c.Handlers = append(r.globalMiddleware, handlers...)
			c.Next()
			return
		}
	}

	// If no specific handler is found, we execute global middleware like Logger
	if len(r.globalMiddleware) > 0 {
		c.Handlers = r.globalMiddleware
		c.Next()
		// 204 error code is set if no response is written by middleware
		if c.Writer.Status() == 0 {
			c.Text(http.StatusNoContent, "")
		}
		return
	}

	c.JSON(http.StatusNoContent, "")
}

func (r *Router) validatePathParam(value string) bool {
	if value == "" {
		return false
	}

	if len(value) > r.validationConfig.maxLenght {
		return false
	}

	for _, pattern := range r.validationConfig.dangerousPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}

	return r.validationConfig.allowedPattern.MatchString(value)
}
