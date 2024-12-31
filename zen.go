package zen

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Engine is the core framework instance for managing routing and middleware for the Zen framework.
type Engine struct {
	*RouterGroup                // - RouterGroup: Provides group-based routing and middleware chaining.
	router       *Router        // - router: The main router instance that handles route registration and dispatching.
	groups       []*RouterGroup // - groups: A collection of all RouterGroups associated with the engine.
	addr         string         // - addr: The address where the server is bound (host:port).
	ctx          Context        // - ctx: A default context used for server operations like shutdown.
}

// Route represents an individual HTTP route within the framework.
type Route struct {
	Method string // - Method: The HTTP method (e.g., GET, POST) associated with the route.
	Path   string // - Path: The URL path pattern (e.g., "/users/:id") for the route.
}

// New creates a new Engine instance.
// Initializes routing capabilities and returns a new Engine instance.
func New() *Engine {
	engine := &Engine{
		router: NewRouter(),
	}

	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Serve starts an HTTP server on the given address.
// - addr: The address (host:port) where the server will listen.
// - Returns an error if the server fails to start or if address resolution fails.
func (e *Engine) Serve(addr string) error {
	newAddr, err := resolveAddress(addr)
	if err != nil {
		return err
	}

	if IsDevMode() {
		e.printRoutes()
		fmt.Print(e.zenAsciiArt(newAddr))
	}

	return http.ListenAndServe(newAddr, e)
}

// ServeTLS starts an HTTPS server on the given address using TLS.
// - addr: The address (host:port) where the server will listen.
// - certFile: Path to the TLS certificate file.
// - keyFile: Path to the TLS key file.
// - Returns an error if the server fails to start or if address resolution fails.
func (e *Engine) ServeTLS(addr, certFile, keyFile string) error {
	newAddr, err := resolveAddress(addr)
	if err != nil {
		return err
	}

	if IsDevMode() {
		e.printRoutes()
		fmt.Print(e.zenAsciiArt(newAddr))
	}

	return http.ListenAndServeTLS(newAddr, certFile, keyFile, e)
}

// ServeWithTimeout starts an HTTP server with timeout settings on the given address.
// - addr: The address (host:port) where the server will listen.
// - timeout: Duration for read, write, and idle timeouts.
// - Returns an error if the server fails to start or if address resolution fails.
func (e *Engine) ServeWithTimeout(addr string, timeout time.Duration) error {
	newAddr, err := resolveAddress(addr)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         newAddr,
		Handler:      e,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout * 2,
	}
	if IsDevMode() {
		e.printRoutes()
		fmt.Print(e.zenAsciiArt(newAddr))
	}

	return srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
// - timeout: The maximum duration to wait for existing connections to close.
// - Returns an error if shutdown fails.
func (engine *Engine) Shutdown(timeout time.Duration) error {
	server := &http.Server{
		Addr:    engine.addr,
		Handler: engine,
	}

	var ctx, cancel = engine.ctx.DefaultContext()
	defer cancel()

	return server.Shutdown(ctx)

}

// ServeHTTP implements the http.Handler interface for the Engine.
// - w: The HTTP response writer.
// - req: The HTTP request.
// - Delegates request handling to the router.
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	e.router.handle(c)
}

// Apply adds middleware to the engine's global middleware stack.
// - middlewares: A variadic list of middleware functions to apply.
// - Middleware is applied globally to all routes.
func (engine *Engine) Apply(middlewares ...HandlerFunc) {
	engine.router.Apply(middlewares...)
}

// Routes retrieves all registered routes in the engine.
// - Returns: A slice of Route structs representing HTTP routes (method and path).
func (engine *Engine) Routes() []Route {
	var routes []Route

	for method, paths := range engine.router.handlers {
		for path := range paths {
			routes = append(routes, Route{
				Method: method,
				Path:   path,
			})
		}
	}

	return routes
}

// checkPort validates and resolves a port for the given address.
// - addr: The address (host:port) to check.
// - Returns: A valid, available port and an error if resolution fails.
// - If the port is in use, it increments until a free one is found.
func checkPort(addr string) (int, error) {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %v", err)
	}

	port := 0
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %v", err)
	}

	for {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") {
				port++
				continue
			}
			return 0, err
		}
		listener.Close()
		return port, nil
	}
}

// resolveAddress resolves the final address for the server to bind to.
// - addr: The input address (host:port).
// - Returns: A resolved address string and an error if resolution fails.
// - Logs a warning if the port is already in use and changes the port.
func resolveAddress(addr string) (string, error) {
	port, err := checkPort(addr)
	if err != nil {
		Errorf("port check failed: %v", err)
		return "", fmt.Errorf("port check failed: %v", err)
	}

	newAddr := fmt.Sprintf(":%d", port)
	if strings.Contains(addr, ":") {
		host, _, _ := net.SplitHostPort(addr)
		if host != "" {
			newAddr = fmt.Sprintf("%s:%d", host, port)
		}
	}

	if newAddr != addr {
		Infof("Port %s in in use. Using port %d instead", addr, port)
	}

	return newAddr, nil
}
