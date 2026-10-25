// Package zen is a lightweight HTTP framework for Go.
package zen

import (
	"fmt"
	"net/http"
	"net/netip"
	"slices"
	"strings"
)

// HandlerFunc defines the handler used by zen middleware and routes.
type HandlerFunc func(*Context)

// Route describes a registered route.
type Route struct {
	Method string
	Path   string
}

// Engine is the framework instance. It holds the router and global middleware
// and implements http.Handler, so it can be passed to any http.Server.
type Engine struct {
	RouterGroup

	trees          map[string]*node
	routes         []Route
	maxParams      int
	noRoute        []HandlerFunc
	allNoRoute     []HandlerFunc
	trustedProxies []netip.Prefix
}

// New returns a new blank Engine without any middleware attached.
func New() *Engine {
	engine := &Engine{trees: make(map[string]*node)}
	engine.RouterGroup = RouterGroup{basePath: "/", engine: engine}
	engine.rebuildNoRoute()
	return engine
}

// Default returns an Engine with the Logger and Recovery middleware attached.
func Default() *Engine {
	engine := New()
	engine.Apply(Logger(), Recovery())
	return engine
}

// Apply attaches global middleware. Middleware only applies to routes
// registered after the call, so call Apply before defining routes.
func (engine *Engine) Apply(middleware ...HandlerFunc) {
	engine.RouterGroup.Apply(middleware...)
	engine.rebuildNoRoute()
}

// NoRoute sets the handlers for requests that match no route.
// The default responds with 404 and the body "404 NOT FOUND".
func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	engine.noRoute = handlers
	engine.rebuildNoRoute()
}

func (engine *Engine) rebuildNoRoute() {
	handlers := engine.noRoute
	if len(handlers) == 0 {
		handlers = []HandlerFunc{notFound}
	}
	engine.allNoRoute = engine.combineHandlers(handlers)
}

func notFound(c *Context) {
	c.Text(http.StatusNotFound, "404 NOT FOUND")
}

// SetTrustedProxies sets the proxies (IP addresses or CIDR ranges) whose
// X-Forwarded-For and X-Real-IP headers are honoured by Context.GetClientIP.
// By default no proxy is trusted and the connection's remote address is used.
func (engine *Engine) SetTrustedProxies(proxies ...string) error {
	prefixes := make([]netip.Prefix, 0, len(proxies))
	for _, proxy := range proxies {
		var prefix netip.Prefix
		var err error
		if strings.Contains(proxy, "/") {
			prefix, err = netip.ParsePrefix(proxy)
		} else {
			var addr netip.Addr
			if addr, err = netip.ParseAddr(proxy); err == nil {
				addr = addr.Unmap()
				prefix = netip.PrefixFrom(addr, addr.BitLen())
			}
		}
		if err != nil {
			return fmt.Errorf("zen: invalid trusted proxy %q: %w", proxy, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	engine.trustedProxies = prefixes
	return nil
}

func (engine *Engine) isTrustedProxy(addr netip.Addr) bool {
	addr = addr.Unmap()
	return slices.ContainsFunc(engine.trustedProxies, func(p netip.Prefix) bool {
		return p.Contains(addr)
	})
}

func (engine *Engine) addRoute(method, path string, handlers []HandlerFunc) {
	if method == "" {
		panic("zen: HTTP method can not be empty")
	}
	if len(handlers) == 0 {
		panic("zen: there must be at least one handler")
	}

	root := engine.trees[method]
	if root == nil {
		root = &node{}
		engine.trees[method] = root
	}
	if !root.insert(path, handlers) {
		panic(fmt.Sprintf("zen: route %s %s is already registered", method, path))
	}

	engine.routes = append(engine.routes, Route{Method: method, Path: path})
	if n := countParams(path); n > engine.maxParams {
		engine.maxParams = n
	}
}

// Routes returns the registered routes in registration order.
func (engine *Engine) Routes() []Route {
	return slices.Clone(engine.routes)
}

// ServeHTTP implements http.Handler.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := &Context{
		Writer:  newResponseWriter(w),
		Request: req,
		engine:  engine,
		index:   -1,
	}

	handlers := engine.allNoRoute
	if root := engine.trees[req.Method]; root != nil {
		var params []param
		if engine.maxParams > 0 {
			params = make([]param, 0, engine.maxParams)
		}
		if h, p := root.search(req.URL.Path, params); h != nil {
			handlers, c.params = h, p
		}
	}

	c.handlers = handlers
	c.Next()
	c.Writer.writeHeaderNow()
}

// Serve starts an HTTP server on addr. For timeouts or graceful shutdown,
// pass the engine to your own http.Server as its Handler instead.
func (engine *Engine) Serve(addr string) error {
	engine.debugPrintStartup(addr)
	return http.ListenAndServe(addr, engine)
}

// ServeTLS starts an HTTPS server on addr.
func (engine *Engine) ServeTLS(addr, certFile, keyFile string) error {
	engine.debugPrintStartup(addr)
	return http.ListenAndServeTLS(addr, certFile, keyFile, engine)
}
