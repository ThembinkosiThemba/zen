package zen

import (
	"fmt"
	"net/http"
)

// Engine is the framework instance
type Engine struct {
	*RouterGroup
	router *Router
	groups []*RouterGroup
}

type Route struct {
	Method string
	Path   string
}

// New creates a new Engine instance
func New() *Engine {
	engine := &Engine{
		router: newRouter(),
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
