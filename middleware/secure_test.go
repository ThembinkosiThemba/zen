package middleware

import (
	"testing"

	"github.com/ThembinkosiThemba/zen"
)

func TestSecureHeaders(t *testing.T) {
	app := zen.New()
	app.Apply(Secure())
	app.GET("/", func(c *zen.Context) { c.Text(200, "ok") })

	// Query strings with ";" or "--" are ordinary input and must not be rejected.
	w := request(app, "GET", "/?q=a;b--c")
	if w.Code != 200 || w.Body.String() != "ok" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}

	want := map[string]string{
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
		"X-Frame-Options":           "SAMEORIGIN",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}
	for key, value := range want {
		if got := w.Result().Header.Get(key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
}

func TestSecureEmptyConfigSetsNothing(t *testing.T) {
	app := zen.New()
	app.Apply(SecureWithConfig(SecureConfig{FrameOptions: "DENY"}))
	app.GET("/", func(c *zen.Context) {})

	h := request(app, "GET", "/").Result().Header
	if h.Get("X-Frame-Options") != "DENY" {
		t.Errorf("X-Frame-Options = %q", h.Get("X-Frame-Options"))
	}
	for _, key := range []string{"Strict-Transport-Security", "Content-Security-Policy", "Referrer-Policy", "X-Content-Type-Options"} {
		if _, ok := h[key]; ok {
			t.Errorf("%s should not be set", key)
		}
	}
}
