package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ThembinkosiThemba/zen/pkg/zen"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	secretKey := "test-secret-key"

	t.Run("successful authentication", func(t *testing.T) {
		// Create a test token
		claims := &BaseClaims{
			UserID: "123",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token, err := GenerateToken(claims, secretKey)
		assert.NoError(t, err)

		// Create test request with token
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		// Create context
		c := zen.NewContext(rec, req)

		// Setup and execute middleware
		middleware := Auth(secretKey)
		c.Handlers = []zen.HandlerFunc{middleware}
		c.Next()

		// Assert response
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify claims are set in request context
		requestCtx := c.Request.Context()
		retrievedClaims, ok := requestCtx.Value(claimsKey{}).(*BaseClaims)
		if assert.True(t, ok) {
			assert.Equal(t, "123", retrievedClaims.UserID)
			assert.Equal(t, "user", retrievedClaims.Role)
		}
	})
}
func TestDefaultAuthConfig(t *testing.T) {
	config := DefaultAuthConfig()
	assert.Equal(t, "header:Authorization", config.TokenLookup)
	assert.Equal(t, "Bearer", config.TokenHeadName)
	assert.NotNil(t, config.ClaimsFactory)
	assert.NotNil(t, config.Unauthorized)
}

func TestAuthWithConfig(t *testing.T) {
	secretKey := "test-secret-key"

	t.Run("custom token lookup - query", func(t *testing.T) {
		claims := &BaseClaims{
			UserID: "123",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token, err := GenerateToken(claims, secretKey)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/test?token="+token, nil)
		rec := httptest.NewRecorder()
		c := zen.NewContext(rec, req)

		config := DefaultAuthConfig()
		config.SecretKey = secretKey
		config.TokenLookup = "query:token"

		middleware := AuthWithConfig(config)
		middleware(c)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("custom token lookup - cookie", func(t *testing.T) {
		claims := &BaseClaims{
			UserID: "123",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token, err := GenerateToken(claims, secretKey)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "auth",
			Value: token,
		})
		rec := httptest.NewRecorder()
		c := zen.NewContext(rec, req)

		config := DefaultAuthConfig()
		config.SecretKey = secretKey
		config.TokenLookup = "cookie:auth"

		middleware := AuthWithConfig(config)
		middleware(c)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("skip paths", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public", nil)
		rec := httptest.NewRecorder()
		c := zen.NewContext(rec, req)

		config := DefaultAuthConfig()
		config.SecretKey = secretKey
		config.SkipPaths = []string{"/public"}

		middleware := AuthWithConfig(config)
		middleware(c)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("custom unauthorized handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := zen.NewContext(rec, req)

		config := DefaultAuthConfig()
		config.SecretKey = secretKey
		config.Unauthorized = func(c *zen.Context, err error) {
			c.JSON(http.StatusForbidden, map[string]interface{}{
				"custom_error": err.Error(),
			})
		}

		middleware := AuthWithConfig(config)
		middleware(c)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestGenerateToken(t *testing.T) {
	secretKey := "secret"
	t.Run("successful token generation", func(t *testing.T) {
		claims := &BaseClaims{
			UserID: "123",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}

		token, err := GenerateToken(claims, secretKey)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token can be parsed back
		parsedClaims := &BaseClaims{}
		err = validateToken(token, secretKey, parsedClaims)
		assert.NoError(t, err)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)
		assert.Equal(t, claims.Role, parsedClaims.Role)
	})
}

func TestValidateToken(t *testing.T) {
	secretKey := "test-secret-key"

	t.Run("valid token", func(t *testing.T) {
		claims := &BaseClaims{
			UserID: "123",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token, _ := GenerateToken(claims, secretKey)

		parsedClaims := &BaseClaims{}
		err := validateToken(token, secretKey, parsedClaims)
		assert.NoError(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		claims := &BaseClaims{
			UserID: "123",
			Role:   "user",
		}
		token, _ := GenerateToken(claims, secretKey)

		parsedClaims := &BaseClaims{}
		err := validateToken(token, "wrong-secret", parsedClaims)
		assert.Error(t, err)
	})

	t.Run("malformed token", func(t *testing.T) {
		parsedClaims := &BaseClaims{}
		err := validateToken("malformed.token.string", secretKey, parsedClaims)
		assert.Error(t, err)
	})
}
