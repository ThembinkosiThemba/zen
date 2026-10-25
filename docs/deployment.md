# Deployment

## Starting a server

```go
app.Serve(":8080")
app.ServeTLS(":8443", "cert.pem", "key.pem")
```

These are shortcuts for `http.ListenAndServe` and `http.ListenAndServeTLS`,
which set no timeouts. They return an error if the address is already in use.

## Timeouts and graceful shutdown

For production, pass the engine to your own `http.Server`. This lets you set
timeouts and finish in-flight requests on shutdown:

```go
func main() {
	zen.SetCurrentMode(zen.Production)

	app := zen.Default()
	app.GET("/", home)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           app,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
```

## Production mode

`zen.SetCurrentMode(zen.Production)` stops zen from printing the route table
and startup messages. Call it before `Serve`.

## Running behind a proxy

By default, `c.GetClientIP()` returns the address of whatever connected to
your server. Behind a load balancer or reverse proxy, that is the proxy.

Tell zen which proxies to trust, as IP addresses or CIDR ranges:

```go
if err := app.SetTrustedProxies("10.0.0.0/8", "127.0.0.1"); err != nil {
	log.Fatal(err)
}
```

For requests from a trusted proxy, zen reads `X-Forwarded-For` from right to
left, skips your own proxies, and returns the first address it doesn't trust.
If that header is missing, it uses `X-Real-IP`. Requests from any other
address ignore both headers, because clients can set them to anything.

This matters for anything that relies on the client IP, such as the
[rate limiter](middleware.md#rate-limiting) and the request logger.

## Running more than one instance

The rate limiter keeps its counters in memory, so each instance limits
separately. With `N` instances behind a load balancer, a client can make up to
`N` times the configured limit. If you need a shared limit, enforce it at the
load balancer or in a store shared by all instances.
