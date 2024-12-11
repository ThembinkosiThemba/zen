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
	router    *Router
	groups    []*RouterGroup
	addr      string
	cus_contx Context
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
func (engine *Engine) Serve(addr string) error {
	port, err := checkPort(addr)
	if err != nil {
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
		fmt.Printf("Port %s was in use. Using port %d instead.\n", addr, port)
	}

	engine.printRoutes()
	fmt.Print(engine.zenAsciiArt(newAddr))
	return http.ListenAndServe(newAddr, engine)
}

// ServeTLS start the HTTPS server
func (engine *Engine) ServeTLS(addr, certFile, keyFile string) error {
	port, err := checkPort(addr)
	if err != nil {
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
		fmt.Printf("Port %s was in use. Using port %d instead.\n", addr, port)
	}

	engine.printRoutes()
	fmt.Print(engine.zenAsciiArt(newAddr))
	return http.ListenAndServeTLS(newAddr, certFile, keyFile, engine)
}

// ServeWithTimeout starts the HTTP server with timeout settings
func (engine *Engine) ServeWithTimeout(addr string, timeout time.Duration) error {
	port, err := checkPort(addr)
	if err != nil {
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
		fmt.Printf("Port %s was in use. Using port %d instead.\n", addr, port)
	}

	srv := &http.Server{
		Addr:         newAddr,
		Handler:      engine,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout * 2,
	}
	engine.printRoutes()
	fmt.Print(engine.zenAsciiArt(newAddr))
	return srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (engine *Engine) Shutdown(timeout time.Duration) error {
	server := &http.Server{
		Addr:    engine.addr,
		Handler: engine,
	}

	var ctx, cancel = engine.cus_contx.CustomContext()
	defer cancel()

	return server.Shutdown(ctx)

}

// ServeHTTP implements the http.Handler interface
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	engine.router.handle(c)
}

func (engine *Engine) Use(middlewares ...HandlerFunc) {
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
