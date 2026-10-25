# Zen

[![CI](https://github.com/ThembinkosiThemba/zen/actions/workflows/zen.yml/badge.svg)](https://github.com/ThembinkosiThemba/zen/actions/workflows/zen.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ThembinkosiThemba/zen.svg)](https://pkg.go.dev/github.com/ThembinkosiThemba/zen)
![Go version](https://img.shields.io/badge/go-%3E%3D1.25-blue)
![License](https://img.shields.io/badge/license-MIT-green)

Zen is a small, fast HTTP framework for Go. It gives you a trie-based router,
route groups, composable middleware and a thin request context, with no
dependencies outside the standard library.

## Installation

```bash
go get github.com/ThembinkosiThemba/zen
```

Requires Go 1.25 or later.

## Quick start

```go
package main

import (
	"net/http"

	"github.com/ThembinkosiThemba/zen"
)

func main() {
	app := zen.Default() // Logger and Recovery middleware

	app.GET("/hello/:name", func(c *zen.Context) {
		c.JSON(http.StatusOK, zen.M{"message": "Hello, " + c.GetParam("name")})
	})

	app.Serve(":8080")
}
```

```bash
$ go run main.go
$ curl localhost:8080/hello/zen
{"message":"Hello, zen"}
```

## Documentation

- [Routing](docs/routing.md): routes, parameters, groups and middleware
- [Context](docs/context.md): reading requests and writing responses
- [Middleware](docs/middleware.md): Logger, Recovery, CORS, JWT auth, rate limiting and security headers
- [Deployment](docs/deployment.md): timeouts, graceful shutdown, TLS and running behind a proxy
- [API reference](https://pkg.go.dev/github.com/ThembinkosiThemba/zen)
- [Example application](examples/basic/main.go)

## Contributing

Issues and pull requests are welcome. Before opening a pull request, run:

```bash
make fmt vet race
```

## License

Zen is released under the [MIT License](LICENSE).
