# Context

Every handler receives a `*zen.Context`. It wraps the `*http.Request` and the
response writer and carries values through the handler chain.

## Reading the request

```go
c.GetParam("id")         // path parameter from /users/:id
c.GetQueryParam("page")  // first value of ?page=
c.GetHeader("User-Agent")
cookie, err := c.GetCookie("session")
ip := c.GetClientIP()
```

The underlying request is always available as `c.Request`, for example
`c.Request.Method` or `c.Request.URL.Query()["tag"]`.

### JSON bodies

```go
type CreateUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

app.POST("/users", func(c *zen.Context) {
	var body CreateUser
	if err := c.ParseJSON(&body); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}
	// ...
})
```

`ParseJSON` returns `zen.ErrEmptyBody` when there is no body, and an error
wrapping `zen.ErrBadJSON` when decoding fails. Check them with `errors.Is`.

`ParseJSONWithError` does the same, but on failure it responds with
`400 Bad Request`, stops the chain and returns `false`:

```go
if !c.ParseJSONWithError(&body) {
	return
}
```

### Client IP

`c.GetClientIP()` returns the IP address of the connection, without the port.
It only reads `X-Forwarded-For` or `X-Real-IP` when the request comes from a
proxy you trust, because any client can set those headers. See
[Running behind a proxy](deployment.md#running-behind-a-proxy).

## Writing the response

```go
c.JSON(http.StatusOK, zen.M{"id": 1})        // application/json
c.Text(http.StatusOK, "Hello %s", name)      // text/plain, printf-style
c.Status(http.StatusNoContent)               // status only
c.SetHeader("Cache-Control", "no-store")
c.SetCookie(&http.Cookie{Name: "session", Value: id, HttpOnly: true})
```

`zen.M` is shorthand for `map[string]any`.

Headers are sent on the first write to the body, or when the handler chain
finishes. Until then, middleware can still change the status code and headers.
`c.Writer` implements `http.ResponseWriter`, `http.Flusher` and
`http.Hijacker`, and also provides `Status()`, `Size()` and `Written()`.

### Response envelope

`Success` and `Error` write a consistent JSON shape:

```go
c.Success(http.StatusOK, user, "user retrieved")
c.Error(http.StatusBadRequest, "validation failed", fieldErrors)
```

```json
{ "status": 200, "success": true, "data": { "id": 1 }, "message": "user retrieved" }
{ "status": 400, "success": false, "data": { "email": "required" }, "message": "validation failed" }
```

The last argument to `Error` is optional and becomes `data`.

## Request-scoped values

Use `Set` and `Get` to pass values between middleware and handlers:

```go
c.Set("requestID", id)

if id, ok := c.Get("requestID"); ok {
	// ...
}
```

Keys can be any comparable type. Prefer an unexported key type in packages so
keys can't collide.

## Using Context as a context.Context

`*zen.Context` implements `context.Context`. Its deadline and cancellation come
from the request, so work stops when the client disconnects:

```go
rows, err := db.QueryContext(c, "SELECT ...")
```

`c.Value(key)` returns values stored with `Set` first, then falls back to the
request's context.

> A `*zen.Context` is only valid while the handler is running. Don't use it
> from goroutines that outlive the request.
