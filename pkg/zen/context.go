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
	Handlers []HandlerFunc
	Index    int
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
	c.Index++
	for c.Index < len(c.Handlers) {
		c.Handlers[c.Index](c)
		c.Index++
	}
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
