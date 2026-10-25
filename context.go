package zen

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmptyBody = errors.New("request body is empty")
	ErrBadJSON   = errors.New("invalid JSON format")
)

// Context carries the request and response through the handler chain.
// It implements context.Context, so it can be passed to code that expects one.
type Context struct {
	Writer  *ResponseWriter
	Request *http.Request

	engine   *Engine
	handlers []HandlerFunc
	index    int
	params   []param

	mu   sync.RWMutex
	keys map[any]any
}

// Next runs the remaining handlers in the chain. Middleware calls it to
// run code after the downstream handlers have finished.
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Quit stops the remaining handlers in the chain from running.
// It does not stop the current handler.
func (c *Context) Quit() {
	c.index = len(c.handlers)
}

// QuitWithStatus stops the chain and sets the response status code.
func (c *Context) QuitWithStatus(code int) {
	c.Status(code)
	c.Quit()
}

// IsQuit reports whether the chain was stopped.
func (c *Context) IsQuit() bool {
	return c.index >= len(c.handlers)
}

// Set stores a value on the context for the lifetime of the request.
func (c *Context) Set(key, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.keys == nil {
		c.keys = make(map[any]any)
	}
	c.keys[key] = value
}

// Get returns a value stored with Set.
func (c *Context) Get(key any) (value any, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok = c.keys[key]
	return value, ok
}

// Deadline implements context.Context using the request's context.
func (c *Context) Deadline() (time.Time, bool) { return c.Request.Context().Deadline() }

// Done implements context.Context using the request's context.
func (c *Context) Done() <-chan struct{} { return c.Request.Context().Done() }

// Err implements context.Context using the request's context.
func (c *Context) Err() error { return c.Request.Context().Err() }

// Value implements context.Context. It returns values stored with Set first,
// then falls back to the request's context.
func (c *Context) Value(key any) any {
	if value, ok := c.Get(key); ok {
		return value
	}
	return c.Request.Context().Value(key)
}

// GetParam returns the value of a URL path parameter.
//
//	app.GET("/users/:id", func(c *zen.Context) {
//	    id := c.GetParam("id")
//	})
func (c *Context) GetParam(key string) string {
	for _, p := range c.params {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

// GetQueryParam returns the first value of a URL query parameter.
func (c *Context) GetQueryParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetHeader returns a request header.
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// GetCookie returns the named request cookie.
func (c *Context) GetCookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie adds a Set-Cookie header to the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// GetClientIP returns the client's IP address. Forwarding headers
// (X-Forwarded-For, X-Real-IP) are only used when the request comes from a
// proxy configured with Engine.SetTrustedProxies, since any client can set them.
func (c *Context) GetClientIP() string {
	remote := c.Request.RemoteAddr
	if host, _, err := net.SplitHostPort(remote); err == nil {
		remote = host
	}

	addr, err := netip.ParseAddr(remote)
	if err != nil || c.engine == nil || !c.engine.isTrustedProxy(addr) {
		return remote
	}

	// Walk X-Forwarded-For from the right, skipping our own proxies.
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		for i, hop := range slices.Backward(strings.Split(xff, ",")) {
			hopAddr, err := netip.ParseAddr(strings.TrimSpace(hop))
			if err != nil {
				break
			}
			if i == 0 || !c.engine.isTrustedProxy(hopAddr) {
				return hopAddr.String()
			}
		}
	}
	if realIP, err := netip.ParseAddr(strings.TrimSpace(c.GetHeader("X-Real-IP"))); err == nil {
		return realIP.String()
	}
	return remote
}

// Status sets the response status code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// JSON writes obj as a JSON response.
func (c *Context) JSON(code int, obj any) {
	data, err := json.Marshal(obj)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	c.Writer.Write(data)
}

// Text writes a plain text response. If values are given, format is
// passed to fmt.Sprintf.
func (c *Context) Text(code int, format string, values ...any) {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.Status(code)
	if len(values) > 0 {
		fmt.Fprintf(c.Writer, format, values...)
		return
	}
	io.WriteString(c.Writer, format)
}

// ParseJSON decodes the JSON request body into obj. It returns ErrEmptyBody
// if there is no body, and an error wrapping ErrBadJSON if decoding fails.
func (c *Context) ParseJSON(obj any) error {
	if c.Request.Body == nil || c.Request.Body == http.NoBody {
		return ErrEmptyBody
	}
	if err := json.NewDecoder(c.Request.Body).Decode(obj); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyBody
		}
		return fmt.Errorf("%w: %v", ErrBadJSON, err)
	}
	return nil
}

// ParseJSONWithError is like ParseJSON, but on failure it responds with
// 400 Bad Request, stops the chain, and returns false.
func (c *Context) ParseJSONWithError(obj any) bool {
	if err := c.ParseJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, M{"error": err.Error()})
		c.Quit()
		return false
	}
	return true
}
