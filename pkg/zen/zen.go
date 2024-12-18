package zen

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Engine is the framework instance
type Engine struct {
	*RouterGroup
	router *Router
	groups []*RouterGroup
	addr   string
	ctx    Context
}

type Route struct {
	Method string
	Path   string
}

// New creates a new Engine instance
func New() *Engine {
	engine := &Engine{
		router: NewRouter(),
	}

	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Run starts the HTTP server
func (e *Engine) Serve(addr string) error {
	port, err := checkPort(addr)
	if err != nil {
		Errorf("port check failed: %v", err)
		return fmt.Errorf("port check failed: %v", err)
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

	if IsDevMode() {
		e.printRoutes()
		fmt.Print(e.zenAsciiArt(newAddr))
	}

	return http.ListenAndServe(newAddr, e)
}

// ServeTLS start the HTTPS server
func (e *Engine) ServeTLS(addr, certFile, keyFile string) error {
	port, err := checkPort(addr)
	if err != nil {
		Errorf("port check failed: %v", err)
		return fmt.Errorf("port check failed: %v", err)
	}

	newAddr := fmt.Sprintf("%d", port)
	if strings.Contains(addr, ":") {
		host, _, _ := net.SplitHostPort(addr)
		if host != "" {
			newAddr = fmt.Sprintf("%s:%d", host, port)
		}
	}

	if newAddr != addr {
		Infof("Port %s in in use. Using port %d instead", addr, port)
	}

	e.printRoutes()
	fmt.Print(e.zenAsciiArt(newAddr))
	return http.ListenAndServeTLS(newAddr, certFile, keyFile, e)
}

// ServeWithTimeout starts the HTTP server with timeout settings
func (e *Engine) ServeWithTimeout(addr string, timeout time.Duration) error {
	port, err := checkPort(addr)
	if err != nil {
		Errorf("port check failed: %v", err)
		return fmt.Errorf("port check failed: %v", err)
	}

	newAddr := fmt.Sprintf("%d", port)
	if strings.Contains(addr, ":") {
		host, _, _ := net.SplitHostPort(addr)
		if host != "" {
			newAddr = fmt.Sprintf("%s:%d", host, port)
		}
	}

	if newAddr != addr {
		Infof("Port %s in in use. Using port %d instead", addr, port)
	}

	srv := &http.Server{
		Addr:         newAddr,
		Handler:      e,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout * 2,
	}
	e.printRoutes()
	fmt.Print(e.zenAsciiArt(newAddr))
	return srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (engine *Engine) Shutdown(timeout time.Duration) error {
	server := &http.Server{
		Addr:    engine.addr,
		Handler: engine,
	}

	var ctx, cancel = engine.ctx.DefaultContext()
	defer cancel()

	return server.Shutdown(ctx)

}

// ServeHTTP implements the http.Handler interface
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	e.router.handle(c)
}

// Use adds middleware to the engine's global middleware stack
func (engine *Engine) Use(middlewares ...HandlerFunc) {
	engine.router.Use(middlewares...)
	engine.RouterGroup.Use(middlewares...)
}

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
