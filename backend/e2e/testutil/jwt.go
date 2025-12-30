package testutil

import (
	"strings"
	"time"

	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/golang-jwt/jwt/v5"
)

// TestJWTSecret is the secret used for test tokens
const TestJWTSecret = "test-jwt-secret-for-e2e-testing"

// TestJWTIssuer is the issuer for test tokens
const TestJWTIssuer = "wayfarer-test"

// TokenClaims matches the WayfarerClaims structure from middleware
type TokenClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

// toLowerRole converts a RoleType to lowercase for GraphQL directive matching
func toLowerRole(r services.RoleType) string {
	return strings.ToLower(string(r))
}

// GenerateToken creates a JWT token for a test user with the specified roles
func GenerateToken(userID string, roles []string) (string, error) {
	now := time.Now()

	claims := TokenClaims{
		UserID:    userID,
		UserRoles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TestJWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TestJWTSecret))
}

// GenerateUserToken creates a token with user role
func GenerateUserToken(userID string) (string, error) {
	return GenerateToken(userID, []string{toLowerRole(services.RoleUser)})
}

// GenerateAdminToken creates a token with admin role
func GenerateAdminToken(userID string) (string, error) {
	return GenerateToken(userID, []string{toLowerRole(services.RoleAdmin)})
}

// GenerateSuperAdminToken creates a token with superadmin role
func GenerateSuperAdminToken(userID string) (string, error) {
	return GenerateToken(userID, []string{toLowerRole(services.RoleSuperAdmin)})
}

// GenerateProjectAdminToken creates a token with project_admin role
func GenerateProjectAdminToken(userID string) (string, error) {
	return GenerateToken(userID, []string{toLowerRole(services.RoleProjectAdmin)})
}

// GenerateM2MToken creates a token for machine-to-machine access
func GenerateM2MToken() (string, error) {
	return GenerateToken("M2M_SERVICE", []string{toLowerRole(services.RoleM2M)})
}

// GenerateExpiredToken creates an expired token for testing auth rejection
func GenerateExpiredToken(userID string) (string, error) {
	past := time.Now().Add(-24 * time.Hour)

	claims := TokenClaims{
		UserID:    userID,
		UserRoles: []string{toLowerRole(services.RoleUser)},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TestJWTIssuer,
			IssuedAt:  jwt.NewNumericDate(past.Add(-1 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(past),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TestJWTSecret))
}

// GenerateTokenWithWrongSecret creates a token signed with wrong secret
func GenerateTokenWithWrongSecret(userID string) (string, error) {
	now := time.Now()

	claims := TokenClaims{
		UserID:    userID,
		UserRoles: []string{toLowerRole(services.RoleUser)},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TestJWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("wrong-secret"))
}
