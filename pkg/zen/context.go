package zen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrEmptyBody = errors.New("request body is empty")
	ErrBadJSON   = errors.New("invalid JSON format")
)

// Context holds the request and response data
type Context struct {
	Writer   *ResponseWriter
	Request  *http.Request
	Params   map[string]string // URL parameters
	Handlers []HandlerFunc     // Slice of middleware functions
	Index    int               // Current position in the middleware chain
	Ctx      context.Context
}

// newContext creates a new Context instance
func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:  NewResponseWriter(w),
		Request: req,
		Index:   -1,
		Params:  make(map[string]string),
		Ctx:     req.Context(),
	}
}

// Next executes the next handler in the chain
func (c *Context) Next() {
	c.Index++                       // move to the next handler
	for c.Index < len(c.Handlers) { // continue while there are handles left
		c.Handlers[c.Index](c) // execute current handler
		c.Index++              // move to the next one
	}
}

// Quit stops the middleware chain execution
func (c *Context) Quit() {
	c.Index = len(c.Handlers)
}

// QuitWithStatus stops the middleware chain execution and writes the status code
func (c *Context) QuitWithStatus(code int) {
	c.Status(code)
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.Quit()
}

// Deadline returns the time when work done on behalf of this context should be canceled.
// The deadline is represented as a time.Time value. ok will be false when no deadline is set.
//
// Usage:
//
//	deadline, ok := c.Deadline()
//	if ok {
//	    fmt.Printf("Work must be completed by: %v\n", deadline)
//	} else {
//	    fmt.Println("No deadline set")
//	}
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this context
// should be canceled. If this context can never be canceled, Done may return nil.
//
// Usage:
//
//	select {
//	case <-c.Done():
//	    fmt.Println("Context canceled, stopping work")
//	    return
//	case <-time.After(5 * time.Second):
//	    fmt.Println("Work completed")
//	}
func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}

// Err returns nil if Done is not yet closed, or a non-nil error explaining why
// the context was canceled if Done is closed.
//
// Usage:
//
//	if err := c.Err(); err != nil {
//	    fmt.Printf("Context error: %v\n", err)
//	    return
//	}
func (c *Context) Err() error {
	return c.Ctx.Err()
}

// Value returns the value associated with this context for a given key, or nil
// if no value is associated with key. Successive calls to Value with the same key
// returns the same result.
//
// Usage:
//
//	userID := c.Value("userID")
//	if id, ok := userID.(string); ok {
//	    fmt.Printf("Current user: %s\n", id)
//	}
func (c *Context) Value(key interface{}) interface{} {
	return c.Ctx.Value(key)
}

// WithValue returns a copy of context with the provided key-value pair.
// The provided key must be comparable and should not be string or any other
// built-in type to avoid collisions.
//
// Usage:
//
//	type contextKey string
//	userKey := contextKey("userID")
//	newCtx := c.WithValue(userKey, "12345")
func (c *Context) WithValue(key, val interface{}) *Context {
	newCtx := *c
	newCtx.Ctx = context.WithValue(c.Ctx, key, val)
	return &newCtx
}

// JSON sends a JSON response to the client
func (c *Context) JSON(code int, obj interface{}) {
	c.SetContentType("application/json")
	c.Writer.WriteHeader(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Text sends a text response
func (c *Context) Text(code int, format string, values ...interface{}) {
	c.SetContentType("text/plain")
	c.Writer.WriteHeader(code)
	if len(values) > 0 {
		format = fmt.Sprintf(format, values...)
	}
	c.Writer.Write([]byte(format))
}

// Status sets the HTTP response status code
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// GetClientIP returns the client IP address
func (c *Context) GetClientIP() string {
	// First, we need to check X-Real-IP header
	ip := c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}

	// check X-Forwarded-For header
	ip = c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// Get the ip for RemoteAddr
	return c.Request.RemoteAddr
}

// ParseJSON parses request body into the provided struct
func (c *Context) ParseJSON(obj interface{}) error {
	if c.Request.Body == nil {
		return ErrEmptyBody
	}

	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return ErrEmptyBody
	}

	if err := json.Unmarshal(body, obj); err != nil {
		return ErrBadJSON
	}

	return nil
}

// TryParseJSON binds the JSON body and returns a boolean indicating success
func (c *Context) TryParseJSON(obj interface{}) bool {
	return c.ParseJSON(obj) == nil
}

// ParseJSONWithError binds JSON and writes an error response if binding fails
func (c *Context) ParseJSONWithError(obj interface{}) bool {
	if err := c.ParseJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return false
	}
	return true
}

// SetHeader sets a header in the response
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// GetHeader returns the value of a header from the request
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetQueryParam adds a query parameter to the request URL
func (c *Context) SetQueryParam(key, value string) {
	query := c.Request.URL.Query()
	query.Set(key, value)
	c.Request.URL.RawQuery = query.Encode()
}

// GetParam returns the value of a URL path parameter defined in the route.
// It retrieves parameters that are part of the URL path defined with ":" prefix.
//
// Usage:
//
//	// For route: "/users/:id/posts/:postId"
//	// URL: "/users/123/posts/456"
//	router.GET("/users/:id/posts/:postId", func(c *Context) {
//	    userID := c.GetParam("id")     // Returns "123"
//	    postID := c.GetParam("postId") // Returns "456"
//	})
func (c *Context) GetParam(key string) string {
	return c.Params[key]
}

// GetQueryParam returns the value of a URL query parameter.
// It retrieves parameters that come after the '?' in the URL.
//
// Usage:
//
//	// For URL: http://example.com/search?query=golang&page=1
//	query := c.GetQueryParam("query")    // Returns "golang"
//	page := c.GetQueryParam("page")      // Returns "1"
//	missing := c.GetQueryParam("absent") // Returns ""
func (c *Context) GetQueryParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetQueryParams returns all query parameters as a map
func (c *Context) GetQueryParams() map[string][]string {
	return c.Request.URL.Query()
}

// SetContentType sets the Content-Type header
func (c *Context) SetContentType(contentType string) {
	c.SetHeader("Content-Type", contentType)
}

// GetContentType returns the Content-Type header
func (c *Context) GetContentType() string {
	return c.GetHeader("Content-Type")
}

// SetCookie sets a cookie in the response
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// GetCookie returns the value of a cookie from the request
func (c *Context) GetCookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// DefaultContext returns custom context with 10 seconds timeout
func (c *Context) DefaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
