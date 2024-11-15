package zen

import (
	"fmt"
	"net/http"
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
	engine.printRoutes()
	fmt.Print(engine.zenAsciiArt(addr))
	return http.ListenAndServe(addr, engine)
}

// ServeTLS start the HTTPS server
func (engine *Engine) ServeTLS(addr, certFile, keyFile string) error {
	engine.printRoutes()
	fmt.Print(engine.zenAsciiArt(addr))
	return http.ListenAndServeTLS(addr, certFile, keyFile, engine)
}

// ServeWithTimeout starts the HTTP server with timeout settings
func (engine *Engine) ServeWithTimeout(addr string, timeout time.Duration) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout * 2,
	}
	engine.printRoutes()
	fmt.Print(engine.zenAsciiArt(addr))
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
