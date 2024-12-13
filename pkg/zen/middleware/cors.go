package middleware

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// CORSConfig defines the config for CORS middleware
// It allows fine-grained control over Cross-Origin Resource Sharing behavior.
type CORSConfig struct {
	// AllowOrigins defines the origins that are allowed
	// Default is ["*"] which allow all origins
	AllowOrigins []string

	// AllowMethods defines the methods that are allowed.
	// Default is [GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS]
	AllowMethods []string

	// AllowHeaders defines the headers that can be used when making the actual request.
	// Default is [] which allows all headers that are requested.
	// Example: ["Content-Type", "Authorization"]
	AllowHeaders []string

	// AllowCredentials indicates whether the response to the request can be exposed
	// when the credentials flag is true.
	// Default is false.
	// WARNING: Setting this to true can lead to security vulnerabilities if AllowOrigins is ["*"].
	AllowCredentials bool

	// ExposeHeaders defines the headers that are safe to expose to the API of a
	// CORS API specification.
	// Default is [].
	// Example: ["Content-Length", "X-Custom-Header"]
	ExposeHeaders []string

	// MaxAge indicates how long the results of a preflight request can be cached
	// Default is 0 which means no caching.
	// Example: 3600 (1 hour)
	MaxAge int
}

// Cors header constants
const (
	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"
	AccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	AccessControlMaxAge           = "Access-Control-Max-Age"
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	AccessControlRequestHeaders   = "Access-Control-Request-Headers"
	AccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	Vary                          = "Vary"
)

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
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
		ExposeHeaders:    []string{},
		MaxAge:           int(12 * time.Hour.Seconds()), // 12 hours
	}
}

// DefaultCors returns the CORS middleware with default configs
func DefaultCors() zen.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns the CORS middleware with custom config
func CORSWithConfig(config CORSConfig) zen.HandlerFunc {
	log.Println("Here in CORS With COnfig")
	// normalize and validate the configuration
	normalizeConfig(&config)

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(c *zen.Context) {
		origin := c.GetHeader("Origin")

		if origin == "" {
			c.Next()
			return
		}

		c.SetHeader(Vary, "Origin")

		allowOrigin := getAllowOrigin(origin, config.AllowOrigins)
		if allowOrigin == "" {
			c.Status(http.StatusForbidden)
			return
		}

		c.SetHeader(AccessControlAllowOrigin, allowOrigin)

		if config.AllowCredentials {
			c.SetHeader(AccessControlAllowCredentials, "true")
		}

		// preflight
		if c.Request.Method == http.MethodOptions {
			log.Printf("Handling OPTIONS request from origin: %s", origin)

			c.SetHeader(AccessControlAllowMethods, allowMethods)

			requestHeaders := c.GetHeader(AccessControlRequestHeaders)
			if requestHeaders != "" {
				if allowHeaders != "" {
					c.SetHeader(AccessControlAllowHeaders, allowHeaders)
				} else {
					c.SetHeader(AccessControlAllowHeaders, requestHeaders)
				}
			}

			if config.MaxAge > 0 {
				c.SetHeader(AccessControlMaxAge, maxAge)
			}

			c.Status(http.StatusNoContent)
			return

		}

		if exposeHeaders != "" {
			c.SetHeader(AccessControlExposeHeaders, exposeHeaders)
		}

		c.Next()
	}
}

// getAllowOrigin determines the appropriate Allow-Origin header value
func getAllowOrigin(origin string, allowOrigins []string) string {
	if len(allowOrigins) == 1 && allowOrigins[0] == "*" {
		return "*"
	}

	for _, allowOrigin := range allowOrigins {
		if allowOrigin == origin {
			return origin
		}
	}

	return ""
}

// normalizeConfig ensures the configuration has valid values
func normalizeConfig(config *CORSConfig) {
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig().AllowMethods
	}

	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig().AllowOrigins
	}

	// Validate AllowCredentials
	if config.AllowCredentials && contains(config.AllowOrigins, "*") {
		config.AllowOrigins = DefaultCORSConfig().AllowOrigins
		config.AllowCredentials = false // for security
	}
}

// contains checks if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
