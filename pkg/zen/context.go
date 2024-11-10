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

// TODO: funciton to set headers, query etc

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

// context interface methods
// Deadline, Done, Err, Value, WithValue
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}

func (c *Context) Err() error {
	return c.Ctx.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Ctx.Value(key)
}

func (c *Context) WithValue(key, val interface{}) *Context {
	newCtx := *c
	newCtx.Ctx = context.WithValue(c.Ctx, key, val)
	return &newCtx
}

// WriteJson sends a JSON response to the client
func (c *Context) JSON(code int, obj interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Text sends a text response
func (c *Context) Text(code int, format string, values ...interface{}) {
	c.Writer.Header().Set("Content-Type", "text/plain")
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

// GetParam returns the value of the URL parameter
func (c *Context) GetParam(key string) string {
	return c.Params[key]
}

// ClientIP returns the client IP address
func (c *Context) ClientIP() string {
	// First, we need to check X-Real-IP header
	ip := c.Request.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// check X-Forwarded-For header
	ip = c.Request.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// Get the ip for RemoteAddr
	return c.Request.RemoteAddr
}

// BindJSON binds request body to a struct
func (c *Context) BindJSON(obj interface{}) error {
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

// ShouldBindJSON binds the JSON body and returns a boolean indicating success
func (c *Context) ShouldBindJSON(obj interface{}) bool {
	return c.BindJSON(obj) == nil
}

// BindJSONWithError binds JSON and writes an error response if binding fails
func (c *Context) BindJSONWithError(obj interface{}) bool {
	if err := c.BindJSON(obj); err != nil {
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

// GetQueryParam returns the value of a query parameter
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

// CustomContext returns custom context with 10 seconds timeout
func (c *Context) CustomContext() (context.Context, context.CancelFunc) {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	return ctx, cancel
}
