# Middleware

Zen ships two middleware in the core package and four optional ones in
`github.com/ThembinkosiThemba/zen/middleware`. Each has a `...WithConfig`
variant that takes a config struct. Empty fields fall back to the defaults.

| Middleware | Package      | Purpose                                        |
| ---------- | ------------ | ---------------------------------------------- |
| Logger     | `zen`        | Logs each request                              |
| Recovery   | `zen`        | Turns panics into `500` responses              |
| CORS       | `middleware` | Cross-origin resource sharing                  |
| Auth       | `middleware` | JWT authentication                             |
| RateLimit  | `middleware` | Limits request rate per client                 |
| Secure     | `middleware` | Sets security response headers                 |

## Logger

```go
app.Apply(zen.Logger())
```

```
[ZEN] 2026/10/25 - 19:48:58 | 200 |      59.246µs |       127.0.0.1 | GET     "/users?page=2"
```

Output goes to `os.Stdout` and uses colors only when writing to a terminal.
Set `NO_COLOR` to turn colors off. To write to a file as well, or to skip
noisy paths:

```go
f, err := os.OpenFile("zen.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
if err != nil {
	log.Fatal(err)
}
defer f.Close()

app.Apply(zen.LoggerWithConfig(zen.LoggerConfig{
	Output:    io.MultiWriter(os.Stdout, f),
	SkipPaths: []string{"/health"},
}))
```

## Recovery

```go
app.Apply(zen.Recovery())
```

Recovers from panics in later handlers, logs the stack trace to `os.Stderr`
and responds with `500 Internal Server Error`. If the response has already
started, it is left as is. Add Recovery after Logger so the logger records the
500.

## CORS

```go
app.Apply(middleware.CORS()) // allow any origin
```

```go
app.Apply(middleware.CORSWithConfig(middleware.CORSConfig{
	AllowOrigins:     []string{"https://example.com"},
	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
	AllowHeaders:     []string{"Content-Type", "Authorization"},
	ExposeHeaders:    []string{"X-Total-Count"},
	AllowCredentials: true,
	MaxAge:           3600, // seconds
}))
```

- Preflight requests get `204 No Content` without reaching your routes, so you
  don't need to register `OPTIONS` routes.
- Requests from origins that are not allowed get `403 Forbidden`.
- If `AllowHeaders` is empty, the headers the browser asks for are allowed.
- Combining `AllowCredentials` with the `"*"` origin panics, because browsers
  reject that combination.

Apply CORS globally, so preflight requests are handled even for paths with no
`OPTIONS` route.

## JWT authentication

Apply Auth to the routes that need it, usually through a group:

```go
api := app.Group("/api", middleware.Auth("your-secret-key"))

api.GET("/me", func(c *zen.Context) {
	claims, ok := middleware.GetClaims[*middleware.BaseClaims](c)
	if !ok {
		c.Error(http.StatusUnauthorized, "no claims")
		return
	}
	c.Success(http.StatusOK, claims.UserID, "ok")
})
```

Requests without a valid token are rejected with `401` and never reach the
handler. Only HMAC-signed tokens (HS256, HS384, HS512) are accepted, and
expiry is checked.

### Issuing tokens

```go
token, err := middleware.GenerateToken(&middleware.BaseClaims{
	UserID: "42",
	Role:   "admin",
	RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	},
}, "your-secret-key")
```

`jwt` is `github.com/golang-jwt/jwt/v5`.

### Configuration

```go
middleware.AuthWithConfig(middleware.AuthConfig{
	SecretKey: "your-secret-key",

	// Where to read the token: "header:<name>", "query:<name>" or "cookie:<name>".
	TokenLookup: "cookie:token",

	// Scheme expected before the token in a header. Default "Bearer".
	TokenHeadName: "Bearer",

	// Use your own claims type.
	ClaimsFactory: func() jwt.Claims { return &MyClaims{} },

	// Customise the rejection response.
	Unauthorized: func(c *zen.Context, err error) {
		c.Error(http.StatusUnauthorized, err.Error())
	},
})
```

With a custom claims type, read the claims with
`middleware.GetClaims[*MyClaims](c)`. A missing `SecretKey` or an invalid
`TokenLookup` panics at startup.

## Rate limiting

```go
app.Apply(middleware.RateLimit()) // 100 requests per minute per client IP
```

```go
api := app.Group("/api", middleware.RateLimitWithConfig(middleware.RateLimitConfig{
	Limit:  20,
	Window: time.Second,

	// Count requests per API key instead of per IP.
	KeyFunc: func(c *zen.Context) string { return c.GetHeader("X-API-Key") },

	// Customise the rejection response.
	LimitReached: func(c *zen.Context) {
		c.Error(http.StatusTooManyRequests, "slow down")
	},
}))
```

Rejected requests get `429 Too Many Requests` with a `Retry-After` header.
The limiter uses a sliding window held in memory, so each server process
counts separately. Behind a proxy, configure
[trusted proxies](deployment.md#running-behind-a-proxy) so clients are told
apart by their real IP.

## Security headers

```go
app.Apply(middleware.Secure())
```

The defaults set:

| Header                      | Value                                 |
| --------------------------- | ------------------------------------- |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` |
| `Content-Security-Policy`   | `default-src 'self'`                  |
| `X-Frame-Options`           | `SAMEORIGIN`                          |
| `X-Content-Type-Options`    | `nosniff`                             |
| `Referrer-Policy`           | `strict-origin-when-cross-origin`     |

Start from the defaults and change what you need. An empty string or zero
value leaves that header unset:

```go
cfg := middleware.DefaultSecureConfig()
cfg.ContentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'"
cfg.FrameOptions = "DENY"
cfg.HSTSPreload = true
app.Apply(middleware.SecureWithConfig(cfg))
```

## Recommended order

```go
app := zen.New()
app.Apply(
	zen.Logger(),          // first, so every request is logged
	zen.Recovery(),
	middleware.Secure(),
	middleware.CORS(),     // before auth, so preflight requests succeed
	middleware.RateLimit(),
)

api := app.Group("/api", middleware.Auth(secret))
```
