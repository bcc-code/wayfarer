package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeGender(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "MALE uppercase",
			input:    "MALE",
			expected: "MALE",
		},
		{
			name:     "male lowercase",
			input:    "male",
			expected: "MALE",
		},
		{
			name:     "M uppercase",
			input:    "M",
			expected: "MALE",
		},
		{
			name:     "m lowercase",
			input:    "m",
			expected: "MALE",
		},
		{
			name:     "FEMALE uppercase",
			input:    "FEMALE",
			expected: "FEMALE",
		},
		{
			name:     "female lowercase",
			input:    "female",
			expected: "FEMALE",
		},
		{
			name:     "F uppercase",
			input:    "F",
			expected: "FEMALE",
		},
		{
			name:     "f lowercase",
			input:    "f",
			expected: "FEMALE",
		},
		{
			name:     "Male mixed case",
			input:    "Male",
			expected: "MALE",
		},
		{
			name:     "Female mixed case",
			input:    "Female",
			expected: "FEMALE",
		},
		{
			name:     "with leading/trailing spaces",
			input:    "  male  ",
			expected: "MALE",
		},
		{
			name:     "empty string defaults to MALE",
			input:    "",
			expected: "MALE",
		},
		{
			name:     "unknown value defaults to MALE",
			input:    "unknown",
			expected: "MALE",
		},
		{
			name:     "null string defaults to MALE",
			input:    "null",
			expected: "MALE",
		},
		{
			name:     "other defaults to MALE",
			input:    "other",
			expected: "MALE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeGender(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateWayfarerToken(t *testing.T) {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-jwt-signing",
			Issuer: "wayfarer-test",
		},
	}

	mockQueries := mocks.NewMockRoleQuerier(t)
	roleService := services.NewRoleService(mockQueries)

	handler := &AuthHandler{
		Cfg:         testConfig,
		RoleService: roleService,
	}

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "valid user ID",
			userID:  "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantErr: false,
		},
		{
			name:    "another valid user ID",
			userID:  "US26CHARACTERS1234567890ABC",
			wantErr: false,
		},
		{
			name:    "empty user ID still generates token",
			userID:  "",
			wantErr: false,
		},
		{
			name:    "short user ID",
			userID:  "US123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the GetUserRoles call to return a default user role
			mockQueries.On("GetUserRoles", context.Background(), tt.userID).Return([]*sqlc.UserRole{
				{
					ID:     "UR01ARZ3NDEKTSV4RRFFQ69G5FAV",
					UserID: tt.userID,
					Role:   "USER",
				},
			}, nil).Maybe()

			token, err := handler.generateWayfarerToken(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)

				// Parse token and verify claims
				parsedToken, err := jwt.ParseWithClaims(token, &WayfarerClaims{}, func(token *jwt.Token) (interface{}, error) {
					// Verify signing method is HS256
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					return []byte(testConfig.JWT.Secret), nil
				})
				require.NoError(t, err)
				require.True(t, parsedToken.Valid)

				claims, ok := parsedToken.Claims.(*WayfarerClaims)
				require.True(t, ok)

				// Verify user ID
				assert.Equal(t, tt.userID, claims.UserID)

				// Verify roles array exists and has at least one role
				assert.NotEmpty(t, claims.UserRoles)

				// Verify issuer
				assert.Equal(t, testConfig.JWT.Issuer, claims.Issuer)

				// Verify expiration (should be ~24 hours from now)
				expiresAt := claims.ExpiresAt.Time
				now := time.Now()
				expectedExpiry := now.Add(24 * time.Hour)

				// Allow 5 second tolerance for test execution time
				assert.WithinDuration(t, expectedExpiry, expiresAt, 5*time.Second)

				// Verify issued at time is recent
				issuedAt := claims.IssuedAt.Time
				assert.WithinDuration(t, now, issuedAt, 5*time.Second)

				// Verify expiration is after issued at
				assert.True(t, expiresAt.After(issuedAt), "expiration should be after issued at")
			}
		})
	}
}

