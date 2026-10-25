package zen

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	app := New()
	app.Apply(LoggerWithConfig(LoggerConfig{Output: &buf, SkipPaths: []string{"/health"}}))
	app.GET("/users/:id", func(c *Context) { c.Status(http.StatusCreated) })
	app.GET("/health", text("ok"))

	req := httptest.NewRequest("GET", "/users/1?x=y", nil)
	req.RemoteAddr = "1.2.3.4:5000"
	app.ServeHTTP(httptest.NewRecorder(), req)

	line := buf.String()
	for _, want := range []string{"[ZEN]", "201", "GET", `"/users/1?x=y"`, "1.2.3.4"} {
		if !strings.Contains(line, want) {
			t.Errorf("log line %q missing %q", line, want)
		}
	}
	if strings.Contains(line, "\033[") {
		t.Error("colors should be disabled for non-terminal output")
	}

	buf.Reset()
	perform(app, "GET", "/health")
	perform(app, "GET", "/missing")
	if strings.Contains(buf.String(), "/health") || !strings.Contains(buf.String(), "404") {
		t.Errorf("unexpected log output %q", buf.String())
	}
}

func TestRecovery(t *testing.T) {
	app := New()
	app.Apply(Recovery())
	app.GET("/panic", func(c *Context) { panic("boom") })
	app.GET("/partial", func(c *Context) {
		c.Text(http.StatusOK, "partial")
		panic("boom")
	})

	w := perform(app, "GET", "/panic")
	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("got %d %q", w.Code, w.Body.String())
	}

	// Once the response has started, Recovery must not append a second body.
	w = perform(app, "GET", "/partial")
	if w.Code != http.StatusOK || w.Body.String() != "partial" {
		t.Errorf("got %d %q", w.Code, w.Body.String())
	}
}

func TestRecoveryRepanicsAbortHandler(t *testing.T) {
	app := New()
	app.Apply(Recovery())
	app.GET("/", func(c *Context) { panic(http.ErrAbortHandler) })

	defer func() {
		if recover() != http.ErrAbortHandler {
			t.Error("expected http.ErrAbortHandler to propagate")
		}
	}()
	perform(app, "GET", "/")
}

func TestDefault(t *testing.T) {
	app := Default()
	app.GET("/", func(c *Context) { panic("boom") })
	// Logger writes to stdout here; just check Recovery is installed.
	if w := perform(app, "GET", "/"); w.Code != http.StatusInternalServerError {
		t.Errorf("got %d", w.Code)
	}
}
