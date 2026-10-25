// Package middleware provides optional middleware for zen.
package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ThembinkosiThemba/zen"
)

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	// AllowOrigins lists the allowed origins. ["*"] allows any origin.
	AllowOrigins []string

	// AllowMethods lists the methods allowed in preflight requests.
	AllowMethods []string

	// AllowHeaders lists the request headers allowed in preflight requests.
	// If empty, the headers requested by the browser are allowed.
	AllowHeaders []string

	// ExposeHeaders lists response headers the browser may expose to scripts.
	ExposeHeaders []string

	// AllowCredentials allows cookies and auth headers on cross-origin requests.
	// It can not be combined with a wildcard origin.
	AllowCredentials bool

	// MaxAge is how long, in seconds, browsers may cache a preflight response.
	MaxAge int
}

// DefaultCORSConfig returns a permissive config that allows any origin.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
			http.MethodDelete, http.MethodHead, http.MethodOptions,
		},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       12 * 60 * 60,
	}
}

// CORS returns CORS middleware using DefaultCORSConfig.
func CORS() zen.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns CORS middleware with the given config. Requests from
// origins that are not allowed are rejected with 403 Forbidden, and preflight
// requests are answered with 204 No Content without reaching the route.
// It panics if AllowCredentials is combined with a wildcard origin.
func CORSWithConfig(config CORSConfig) zen.HandlerFunc {
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig().AllowMethods
	}

	allowAll := len(config.AllowOrigins) == 0
	origins := make(map[string]struct{}, len(config.AllowOrigins))
	for _, o := range config.AllowOrigins {
		if o == "*" {
			allowAll = true
		}
		origins[o] = struct{}{}
	}
	if allowAll && config.AllowCredentials {
		panic("zen/middleware: CORS AllowCredentials can not be used with a wildcard origin")
	}

	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(c *zen.Context) {
		header := c.Writer.Header()
		origin := c.GetHeader("Origin")
		if !allowAll {
			header.Add("Vary", "Origin")
		}
		if origin == "" {
			return
		}

		if _, ok := origins[origin]; !ok && !allowAll {
			c.QuitWithStatus(http.StatusForbidden)
			return
		}

		if allowAll {
			header.Set("Access-Control-Allow-Origin", "*")
		} else {
			header.Set("Access-Control-Allow-Origin", origin)
		}
		if config.AllowCredentials {
			header.Set("Access-Control-Allow-Credentials", "true")
		}

		isPreflight := c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != ""
		if !isPreflight {
			if exposeHeaders != "" {
				header.Set("Access-Control-Expose-Headers", exposeHeaders)
			}
			return
		}

		header.Set("Access-Control-Allow-Methods", allowMethods)
		if allowHeaders != "" {
			header.Set("Access-Control-Allow-Headers", allowHeaders)
		} else if requested := c.GetHeader("Access-Control-Request-Headers"); requested != "" {
			header.Set("Access-Control-Allow-Headers", requested)
			header.Add("Vary", "Access-Control-Request-Headers")
		}
		if config.MaxAge > 0 {
			header.Set("Access-Control-Max-Age", maxAge)
		}
		c.QuitWithStatus(http.StatusNoContent)
	}
}
