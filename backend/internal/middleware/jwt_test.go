package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration
var testJWTConfig = config.JWTConfig{
	Secret: "test-secret-key-for-testing",
	Issuer: "wayfarer-test",
}

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// generateTestToken creates a valid test JWT token
func generateTestToken(userID, userRole string, expiresIn time.Duration) string {
	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: []string{userRole},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testJWTConfig.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTConfig.Secret))
	return tokenString
}

// generateExpiredToken creates an expired test JWT token
func generateExpiredToken(userID, userRole string) string {
	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: []string{userRole},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testJWTConfig.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTConfig.Secret))
	return tokenString
}

// generateTokenWithWrongSecret creates a token signed with wrong secret
func generateTokenWithWrongSecret(userID, userRole string) string {
	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: []string{userRole},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testJWTConfig.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret-key"))
	return tokenString
}

// generateTokenWithWrongIssuer creates a token with wrong issuer
func generateTokenWithWrongIssuer(userID, userRole string) string {
	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: []string{userRole},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wrong-issuer",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTConfig.Secret))
	return tokenString
}

func TestJWTAuth_ValidToken(t *testing.T) {
	// Create a valid token
	token := generateTestToken("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user", time.Hour)

	// Create test router
	router := gin.New()
	router.Use(JWTAuth(testJWTConfig))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", userID)

		userRoles, exists := c.Get("user_roles")
		assert.True(t, exists)
		assert.Equal(t, []string{"user"}, userRoles)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuth_MissingAuthorizationHeader(t *testing.T) {
	// Create test router
	router := gin.New()
	router.Use(JWTAuth(testJWTConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestJWTAuth_InvalidAuthorizationHeaderFormat(t *testing.T) {
	testCases := []struct {
		name          string
		header        string
		expectedError string
	}{
		{
			name:          "Missing Bearer prefix",
			header:        "some-token",
			expectedError: "Invalid authorization header format",
		},
		{
			name:          "Wrong prefix",
			header:        "Basic some-token",
			expectedError: "Invalid authorization header format",
		},
		{
			name:          "Bearer without token",
			header:        "Bearer",
			expectedError: "Invalid authorization header format",
		},
		{
			name:          "Empty after Bearer",
			header:        "Bearer ",
			expectedError: "Invalid or expired token", // Empty token is treated as invalid token
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test router
			router := gin.New()
			router.Use(JWTAuth(testJWTConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.header)

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedError)
		})
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	// Create an expired token
	token := generateExpiredToken("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user")

	// Create test router
	router := gin.New()
	router.Use(JWTAuth(testJWTConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestJWTAuth_InvalidSignature(t *testing.T) {
	// Create a token with wrong secret
	token := generateTokenWithWrongSecret("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user")

	// Create test router
	router := gin.New()
	router.Use(JWTAuth(testJWTConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestJWTAuth_InvalidIssuer(t *testing.T) {
	// Create a token with wrong issuer
	token := generateTokenWithWrongIssuer("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user")

	// Create test router
	router := gin.New()
	router.Use(JWTAuth(testJWTConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestJWTAuth_MalformedToken(t *testing.T) {
	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "Random string",
			token: "not-a-jwt-token",
		},
		{
			name:  "Incomplete JWT",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.incomplete",
		},
		{
			name:  "Empty token",
			token: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test router
			router := gin.New()
			router.Use(JWTAuth(testJWTConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestJWTAuth_DifferentRoles(t *testing.T) {
	testCases := []struct {
		name     string
		userID   string
		userRole string
	}{
		{
			name:     "User role",
			userID:   "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			userRole: "user",
		},
		{
			name:     "Admin role",
			userID:   "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			userRole: "admin",
		},
		{
			name:     "M2M role",
			userID:   "SY01ARZ3NDEKTSV4RRFFQ69G5FAV",
			userRole: "m2m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a valid token
			token := generateTestToken(tc.userID, tc.userRole, time.Hour)

			// Create test router
			router := gin.New()
			router.Use(JWTAuth(testJWTConfig))
			router.GET("/test", func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				require.True(t, exists)
				assert.Equal(t, tc.userID, userID)

				userRoles, exists := c.Get("user_roles")
				require.True(t, exists)
				assert.Equal(t, []string{tc.userRole}, userRoles)

				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestValidateToken(t *testing.T) {
	t.Run("Valid token", func(t *testing.T) {
		token := generateTestToken("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user", time.Hour)
		claims, err := validateToken(token, testJWTConfig)
		require.NoError(t, err)
		assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", claims.UserID)
		assert.Equal(t, []string{"user"}, claims.UserRoles)
		assert.Equal(t, testJWTConfig.Issuer, claims.Issuer)
	})

	t.Run("Expired token", func(t *testing.T) {
		token := generateExpiredToken("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user")
		_, err := validateToken(token, testJWTConfig)
		assert.Error(t, err)
	})

	t.Run("Invalid signature", func(t *testing.T) {
		token := generateTokenWithWrongSecret("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user")
		_, err := validateToken(token, testJWTConfig)
		assert.Error(t, err)
	})

	t.Run("Invalid issuer", func(t *testing.T) {
		token := generateTokenWithWrongIssuer("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user")
		_, err := validateToken(token, testJWTConfig)
		assert.Error(t, err)
	})

	t.Run("Empty issuer config - no validation", func(t *testing.T) {
		configNoIssuer := config.JWTConfig{
			Secret: testJWTConfig.Secret,
			Issuer: "",
		}

		token := generateTestToken("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "user", time.Hour)
		claims, err := validateToken(token, configNoIssuer)
		require.NoError(t, err)
		assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", claims.UserID)
	})
}

// TestValidateToken_Security covers the authentication-bypass classes that a
// missing/weak secret or a permissive algorithm check would otherwise allow.
func TestValidateToken_Security(t *testing.T) {
	// signToken signs claims with the given method and key so we can craft
	// attacker-controlled tokens.
	signToken := func(method jwt.SigningMethod, key any) string {
		claims := WayfarerClaims{
			UserID:    "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			UserRoles: []string{"superadmin"},
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    testJWTConfig.Issuer,
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		tokenString, err := jwt.NewWithClaims(method, claims).SignedString(key)
		require.NoError(t, err)
		return tokenString
	}

	t.Run("rejects empty configured secret (fail-open bypass)", func(t *testing.T) {
		// An attacker forges a token signed with the empty key, matching a
		// server whose JWT_SECRET was never set.
		emptySecretCfg := config.JWTConfig{Secret: "", Issuer: testJWTConfig.Issuer}
		token := signToken(jwt.SigningMethodHS256, []byte(""))

		_, err := validateToken(token, emptySecretCfg)
		assert.Error(t, err, "empty secret must never validate a token")
	})

	t.Run("rejects HS384 algorithm downgrade", func(t *testing.T) {
		// Same secret, different HMAC variant. Must be rejected because only
		// HS256 is pinned.
		token := signToken(jwt.SigningMethodHS384, []byte(testJWTConfig.Secret))

		_, err := validateToken(token, testJWTConfig)
		assert.Error(t, err, "non-HS256 HMAC variants must be rejected")
	})

	t.Run("rejects none algorithm", func(t *testing.T) {
		token := signToken(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType)

		_, err := validateToken(token, testJWTConfig)
		assert.Error(t, err, "unsigned (alg=none) tokens must be rejected")
	})
}
