package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
)

// SecurityStrategy defines the type of security measures to enable
type SecStrategy uint

// security strategies - can be combined using bitwise OR
const (
	HeaderSecurity SecStrategy = 1 << iota
	RequestSanitization
	IPSecurity
	SessionSecurity
)

// ContentSecurityPolicyDirective represents CSP directives
type ContentSecurityPolicyDirective struct {
	DefaultSrc []string
	ScriptSrc  []string
	StyleSrc   []string
	ImgSrc     []string
	ConnectSrc []string
	FontSrc    []string
	ObjectSrc  []string
	MediaSrc   []string
	FrameSrc   []string
	ReportURI  string
}

// SecurityConfig holds the configuration for the Security middleware
type SecurityConfig struct {
	// Enabled strategies
	Strategies SecStrategy

	// Security headers configuration
	HSTS                  bool // HTTP Strict Transport Security
	HSTSMaxAge            int  // Max age for HSTS in seconds
	HSTSIncludeSubdomains bool // Include subdomains in HSTS
	HSTSPreload           bool // Include in HSTS preload list

	CSPDirectives      *ContentSecurityPolicyDirective
	FrameOptions       string // DENY, SAMEORIGIN, or ALLOW-FROM uri
	ContentTypeOptions bool   // X-Content-Type-Options: nosniff
	XSSProtection      bool   // X-XSS-Protection
	ReferrerPolicy     string // Referrer-Policy header value

	// Request Sanitization Configurations
	MaxRequestSize    int64    // maximum request size in bytes
	AllowedFileTypes  []string // Allowed file extensions for uploads
	SanitizeHTML      bool     // Whether to sanitize HTML in request bodies
	SQLInjectionCheck bool     // SQL injection attacks
	XSSCheck          bool     // Check for XSS attempts

	// IP Security Configuration
	AllowedIPs        []string // Allowed IP addresses/ranges
	BlockedIPs        []string // Blocked IP addresses/ranges
	EnableGeoBlocking bool     // Enable geographic blocking
	AllowedCountries  []string // Allowed country codes

	// Session Security Configuration
	SessionTimeout  int           // Session timeout in seconds
	RotateSessionID bool          // Whether to rotate session IDs
	SecureCookies   bool          // Whether to set Secure flag on cookies
	SameSiteCookies http.SameSite // SameSite cookie policy

	// Custom handlers
	CustomErrorHandler  func(*zen.Context, error)
	OnSecurityViolation func(*zen.Context, string) // Called when a security violation occurs
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Strategies: HeaderSecurity | RequestSanitization,

		// Default security headers
		HSTS:                  true,
		HSTSMaxAge:            31536000, // one year in seconds - // TODO: remove magic numbers
		HSTSIncludeSubdomains: true,
		ContentTypeOptions:    true,
		XSSProtection:         true,
		FrameOptions:          "SAMEORIGIN",
		ReferrerPolicy:        "strict-origin-when-cross-origin",

		CSPDirectives: &ContentSecurityPolicyDirective{
			DefaultSrc: []string{"'self'"},
			ScriptSrc:  []string{"'self'", "'unsafe-inline'"},
			StyleSrc:   []string{"'self'", "'unsafe-inline'"},
			ImgSrc:     []string{"'self'", "data:", "https:"},
			ConnectSrc: []string{"'self'"},
		},

		// Default Request Sanitization
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		AllowedFileTypes:  []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"},
		SanitizeHTML:      true,
		SQLInjectionCheck: true,
		XSSCheck:          true,

		// Default Session Security
		SessionTimeout:  3600, // 1 hour
		RotateSessionID: true,
		SecureCookies:   true,
		SameSiteCookies: http.SameSiteStrictMode,
	}
}

// securityError represents a security-related error
type securityError struct {
	Code    int
	Message string
}

func (e *securityError) Error() string {
	return e.Message
}

// SecurityMiddleware creates a new security middleware with the given configuration
func SecurityMiddleware(config ...SecurityConfig) zen.HandlerFunc {
	var cfg SecurityConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultSecurityConfig()
	}

	return func(c *zen.Context) {
		if err := handleSecurity(c, cfg); err != nil {
			handleSecurityError(c, err, cfg)
			return
		}
		c.Next()
	}
}

// handleSecurity processes all security measures based on configuration
func handleSecurity(c *zen.Context, cfg SecurityConfig) error {
	if cfg.Strategies&HeaderSecurity != 0 {
		setSecurityHeaders(c, cfg)
	}

	// TODO: sanitize requests

	// TODO: ip security

	return nil
}

// setSecurityHeaders sets security-related HTTP headers
func setSecurityHeaders(c *zen.Context, cfg SecurityConfig) {
	// HSTS
	if cfg.HSTS {
		hstsValue := fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge)
		if cfg.HSTSIncludeSubdomains {
			hstsValue += "; includeSubDomains"
		}
		if cfg.HSTSPreload {
			hstsValue += "; preload"
		}
		c.Request.Header.Set("Strict-Transport-Security", hstsValue)
	}

	// CSP
	if cfg.CSPDirectives != nil {
		c.Request.Header.Set("Content-Security-Policy", buildCSPHeader(cfg.CSPDirectives))
	}

	if cfg.ContentTypeOptions {
		c.Request.Header.Set("X-Content-Type-Options", "nosniff")
	}

	if cfg.XSSProtection {
		c.Request.Header.Set("X-XSS-Protection", "1; mode=block")
	}
	c.Request.Header.Set("X-Frame-Options", cfg.FrameOptions)
	c.Request.Header.Set("Referrer-Policy", cfg.ReferrerPolicy)
}

func buildCSPHeader(directives *ContentSecurityPolicyDirective) string {
	var policies []string

	if len(directives.DefaultSrc) > 0 {
		policies = append(policies, fmt.Sprintf("default-src %s", strings.Join(directives.DefaultSrc, " ")))
	}

	if len(directives.ScriptSrc) > 0 {
		policies = append(policies, fmt.Sprintf("script-src %s", strings.Join(directives.ScriptSrc, " ")))
	}

	// TODO: finish all directives

	return ""
}

func handleSecurityError(c *zen.Context, err error, cfg SecurityConfig) {
	if cfg.CustomErrorHandler != nil {
		cfg.CustomErrorHandler(c, err)
		return
	}

	if secErr, ok := err.(*securityError); ok {
		c.Text(secErr.Code, secErr.Message)
		if cfg.OnSecurityViolation != nil {
			cfg.OnSecurityViolation(c, secErr.Message)
		}
		return
	}
	c.Text(http.StatusInternalServerError, "Security error occured")
}
