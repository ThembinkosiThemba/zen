package middleware

import (
	"net/http"
	"strings"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// CORSConfig defines the config for CORS middleware
type CORSConfig struct {
	// AllowOrigins defines the origins that are allowed
	// Default is ["*"] which allow all origins
	AllowOrigins []string

	// AllowMethods defines the methods that are allowed.
	// Default is [GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS]
	AllowMethods []string

	// AllowHeaders  defines the headers that are allowed.
	AllowHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	// Default is false
	AllowCredentials bool

	// ExposeHeaders defines the headers that are safe to expose
	// Default is []
	ExposeHeaders []string

	// MaxAge indicates how long the results of a preflight request can be cached
	// Default is 0
	MaxAge int
}

// DefaultCORSConfig returns the default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders:     []string{},
		AllowCredentials: false,
		ExposeHeaders:    []string{},
		MaxAge:           0,
	}
}

// DefaultCors returns the CORS middleware with default configs
func DefaultCors() zen.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns the CORS middleware with custom config
func CORSWithConfig(config CORSConfig) zen.HandlerFunc {
	// we should use the default config is some fields are empty
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig().AllowOrigins
	}

	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig().AllowOrigins
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")

	return func(c *zen.Context) {
		origin := c.Request.Header.Get("Origin")

		// if no origin header is present, skip it
		if origin == "" {
			c.Next()
			return
		}

		// now, we check if origin is allowed
		allowOrigin := "*"
		if len(config.AllowOrigins) != 1 || config.AllowOrigins[0] != "*" {
			for _, o := range config.AllowOrigins {
				if o == origin {
					allowOrigin = origin
					break
				}
			}
		}

		// Setting cors header
		header := c.Writer.Header()
		// TODO: define headers as constants
		header.Set("Access-Control-Allow-Origin", allowOrigin)

		// handling preflight request
		if c.Request.Method == http.MethodOptions {
			header.Set("Access-Control-Allow-Methods", allowMethods)

			if allowHeaders != "" {
				header.Set("Access-Control-Allow-Headers", allowHeaders)
			} else {
				requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
				if requestHeaders != "" {
					header.Set("Access-Control-Allow-Headers", requestHeaders)
				}
			}

			if config.MaxAge > 0 {
				header.Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			if config.AllowCredentials {
				header.Set("Access-Control-Allow-Credentials", "true")
			}

			c.Status(http.StatusNoContent)
		}

		// handling the actual request
		if exposeHeaders != "" {
			header.Set("Access-Control-Expose-Headers", exposeHeaders)
		}

		if config.AllowCredentials {
			header.Set("Access-Control-Allow-Credentials", "true")
		}

		c.Next()
	}
}
