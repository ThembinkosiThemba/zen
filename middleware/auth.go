package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ThembinkosiThemba/zen"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingToken = errors.New("missing authorization token")
	ErrInvalidToken = errors.New("invalid authorization token")
)

type claimsKey struct{}

// BaseClaims are the default JWT claims.
type BaseClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthConfig configures the JWT auth middleware.
type AuthConfig struct {
	// SecretKey is the HMAC key used to verify tokens. Required.
	SecretKey string

	// TokenLookup is "<source>:<name>", where source is header, query or cookie.
	// Defaults to "header:Authorization".
	TokenLookup string

	// TokenHeadName is the scheme expected before the token in a header.
	// Defaults to "Bearer".
	TokenHeadName string

	// Unauthorized writes the response when authentication fails.
	// Defaults to 401 with a JSON error.
	Unauthorized func(*zen.Context, error)

	// ClaimsFactory returns a new claims value to decode tokens into.
	// Defaults to *BaseClaims.
	ClaimsFactory func() jwt.Claims
}

// DefaultAuthConfig returns the default auth configuration, without a secret key.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		ClaimsFactory: func() jwt.Claims { return &BaseClaims{} },
		Unauthorized: func(c *zen.Context, err error) {
			c.JSON(http.StatusUnauthorized, zen.M{"error": err.Error()})
		},
	}
}

// Auth returns JWT auth middleware that verifies tokens signed with secretKey.
func Auth(secretKey string) zen.HandlerFunc {
	config := DefaultAuthConfig()
	config.SecretKey = secretKey
	return AuthWithConfig(config)
}

// AuthWithConfig returns JWT auth middleware with the given config. Requests
// without a valid token are rejected and the chain is stopped. On success the
// claims are available through GetClaims. To leave some routes public, apply
// the middleware to a route group instead of the whole engine.
func AuthWithConfig(config AuthConfig) zen.HandlerFunc {
	defaults := DefaultAuthConfig()
	if config.SecretKey == "" {
		panic("zen/middleware: auth requires a secret key")
	}
	if config.TokenLookup == "" {
		config.TokenLookup = defaults.TokenLookup
	}
	if config.TokenHeadName == "" {
		config.TokenHeadName = defaults.TokenHeadName
	}
	if config.Unauthorized == nil {
		config.Unauthorized = defaults.Unauthorized
	}
	if config.ClaimsFactory == nil {
		config.ClaimsFactory = defaults.ClaimsFactory
	}

	extract, err := tokenExtractor(config.TokenLookup, config.TokenHeadName)
	if err != nil {
		panic("zen/middleware: " + err.Error())
	}
	key := []byte(config.SecretKey)

	return func(c *zen.Context) {
		token, err := extract(c)
		if err != nil {
			config.Unauthorized(c, err)
			c.Quit()
			return
		}

		claims := config.ClaimsFactory()
		if err := validateToken(token, key, claims); err != nil {
			config.Unauthorized(c, ErrInvalidToken)
			c.Quit()
			return
		}

		c.Set(claimsKey{}, claims)
	}
}

func tokenExtractor(lookup, scheme string) (func(*zen.Context) (string, error), error) {
	source, name, ok := strings.Cut(lookup, ":")
	if !ok || name == "" {
		return nil, fmt.Errorf("invalid TokenLookup %q, want \"<source>:<name>\"", lookup)
	}

	switch source {
	case "header":
		prefix := scheme + " "
		return func(c *zen.Context) (string, error) {
			auth := c.GetHeader(name)
			if auth == "" {
				return "", ErrMissingToken
			}
			if len(auth) <= len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
				return "", ErrInvalidToken
			}
			return auth[len(prefix):], nil
		}, nil
	case "query":
		return func(c *zen.Context) (string, error) {
			if token := c.GetQueryParam(name); token != "" {
				return token, nil
			}
			return "", ErrMissingToken
		}, nil
	case "cookie":
		return func(c *zen.Context) (string, error) {
			cookie, err := c.GetCookie(name)
			if err != nil || cookie.Value == "" {
				return "", ErrMissingToken
			}
			return cookie.Value, nil
		}, nil
	default:
		return nil, fmt.Errorf("invalid TokenLookup source %q, want header, query or cookie", source)
	}
}

func validateToken(token string, key []byte, claims jwt.Claims) error {
	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}))
	return err
}

// GenerateToken signs claims with HS256.
func GenerateToken(claims jwt.Claims, secretKey string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
}

// GetClaims returns the claims stored by the auth middleware.
//
//	claims, ok := middleware.GetClaims[*middleware.BaseClaims](c)
func GetClaims[T jwt.Claims](c *zen.Context) (T, bool) {
	value, _ := c.Get(claimsKey{})
	claims, ok := value.(T)
	return claims, ok
}
