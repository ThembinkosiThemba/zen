package middleware

import (
	"maps"
	"strconv"

	"github.com/ThembinkosiThemba/zen"
)

// SecureConfig configures the security headers middleware.
// An empty string or zero value leaves that header unset.
type SecureConfig struct {
	// HSTSMaxAge is the Strict-Transport-Security max-age in seconds.
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// ContentSecurityPolicy is the Content-Security-Policy header value.
	ContentSecurityPolicy string

	// FrameOptions is the X-Frame-Options header value (DENY or SAMEORIGIN).
	FrameOptions string

	// ContentTypeNosniff sets X-Content-Type-Options: nosniff.
	ContentTypeNosniff bool

	// ReferrerPolicy is the Referrer-Policy header value.
	ReferrerPolicy string
}

// DefaultSecureConfig returns a config with common secure defaults.
func DefaultSecureConfig() SecureConfig {
	return SecureConfig{
		HSTSMaxAge:            365 * 24 * 60 * 60,
		HSTSIncludeSubdomains: true,
		ContentSecurityPolicy: "default-src 'self'",
		FrameOptions:          "SAMEORIGIN",
		ContentTypeNosniff:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// Secure returns middleware that sets security headers using DefaultSecureConfig.
func Secure() zen.HandlerFunc {
	return SecureWithConfig(DefaultSecureConfig())
}

// SecureWithConfig returns middleware that sets security headers on every response.
func SecureWithConfig(config SecureConfig) zen.HandlerFunc {
	headers := map[string]string{
		"Content-Security-Policy": config.ContentSecurityPolicy,
		"X-Frame-Options":         config.FrameOptions,
		"Referrer-Policy":         config.ReferrerPolicy,
	}
	if config.ContentTypeNosniff {
		headers["X-Content-Type-Options"] = "nosniff"
	}
	if config.HSTSMaxAge > 0 {
		hsts := "max-age=" + strconv.Itoa(config.HSTSMaxAge)
		if config.HSTSIncludeSubdomains {
			hsts += "; includeSubDomains"
		}
		if config.HSTSPreload {
			hsts += "; preload"
		}
		headers["Strict-Transport-Security"] = hsts
	}
	maps.DeleteFunc(headers, func(_, value string) bool { return value == "" })

	return func(c *zen.Context) {
		h := c.Writer.Header()
		for key, value := range headers {
			h.Set(key, value)
		}
	}
}
