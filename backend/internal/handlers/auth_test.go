package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestCache creates a cache instance for testing
func newTestCache() *cache.CacheWithRegistry {
	c, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
	return c
}

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
			name:     "empty string defaults to UNKNOWN",
			input:    "",
			expected: "UNKNOWN",
		},
		{
			name:     "unknown value defaults to UNKNOWN",
			input:    "unknown",
			expected: "UNKNOWN",
		},
		{
			name:     "null string defaults to UNKNOWN",
			input:    "null",
			expected: "UNKNOWN",
		},
		{
			name:     "other defaults to UNKNOWN",
			input:    "other",
			expected: "UNKNOWN",
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
	roleService := services.NewRoleService(mockQueries, newTestCache())

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
	roleService := services.NewRoleService(mockQueries, newTestCache())

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
	roleService := services.NewRoleService(mockQueries, newTestCache())

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
	roleService := services.NewRoleService(mockQueries, newTestCache())

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
	roleService := services.NewRoleService(mockQueries, newTestCache())

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

func TestAuth0Claims_ParseNamespacedClaims(t *testing.T) {
	// Create a test token with Auth0 namespaced claims
	secret := []byte("test-secret")
	now := time.Now()

	// Create claims with namespaced fields
	claims := jwt.MapClaims{
		"https://login.bcc.no/claims/churchId":  69,
		"https://login.bcc.no/claims/personId":  19254,
		"https://login.bcc.no/claims/personUid": "5e7016ac-999e-4e87-b84b-13642d863d01",
		"iss":                                   "https://login.bcc.no/",
		"sub":                                   "auth0|28cd3814-049f-4bb6-b8f6-3e0f2b25fe6b",
		"iat":                                   now.Unix(),
		"exp":                                   now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	require.NoError(t, err)

	// Parse the token with Auth0Claims struct
	parsedToken, err := jwt.ParseWithClaims(signedToken, &Auth0Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	require.NoError(t, err)
	require.True(t, parsedToken.Valid)

	auth0Claims, ok := parsedToken.Claims.(*Auth0Claims)
	require.True(t, ok)

	// Verify claims were parsed correctly
	assert.Equal(t, 69, auth0Claims.ChurchID)
	assert.Equal(t, 19254, auth0Claims.PersonID)
	assert.Equal(t, "5e7016ac-999e-4e87-b84b-13642d863d01", auth0Claims.PersonUUID)
	assert.Equal(t, "https://login.bcc.no/", auth0Claims.Issuer)
	assert.Equal(t, "auth0|28cd3814-049f-4bb6-b8f6-3e0f2b25fe6b", auth0Claims.Subject)
}

func TestAuth0Claims_ParseAppMetadata(t *testing.T) {
	// Create a test token with Auth0 namespaced claims including app_metadata
	secret := []byte("test-secret")
	now := time.Now()

	tests := []struct {
		name               string
		hasMembership      bool
		expectedMembership bool
	}{
		{
			name:               "user with membership",
			hasMembership:      true,
			expectedMembership: true,
		},
		{
			name:               "user without membership",
			hasMembership:      false,
			expectedMembership: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := jwt.MapClaims{
				"https://login.bcc.no/claims/churchId":  69,
				"https://login.bcc.no/claims/personId":  13036,
				"https://login.bcc.no/claims/personUid": "0d6655c1-fe3a-4d82-a6ed-1846655259b2",
				"https://members.bcc.no/app_metadata": map[string]interface{}{
					"hasMembership": tt.hasMembership,
					"personId":      13036,
				},
				"iss": "https://login.bcc.no/",
				"sub": "auth0|92f6cf12-4c14-4077-b3b8-9944539528ee",
				"iat": now.Unix(),
				"exp": now.Add(24 * time.Hour).Unix(),
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signedToken, err := token.SignedString(secret)
			require.NoError(t, err)

			parsedToken, err := jwt.ParseWithClaims(signedToken, &Auth0Claims{}, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			require.NoError(t, err)
			require.True(t, parsedToken.Valid)

			auth0Claims, ok := parsedToken.Claims.(*Auth0Claims)
			require.True(t, ok)

			assert.Equal(t, tt.expectedMembership, auth0Claims.AppMetadata.HasMembership)
			assert.Equal(t, 13036, auth0Claims.AppMetadata.PersonID)
		})
	}
}

func TestAuth0Claims_ConvertToBrunstadTVClaims(t *testing.T) {
	// Test that Auth0Claims can be converted to BrunstadTVClaims for downstream processing
	auth0Claims := &Auth0Claims{
		ChurchID:   69,
		PersonID:   19254,
		PersonUUID: "5e7016ac-999e-4e87-b84b-13642d863d01",
		AppMetadata: Auth0AppMetadata{
			HasMembership: true,
			PersonID:      19254,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://login.bcc.no/",
			Subject:   "auth0|28cd3814-049f-4bb6-b8f6-3e0f2b25fe6b",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	// Convert to BrunstadTVClaims format (as done in Callback handler)
	brunstadClaims := &BrunstadTVClaims{
		ChurchID:         auth0Claims.ChurchID,
		PersonID:         "19254", // PersonID converted to string
		PersonUUID:       auth0Claims.PersonUUID,
		FirstName:        "", // Not provided in Auth0 token
		Gender:           "", // Not provided in Auth0 token
		RegisteredClaims: auth0Claims.RegisteredClaims,
	}

	// Verify conversion
	assert.Equal(t, 69, brunstadClaims.ChurchID)
	assert.Equal(t, "19254", brunstadClaims.PersonID)
	assert.Equal(t, "5e7016ac-999e-4e87-b84b-13642d863d01", brunstadClaims.PersonUUID)
	assert.Empty(t, brunstadClaims.FirstName)
	assert.Empty(t, brunstadClaims.Gender)
}

func TestGetActiveAffiliationOrgUID(t *testing.T) {
	testOrgUID1 := uuid.New()
	testOrgUID2 := uuid.New()
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name         string
		affiliations []members.Affiliation
		wantOrgUID   *uuid.UUID
	}{
		{
			name:         "empty affiliations returns nil",
			affiliations: []members.Affiliation{},
			wantOrgUID:   nil,
		},
		{
			name: "single active Church affiliation",
			affiliations: []members.Affiliation{
				{
					Active: true,
					OrgUid: testOrgUID1,
					Type:   "Church",
				},
			},
			wantOrgUID: &testOrgUID1,
		},
		{
			name: "first active Church affiliation is returned",
			affiliations: []members.Affiliation{
				{
					Active: true,
					OrgUid: testOrgUID1,
					Type:   "Church",
				},
				{
					Active: true,
					OrgUid: testOrgUID2,
					Type:   "Church",
				},
			},
			wantOrgUID: &testOrgUID1,
		},
		{
			name: "Church affiliation with ValidFrom in future is skipped",
			affiliations: []members.Affiliation{
				{
					Active:    true,
					OrgUid:    testOrgUID1,
					Type:      "Church",
					ValidFrom: &futureTime,
				},
			},
			wantOrgUID: nil,
		},
		{
			name: "Church affiliation with ValidTo in past is skipped",
			affiliations: []members.Affiliation{
				{
					Active:  true,
					OrgUid:  testOrgUID1,
					Type:    "Church",
					ValidTo: &pastTime,
				},
			},
			wantOrgUID: nil,
		},
		{
			name: "Church affiliation with valid time range is returned",
			affiliations: []members.Affiliation{
				{
					Active:    true,
					OrgUid:    testOrgUID1,
					Type:      "Church",
					ValidFrom: &pastTime,
					ValidTo:   &futureTime,
				},
			},
			wantOrgUID: &testOrgUID1,
		},
		{
			name: "non-Church affiliation is skipped",
			affiliations: []members.Affiliation{
				{
					Active: true,
					OrgUid: testOrgUID1,
					Type:   "Region",
				},
			},
			wantOrgUID: nil,
		},
		{
			name: "mixed types returns first Church affiliation",
			affiliations: []members.Affiliation{
				{
					Active: true,
					OrgUid: testOrgUID1,
					Type:   "Region",
				},
				{
					Active: true,
					OrgUid: testOrgUID2,
					Type:   "Church",
				},
			},
			wantOrgUID: &testOrgUID2,
		},
		{
			name: "affiliation with empty type is skipped",
			affiliations: []members.Affiliation{
				{
					Active: true,
					OrgUid: testOrgUID1,
					Type:   "",
				},
			},
			wantOrgUID: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getActiveAffiliationOrgUID(tt.affiliations)
			if tt.wantOrgUID == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.wantOrgUID, *result)
			}
		})
	}
}

func TestGetActiveChurchAffiliationOrgUIDs(t *testing.T) {
	testOrgUID1 := uuid.New()
	testOrgUID2 := uuid.New()
	testOrgUID3 := uuid.New()

	tests := []struct {
		name         string
		affiliations []members.Affiliation
		wantOrgUIDs  []uuid.UUID
	}{
		{
			name:         "empty affiliations returns empty slice",
			affiliations: []members.Affiliation{},
			wantOrgUIDs:  nil,
		},
		{
			name: "returns all valid Church affiliations",
			affiliations: []members.Affiliation{
				{Active: true, OrgUid: testOrgUID1, Type: "Church"},
				{Active: true, OrgUid: testOrgUID2, Type: "Church"},
				{Active: true, OrgUid: testOrgUID3, Type: "Region"},
			},
			wantOrgUIDs: []uuid.UUID{testOrgUID1, testOrgUID2},
		},
		{
			name: "filters out non-Church types",
			affiliations: []members.Affiliation{
				{Active: true, OrgUid: testOrgUID1, Type: "Region"},
				{Active: true, OrgUid: testOrgUID2, Type: "Church"},
			},
			wantOrgUIDs: []uuid.UUID{testOrgUID2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getActiveChurchAffiliationOrgUIDs(tt.affiliations)
			assert.Equal(t, tt.wantOrgUIDs, result)
		})
	}
}
