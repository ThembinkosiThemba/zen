package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ThembinkosiThemba/zen"
	"github.com/golang-jwt/jwt/v5"
)

const secret = "test-secret-key"

func newToken(t *testing.T, key string, ttl time.Duration) string {
	t.Helper()
	token, err := GenerateToken(&BaseClaims{
		UserID:           "123",
		Role:             "admin",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl))},
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func protectedApp(config AuthConfig) *zen.Engine {
	app := zen.New()
	app.GET("/public", func(c *zen.Context) { c.Text(200, "public") })
	api := app.Group("/api", AuthWithConfig(config))
	api.GET("/me", func(c *zen.Context) {
		claims, ok := GetClaims[*BaseClaims](c)
		if !ok {
			c.Text(500, "claims missing")
			return
		}
		c.Text(200, "%s:%s", claims.UserID, claims.Role)
	})
	return app
}

func withSecret(config AuthConfig) AuthConfig {
	config.SecretKey = secret
	return config
}

func TestAuthHeader(t *testing.T) {
	app := protectedApp(withSecret(DefaultAuthConfig()))

	tests := []struct {
		name, auth string
		code       int
		body       string
	}{
		{"valid", "Bearer " + newToken(t, secret, time.Hour), 200, "123:admin"},
		{"scheme is case-insensitive", "bearer " + newToken(t, secret, time.Hour), 200, "123:admin"},
		{"missing", "", 401, `{"error":"missing authorization token"}`},
		{"wrong scheme", "Basic abc", 401, `{"error":"invalid authorization token"}`},
		{"wrong key", "Bearer " + newToken(t, "other", time.Hour), 401, `{"error":"invalid authorization token"}`},
		{"expired", "Bearer " + newToken(t, secret, -time.Hour), 401, `{"error":"invalid authorization token"}`},
		{"malformed", "Bearer a.b.c", 401, `{"error":"invalid authorization token"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/me", nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			// A rejected request must never reach the handler.
			if w.Code != tt.code || w.Body.String() != tt.body {
				t.Errorf("got %d %q, want %d %q", w.Code, w.Body.String(), tt.code, tt.body)
			}
		})
	}

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/public", nil))
	if w.Code != 200 {
		t.Errorf("public route got %d", w.Code)
	}
}

func TestAuthRejectsNoneAlgorithm(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, &BaseClaims{UserID: "x"}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	app := protectedApp(withSecret(DefaultAuthConfig()))
	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("alg=none token got %d", w.Code)
	}
}

func TestAuthQueryAndCookie(t *testing.T) {
	token := newToken(t, secret, time.Hour)

	app := protectedApp(AuthConfig{SecretKey: secret, TokenLookup: "query:token"})
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/api/me?token="+token, nil))
	if w.Code != 200 {
		t.Errorf("query lookup got %d %q", w.Code, w.Body.String())
	}

	app = protectedApp(AuthConfig{SecretKey: secret, TokenLookup: "cookie:auth"})
	req := httptest.NewRequest("GET", "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "auth", Value: token})
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("cookie lookup got %d %q", w.Code, w.Body.String())
	}
}

func TestAuthCustomUnauthorized(t *testing.T) {
	app := protectedApp(AuthConfig{
		SecretKey:    secret,
		Unauthorized: func(c *zen.Context, err error) { c.Text(http.StatusForbidden, "nope") },
	})
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/api/me", nil))
	if w.Code != http.StatusForbidden || w.Body.String() != "nope" {
		t.Errorf("got %d %q", w.Code, w.Body.String())
	}
}

func TestAuthConfigPanics(t *testing.T) {
	for name, config := range map[string]AuthConfig{
		"no secret":      {},
		"no colon":       {SecretKey: secret, TokenLookup: "header"},
		"unknown source": {SecretKey: secret, TokenLookup: "body:token"},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected panic")
				}
			}()
			AuthWithConfig(config)
		})
	}
}
