package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ThembinkosiThemba/zen"
)

func corsApp(config CORSConfig, handlerRan *bool) *zen.Engine {
	app := zen.New()
	app.Apply(CORSWithConfig(config))
	app.POST("/data", func(c *zen.Context) {
		*handlerRan = true
		c.Text(200, "ok")
	})
	return app
}

func request(app http.Handler, method, path string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w
}

func TestCORSAllowedOrigin(t *testing.T) {
	var ran bool
	app := corsApp(CORSConfig{
		AllowOrigins:     []string{"https://good.com"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Total"},
	}, &ran)

	w := request(app, "POST", "/data", "Origin", "https://good.com")
	h := w.Header()
	if !ran || w.Code != 200 {
		t.Fatalf("handler ran=%v code=%d", ran, w.Code)
	}
	if h.Get("Access-Control-Allow-Origin") != "https://good.com" ||
		h.Get("Access-Control-Allow-Credentials") != "true" ||
		h.Get("Access-Control-Expose-Headers") != "X-Total" ||
		h.Get("Vary") != "Origin" {
		t.Errorf("unexpected headers: %v", h)
	}
}

func TestCORSDisallowedOriginStopsChain(t *testing.T) {
	var ran bool
	app := corsApp(CORSConfig{AllowOrigins: []string{"https://good.com"}}, &ran)

	w := request(app, "POST", "/data", "Origin", "https://evil.com")
	if ran || w.Code != http.StatusForbidden {
		t.Errorf("handler ran=%v code=%d", ran, w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("disallowed origin must not get Access-Control-Allow-Origin")
	}
}

func TestCORSPreflight(t *testing.T) {
	var ran bool
	app := corsApp(DefaultCORSConfig(), &ran)

	// The preflight is answered even though no OPTIONS route exists.
	w := request(app, "OPTIONS", "/data",
		"Origin", "https://any.com",
		"Access-Control-Request-Method", "POST",
		"Access-Control-Request-Headers", "Content-Type")
	h := w.Header()
	if ran || w.Code != http.StatusNoContent {
		t.Fatalf("handler ran=%v code=%d", ran, w.Code)
	}
	if h.Get("Access-Control-Allow-Origin") != "*" ||
		h.Get("Access-Control-Allow-Methods") == "" ||
		h.Get("Access-Control-Allow-Headers") != "Origin, Content-Type, Accept, Authorization" ||
		h.Get("Access-Control-Max-Age") != "43200" {
		t.Errorf("unexpected headers: %v", h)
	}
}

func TestCORSReflectsRequestedHeaders(t *testing.T) {
	var ran bool
	app := corsApp(CORSConfig{AllowOrigins: []string{"*"}}, &ran)
	w := request(app, "OPTIONS", "/data",
		"Origin", "https://any.com",
		"Access-Control-Request-Method", "POST",
		"Access-Control-Request-Headers", "X-Custom")
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "X-Custom" {
		t.Errorf("Access-Control-Allow-Headers = %q", got)
	}
}

func TestCORSNoOrigin(t *testing.T) {
	var ran bool
	app := corsApp(DefaultCORSConfig(), &ran)
	w := request(app, "POST", "/data")
	if !ran || w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("same-origin request: ran=%v headers=%v", ran, w.Header())
	}
}

func TestCORSCredentialsWithWildcardPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	CORSWithConfig(CORSConfig{AllowOrigins: []string{"*"}, AllowCredentials: true})
}