func TestGenerateWayfarerToken_SigningMethod(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	roleService := services.NewRoleService(mockQueries)

	handler := &AuthHandler{
		Cfg: &config.Config{
			JWT: config.JWTConfig{
				Secret: "test-secret",
				Issuer: "test-issuer",
			},
		},
		RoleService: roleService,
	}

	mockQueries.On("GetUserRoles", context.Background(), "US123").Return([]*sqlc.UserRole{
		{ID: "UR01", UserID: "US123", Role: "USER"},
	}, nil)

	token, err := handler.generateWayfarerToken("US123")
	require.NoError(t, err)

	// Parse without validating to check the header
	parsedToken, _, err := jwt.NewParser().ParseUnverified(token, &WayfarerClaims{})
	require.NoError(t, err)

	// Verify signing method is HS256
	assert.Equal(t, "HS256", parsedToken.Method.Alg())
}

func TestGenerateWayfarerToken_InvalidSignature(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	roleService := services.NewRoleService(mockQueries)

	handler := &AuthHandler{
		Cfg: &config.Config{
			JWT: config.JWTConfig{
				Secret: "original-secret",
				Issuer: "test-issuer",
			},
		},
		RoleService: roleService,
	}

	mockQueries.On("GetUserRoles", context.Background(), "US123").Return([]*sqlc.UserRole{
		{ID: "UR01", UserID: "US123", Role: "USER"},
	}, nil)

	token, err := handler.generateWayfarerToken("US123")
	require.NoError(t, err)

	// Try to parse with wrong secret
	_, err = jwt.ParseWithClaims(token, &WayfarerClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
}

func TestGenerateWayfarerToken_TokenStructure(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	roleService := services.NewRoleService(mockQueries)

	handler := &AuthHandler{
		Cfg: &config.Config{
			JWT: config.JWTConfig{
				Secret: "test-secret",
				Issuer: "test-issuer",
			},
		},
		RoleService: roleService,
	}

	mockQueries.On("GetUserRoles", context.Background(), "US123").Return([]*sqlc.UserRole{
		{ID: "UR01", UserID: "US123", Role: "USER"},
	}, nil)

	token, err := handler.generateWayfarerToken("US123")
	require.NoError(t, err)

	// Token should have 3 parts (header.payload.signature)
	parts := splitToken(token)
	assert.Len(t, parts, 3, "JWT should have 3 parts")
}

func TestGenerateWayfarerToken_DifferentUsersDifferentTokens(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	roleService := services.NewRoleService(mockQueries)

	handler := &AuthHandler{
		Cfg: &config.Config{
			JWT: config.JWTConfig{
				Secret: "test-secret",
				Issuer: "test-issuer",
			},
		},
		RoleService: roleService,
	}

	mockQueries.On("GetUserRoles", context.Background(), "US111").Return([]*sqlc.UserRole{
		{ID: "UR01", UserID: "US111", Role: "USER"},
	}, nil)

	mockQueries.On("GetUserRoles", context.Background(), "US222").Return([]*sqlc.UserRole{
		{ID: "UR02", UserID: "US222", Role: "USER"},
	}, nil)

	token1, err := handler.generateWayfarerToken("US111")
	require.NoError(t, err)

	token2, err := handler.generateWayfarerToken("US222")
	require.NoError(t, err)

	// Tokens for different users should be different
	assert.NotEqual(t, token1, token2)

	// Parse both tokens
	claims1 := &WayfarerClaims{}
	_, err = jwt.ParseWithClaims(token1, claims1, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	require.NoError(t, err)

	claims2 := &WayfarerClaims{}
	_, err = jwt.ParseWithClaims(token2, claims2, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	require.NoError(t, err)

	// User IDs should be different
	assert.Equal(t, "US111", claims1.UserID)
	assert.Equal(t, "US222", claims2.UserID)

	// But both should have roles
	assert.NotEmpty(t, claims1.UserRoles)
	assert.NotEmpty(t, claims2.UserRoles)
}

// Helper function to split JWT token into parts
func splitToken(token string) []string {
	parts := []string{}
	start := 0
	for i, c := range token {
		if c == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	if start < len(token) {
		parts = append(parts, token[start:])
	}
	return parts
}
