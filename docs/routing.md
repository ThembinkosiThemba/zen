# Routing

## Engines

```go
app := zen.New()     // no middleware
app := zen.Default() // with zen.Logger() and zen.Recovery()
```

An `*zen.Engine` implements `http.Handler`, so it works anywhere the standard
library expects a handler.

## Routes

```go
app.GET("/users", listUsers)
app.POST("/users", createUser)
app.PUT("/users/:id", replaceUser)
app.PATCH("/users/:id", updateUser)
app.DELETE("/users/:id", deleteUser)
app.HEAD("/users", headUsers)
app.OPTIONS("/users", optionsUsers)

// Any other method.
app.Handle("PROPFIND", "/dav", davHandler)
```

### Path parameters

A segment starting with `:` matches any single path segment:

```go
app.GET("/users/:id/posts/:post", func(c *zen.Context) {
	id := c.GetParam("id")
	post := c.GetParam("post")
	// ...
})
```

A trailing segment starting with `*` matches everything after it. The value
always starts with `/`:

```go
// GET /static/css/app.css -> c.GetParam("filepath") == "/css/app.css"
files := http.StripPrefix("/static", http.FileServer(http.Dir("./public")))
app.GET("/static/*filepath", func(c *zen.Context) {
	files.ServeHTTP(c.Writer, c.Request)
})
```

### Matching rules

- Static segments take priority over parameters, and parameters take priority
  over catch-alls. `/users/new` and `/users/:id` can be registered together.
- Trailing and repeated slashes are ignored, so `/users/` matches `/users`.
- Registering the same method and path twice, or using two different parameter
  names in the same position (`/users/:id` and `/users/:name`), panics at
  startup.

`app.Routes()` returns every registered route in registration order.

## Groups

Groups share a path prefix and a middleware chain:

```go
api := app.Group("/api")
api.GET("/status", status) // GET /api/status

v1 := api.Group("/v1", middleware.Auth(secret))
v1.GET("/users", listUsers) // GET /api/v1/users, requires a token
```

A group copies its parent's middleware when it is created. Groups created from
the same parent don't share middleware with each other.

## Middleware

Middleware is an ordinary `zen.HandlerFunc`. You can attach it at three levels:

```go
app.Apply(zen.Logger())                 // every request
admin := app.Group("/admin", requireAdmin) // every route in the group
admin.DELETE("/users/:id", audit, deleteUser) // a single route
```

Handlers run in order: global middleware, then group middleware, then route
middleware, then the route handler.

> Middleware only applies to routes registered **after** it is added, so call
> `Apply` before defining routes.

### Writing middleware

Call `c.Next()` to run the rest of the chain and continue once it finishes.
Call `c.Quit()` to stop the remaining handlers from running:

```go
func Timer() zen.HandlerFunc {
	return func(c *zen.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s took %v", c.Request.URL.Path, time.Since(start))
	}
}

func requireAdmin(c *zen.Context) {
	claims, ok := middleware.GetClaims[*middleware.BaseClaims](c)
	if !ok || claims.Role != "admin" {
		c.Error(http.StatusForbidden, "admin only")
		c.Quit()
	}
}
```

`c.Quit()` doesn't return from the current function, so `return` after it if
the handler has more work to skip. `c.QuitWithStatus(code)` sets the status code
and stops the chain in one call, and `c.IsQuit()` reports whether the chain was
stopped.

## Not found

Requests that match no route still run global middleware, then get a
`404 NOT FOUND` text response. To customise it:

```go
app.NoRoute(func(c *zen.Context) {
	c.Error(http.StatusNotFound, "route not found")
})
```

## Modes

Zen starts in `zen.DevMode`, which prints the registered routes when the server
starts. Switch to production mode to turn this off:

```go
zen.SetCurrentMode(zen.Production)
```
