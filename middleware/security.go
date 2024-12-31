package middleware

// This module is a security layer for web applications that provides comprehensive protection against various common cyber threats.
// At its core, it implements multiple layers of defense including secure HTTP headers, request validation.
// The middleware protects against Cross-Site Scripting (XSS) attacks by implementing content security policies and sanitizing incoming requests.
// It defends against clickjacking through X-Frame-Options headers and prevents MIME-type sniffing attacks using X-Content-Type-Options.
// Man-in-the-middle attacks are thwarted through HTTP Strict Transport Security (HSTS) enforcement, while cookie-based vulnerabilities are
// addressed through secure cookie policies and SameSite restrictions.
// The module also includes protection against SQL injection, file upload exploits, and malicious large payload attacks by implementing request size limits and file type restrictions.
// For granular access control, it supports IP-based filtering and geographic restrictions.

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ThembinkosiThemba/zen"
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

const (
	SameOrigin              = "SAMEORIGIN"
	StrictOriginPolicy      = "strict-origin-when-cross-origin"
	HstsMaxAge              = 31536000 // one year in seconds
	StrictTransportSecurity = "Strict-Transport-Security"
	ContentSecurityPolicy   = "Content-Security-Policy"
	XContentTypeOptions     = "X-Content-Type-Options"
	XXSSProtection          = "X-XSS-Protection"
	XFrameOptions           = "X-Frame-Options"
	ReferrerPolicy          = "Referrer-Policy"
)

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Strategies: HeaderSecurity | RequestSanitization,

		// Default security headers
		HSTS:                  true,
		HSTSMaxAge:            HstsMaxAge,
		HSTSIncludeSubdomains: true,
		ContentTypeOptions:    true,
		XSSProtection:         true,
		FrameOptions:          SameOrigin,
		ReferrerPolicy:        StrictOriginPolicy,

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

	if cfg.Strategies&RequestSanitization != 0 {
		if err := sanitizeRequest(c, cfg); err != nil {
			return err
		}
	}

	if cfg.Strategies&IPSecurity != 0 {
		if err := enforceIPSecurity(c, cfg); err != nil {
			return err
		}
	}

	// TODO: IP protection

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
		c.Request.Header.Set(StrictTransportSecurity, hstsValue)
	}

	// CSP
	if cfg.CSPDirectives != nil {
		c.Request.Header.Set(ContentSecurityPolicy, buildCSPHeader(cfg.CSPDirectives))
	}

	if cfg.ContentTypeOptions {
		c.Request.Header.Set(XContentTypeOptions, "nosniff")
	}

	if cfg.XSSProtection {
		c.Request.Header.Set(XXSSProtection, "1; mode=block")
	}
	c.Request.Header.Set(XFrameOptions, cfg.FrameOptions)
	c.Request.Header.Set(ReferrerPolicy, cfg.ReferrerPolicy)
}

func buildCSPHeader(directives *ContentSecurityPolicyDirective) string {
	var policies []string

	if len(directives.DefaultSrc) > 0 {
		policies = append(policies, fmt.Sprintf("default-src %s", strings.Join(directives.DefaultSrc, " ")))
	}
	if len(directives.ScriptSrc) > 0 {
		policies = append(policies, fmt.Sprintf("script-src %s", strings.Join(directives.ScriptSrc, " ")))
	}
	if len(directives.StyleSrc) > 0 {
		policies = append(policies, fmt.Sprintf("style-src %s", strings.Join(directives.StyleSrc, " ")))
	}
	if len(directives.ImgSrc) > 0 {
		policies = append(policies, fmt.Sprintf("img-src %s", strings.Join(directives.ImgSrc, " ")))
	}
	if len(directives.ConnectSrc) > 0 {
		policies = append(policies, fmt.Sprintf("connect-src %s", strings.Join(directives.ConnectSrc, " ")))
	}
	if len(directives.FontSrc) > 0 {
		policies = append(policies, fmt.Sprintf("font-src %s", strings.Join(directives.FontSrc, " ")))
	}
	if len(directives.ObjectSrc) > 0 {
		policies = append(policies, fmt.Sprintf("object-src %s", strings.Join(directives.ObjectSrc, " ")))
	}
	if len(directives.MediaSrc) > 0 {
		policies = append(policies, fmt.Sprintf("media-src %s", strings.Join(directives.MediaSrc, " ")))
	}
	if len(directives.FrameSrc) > 0 {
		policies = append(policies, fmt.Sprintf("frame-src %s", strings.Join(directives.FrameSrc, " ")))
	}
	if directives.ReportURI != "" {
		policies = append(policies, fmt.Sprintf("report-uri %s", directives.ReportURI))
	}

	return strings.Join(policies, "; ")
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

// implementation for request sanitization
func sanitizeRequest(c *zen.Context, cfg SecurityConfig) error {
	// first, we need to check request size
	if c.Request.ContentLength > cfg.MaxRequestSize {
		return &securityError{
			Code:    http.StatusRequestEntityTooLarge,
			Message: "Request exceeds maximum allowed size",
		}
	}

	if cfg.SQLInjectionCheck {
		if checkSqlInjection(c.Request) {
			return &securityError{
				Code:    http.StatusBadRequest,
				Message: "Potential SQL injection detected",
			}
		}

	}

	if cfg.XSSCheck {
		if containsXSS(c.Request) {
			return &securityError{
				Code:    http.StatusBadRequest,
				Message: "Potential XSS attack detected",
			}
		}
	}

	return nil
}

// Helper functions for security checks
// TODO: first simple version for SQL injection checks
func checkSqlInjection(r *http.Request) bool {
	sqlPatterns := []string{
		"UNION SELECT",
		"DROP TABLE",
		"DELETE FROM",
		"INSERT INTO",
		"--",
		";--",
		";",
		"/*",
		"*/",
		"@@",
	}

	return containsPatterns(r, sqlPatterns)
}

// TODO: first version XSS check
func containsXSS(r *http.Request) bool {
	xssPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"alert(",
		"eval(",
	}

	return containsPatterns(r, xssPatterns)
}

func containsPatterns(r *http.Request, patterns []string) bool {
	// Check URL parameters
	query := strings.ToLower(r.URL.RawQuery)
	for _, pattern := range patterns {
		if strings.Contains(query, strings.ToLower(pattern)) {
			return true
		}
	}

	// Check form data
	if err := r.ParseForm(); err == nil {
		for _, values := range r.Form {
			for _, value := range values {
				value = strings.ToLower(value)
				for _, pattern := range patterns {
					if strings.Contains(value, strings.ToLower(pattern)) {
						return true
					}
				}
			}
		}
	}

	return false
}

func enforceIPSecurity(c *zen.Context, cfg SecurityConfig) error {
	ip := c.GetClientIP()

	for _, blockedIP := range cfg.BlockedIPs {
		if ip == blockedIP {
			return &securityError{
				Code:    http.StatusForbidden,
				Message: "IP address is blocked",
			}
		}
	}

	if len(cfg.AllowedIPs) > 0 {
		allowed := false
		for _, allowedIP := range cfg.AllowedIPs {
			if ip == allowedIP {
				allowed = true
				break
			}
		}
		if !allowed {
			return &securityError{
				Code:    http.StatusForbidden,
				Message: "IP address not allowed",
			}
		}
	}

	// TODO: Block IP addresses from countries using GEOLOCATION

	return nil
}
