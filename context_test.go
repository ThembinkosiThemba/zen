package zen

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	app := New()
	app.GET("/", func(c *Context) { c.JSON(http.StatusCreated, M{"name": "zen"}) })
	app.GET("/bad", func(c *Context) { c.JSON(http.StatusOK, make(chan int)) })

	w := perform(app, "GET", "/")
	if w.Code != http.StatusCreated || w.Body.String() != `{"name":"zen"}` {
		t.Errorf("got %d %q", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}

	if w := perform(app, "GET", "/bad"); w.Code != http.StatusInternalServerError {
		t.Errorf("unencodable value: got %d", w.Code)
	}
}

func TestText(t *testing.T) {
	app := New()
	app.GET("/fmt", func(c *Context) { c.Text(200, "Hello %s", "World") })

	if w := perform(app, "GET", "/fmt"); w.Body.String() != "Hello World" {
		t.Errorf("body = %q", w.Body.String())
	}
}

func TestResponseEnvelope(t *testing.T) {
	app := New()
	app.GET("/ok", func(c *Context) { c.Success(200, M{"id": 1}, "done") })
	app.GET("/err", func(c *Context) { c.Error(400, "bad input", "name is required") })

	if got := perform(app, "GET", "/ok").Body.String(); got != `{"status":200,"success":true,"data":{"id":1},"message":"done"}` {
		t.Errorf("Success body = %s", got)
	}
	if got := perform(app, "GET", "/err").Body.String(); got != `{"status":400,"success":false,"data":"name is required","message":"bad input"}` {
		t.Errorf("Error body = %s", got)
	}
}

func TestStatusIsSentWithoutBody(t *testing.T) {
	app := New()
	app.GET("/", func(c *Context) { c.Status(http.StatusAccepted) })
	if w := perform(app, "GET", "/"); w.Code != http.StatusAccepted {
		t.Errorf("got %d", w.Code)
	}
}

func TestHeadersSetAfterStatusAreSent(t *testing.T) {
	app := New()
	app.Apply(func(c *Context) {
		c.Next()
		c.SetHeader("X-After", "1")
	})
	app.GET("/", func(c *Context) { c.Status(http.StatusNoContent) })

	w := perform(app, "GET", "/")
	if w.Code != http.StatusNoContent || w.Result().Header.Get("X-After") != "1" {
		t.Errorf("got %d, X-After=%q", w.Code, w.Result().Header.Get("X-After"))
	}
}

func TestParseJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	tests := []struct {
		body    string
		wantErr error
	}{
		{`{"name":"zen"}`, nil},
		{`{"name":"zen"`, ErrBadJSON},
		{`{"name":1}`, ErrBadJSON},
		{``, ErrEmptyBody},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
		c := &Context{Request: req}
		var p payload
		if err := c.ParseJSON(&p); !errors.Is(err, tt.wantErr) {
			t.Errorf("ParseJSON(%q) error = %v, want %v", tt.body, err, tt.wantErr)
		}
	}
}

func TestParseJSONWithError(t *testing.T) {
	app := New()
	app.POST("/", func(c *Context) {
		var v map[string]any
		if !c.ParseJSONWithError(&v) {
			return
		}
		c.Text(200, "ok")
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader("{"))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "invalid JSON") {
		t.Errorf("got %d %q", w.Code, w.Body.String())
	}
}

func TestGetClientIP(t *testing.T) {
	app := New()
	if err := app.SetTrustedProxies("10.0.0.0/8"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, remote, xff, realIP, want string
	}{
		{"strips port", "1.2.3.4:5678", "", "", "1.2.3.4"},
		{"ipv6", "[::1]:5678", "", "", "::1"},
		{"untrusted remote ignores headers", "1.2.3.4:1", "9.9.9.9", "8.8.8.8", "1.2.3.4"},
		{"trusted proxy uses XFF", "10.0.0.1:1", "9.9.9.9", "", "9.9.9.9"},
		{"skips trusted hops", "10.0.0.1:1", "9.9.9.9, 7.7.7.7, 10.0.0.2", "", "7.7.7.7"},
		{"trusted proxy uses X-Real-IP", "10.0.0.1:1", "", "8.8.8.8", "8.8.8.8"},
		{"invalid header falls back", "10.0.0.1:1", "", "garbage", "10.0.0.1"},
		{"ipv4-mapped proxy is trusted", "[::ffff:10.0.0.1]:1", "9.9.9.9", "", "9.9.9.9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remote
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			c := &Context{Request: req, engine: app}
			if got := c.GetClientIP(); got != tt.want {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeysAndContextValue(t *testing.T) {
	type ctxKey struct{}
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "from-request"))
	c := &Context{Request: req}

	if c.Value(ctxKey{}) != "from-request" {
		t.Error("Value should fall back to the request context")
	}
	c.Set("user", "zen")
	if v, ok := c.Get("user"); !ok || v != "zen" {
		t.Errorf("Get = %v, %v", v, ok)
	}
	if c.Value("user") != "zen" {
		t.Error("Value should return keys set with Set")
	}

	// The request can be replaced and the context still follows it.
	ctx, cancel := context.WithCancel(context.Background())
	c.Request = c.Request.WithContext(ctx)
	cancel()
	if c.Err() != context.Canceled {
		t.Errorf("Err() = %v, want context.Canceled", c.Err())
	}
}

func TestParamsQueryHeadersCookies(t *testing.T) {
	app := New()
	app.GET("/u/:id", func(c *Context) {
		cookie, err := c.GetCookie("session")
		if err != nil {
			t.Fatal(err)
		}
		c.SetCookie(&http.Cookie{Name: "seen", Value: "1"})
		c.Text(200, "%s %s %s %s", c.GetParam("id"), c.GetQueryParam("q"), c.GetHeader("X-Test"), cookie.Value)
	})

	req := httptest.NewRequest("GET", "/u/7?q=search", nil)
	req.Header.Set("X-Test", "hdr")
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc"})
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Body.String() != "7 search hdr abc" {
		t.Errorf("body = %q", w.Body.String())
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), "seen=1") {
		t.Errorf("Set-Cookie = %q", w.Header().Get("Set-Cookie"))
	}
}

func TestResponseWriterInterfaces(t *testing.T) {
	var _ http.Flusher = (*ResponseWriter)(nil)
	var _ http.Hijacker = (*ResponseWriter)(nil)

	rec := httptest.NewRecorder()
	w := newResponseWriter(rec)
	if w.Written() || w.Status() != 200 || w.Size() != -1 {
		t.Errorf("fresh writer: written=%v status=%d size=%d", w.Written(), w.Status(), w.Size())
	}
	w.WriteHeader(201)
	w.Write([]byte("hi"))
	w.WriteHeader(500) // ignored once written
	if rec.Code != 201 || w.Status() != 201 || w.Size() != 2 {
		t.Errorf("code=%d status=%d size=%d", rec.Code, w.Status(), w.Size())
	}
	if http.NewResponseController(w).Flush() != nil {
		t.Error("ResponseController should reach the underlying writer")
	}
}
