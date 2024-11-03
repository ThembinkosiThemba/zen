package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
	"github.com/golang-jwt/jwt/v5"
)

// custom errors
var (
	ErrMissingToken = errors.New("missing authorization token")
	ErrInvalidToken = errors.New("invalid authorization token")
)

// Context key for claims
type claimsKey struct{}

// Claims defines the JWT claims
type BaseClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthConfig defines the config for Auth middleware
type AuthConfig struct {
	// secret key used for signing tokens
	SecretKey string

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// TokenHeaderName is a string in the header. Default value is "Bearer"
	TokenHeadName string

	// SkipPaths defines paths that should skip authentication
	SkipPaths []string

	// UnAuthorized defines function to handle unauthorized requests
	Unauthorized func(*zen.Context, error)

	// ClaimsFactory is a function that creates a new claims instance
	// This allows clients to use their own claims structure
	ClaimsFactory func() jwt.Claims
}

// DefaultAuthConfig returns the default auth configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		ClaimsFactory: func() jwt.Claims {
			return &BaseClaims{}
		},
		Unauthorized: func(c *zen.Context, err error) {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": err.Error(),
			})
		},
	}
}

// Auth returns the Auth Middleware with default config
func Auth(secretKey string) zen.HandlerFunc {
	config := DefaultAuthConfig()
	config.SecretKey = secretKey
	return AuthWithConfig(config)
}

// AuthWithConfig returns the Auth middleware with custom config
func AuthWithConfig(config AuthConfig) zen.HandlerFunc {
	if config.SecretKey == "" {
		panic("auth middleware required secret key")
	}

	if config.TokenLookup == "" {
		config.TokenLookup = DefaultAuthConfig().TokenLookup
	}

	if config.TokenHeadName == "" {
		config.TokenHeadName = DefaultAuthConfig().TokenHeadName
	}

	if config.Unauthorized == nil {
		config.Unauthorized = DefaultAuthConfig().Unauthorized
	}

	return func(c *zen.Context) {
		// Check if path shuld be skipped
		for _, path := range config.SkipPaths {
			if path == c.Request.URL.Path {
				c.Next()
				return
			}
		}

		// extracting the token
		token, err := getToken(c, config)
		if err != nil {
			config.Unauthorized(c, err)
			return
		}

		// parsing and validating the token
		claims := config.ClaimsFactory()
		err = validateToken(token, config.SecretKey, claims)
		if err != nil {
			config.Unauthorized(c, ErrInvalidToken)
			return
		}

		// Setting claims to context
		newCtx := c.WithValue(claimsKey{}, claims)
		c.Request = c.Request.WithContext(newCtx.Ctx)
	}
}

func getToken(c *zen.Context, config AuthConfig) (string, error) {
	parts := strings.Split(config.TokenLookup, ":")
	switch parts[0] {
	case "header":
		auth := c.Request.Header.Get(parts[1])
		if auth == "" {
			return "", ErrMissingToken
		}
		if config.TokenHeadName != "" {
			if !strings.HasPrefix(auth, config.TokenHeadName+" ") {
				return "", ErrInvalidToken
			}
			return auth[len(config.TokenHeadName)+1:], nil
		}
		return auth, nil
	case "query":
		token := c.Request.URL.Query().Get(parts[1])
		if token == "" {
			return "", ErrMissingToken
		}
		return token, nil
	case "cookie":
		cookie, err := c.Request.Cookie(parts[1])
		if err != nil {
			return "", ErrMissingToken
		}
		return cookie.Value, nil
	default:
		return "", errors.New("invalid token lookup config")
	}

}

func validateToken(tokenString, secretKey string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}

// GenerateToken generates a new JWT token with custom claims
func GenerateToken(claims jwt.Claims, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// GetClaims retrieves the JWT claims from the context with type assertion
func GetClaims[T jwt.Claims](c *zen.Context) (T, bool) {
	claims, ok := c.Value(claimsKey{}).(T)
	return claims, ok
}
