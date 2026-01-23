package ladder_to_heaven

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecretKey = "test-secret-key-for-cryptex-token"
	testUserID    = "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testChurchID  = "CH01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testProjectID = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
)

func TestCryptexTokenHandler_FeatureDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &cryptexTokenHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       "", // Feature disabled
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/cryptex-token", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.handle(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "feature disabled", response["error"])
}

func TestCryptexTokenHandler_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &cryptexTokenHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       testSecretKey,
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/cryptex-token", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No user_id set in context = not authenticated

	handler.handle(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "authentication required", response["error"])
}

func TestCryptexTokenHandler_MissingRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &cryptexTokenHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       testSecretKey,
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/cryptex-token", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", testUserID)
	// No user_roles set in context

	handler.handle(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "insufficient permissions", response["error"])
}

func TestCryptexTokenHandler_InsufficientRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &cryptexTokenHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       testSecretKey,
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/cryptex-token", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", testUserID)
	c.Set("user_roles", []string{"user"}) // Only user role, not church_admin/admin/superadmin

	handler.handle(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "insufficient permissions", response["error"])
}

func TestCryptexTokenHandler_SuccessWithAllowedRoles(t *testing.T) {
	roles := []string{"church_admin", "admin", "superadmin"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// Create mock querier
			mockQuerier := &mockCryptexQuerierImpl{
				user: &sqlc.GetUserByIDRow{
					ID:         testUserID,
					Name:       "Test User",
					ChurchID:   testChurchID,
					MembersID:  "550e8400-e29b-41d4-a716-446655440000",
					Gender:     "M",
					Email:      "test@example.com",
					Language:   "en",
					PersonUuid: pgtype.UUID{Valid: false},
				},
				church: &sqlc.GetChurchByIDRow{
					ID:       testChurchID,
					Name:     "Test Church",
					Country:  "NO",
					Category: "L",
				},
			}

			// Create mock settings service
			mockSettings := &mockSettingsServiceImpl{
				projectID: testProjectID,
			}

			handler := &cryptexTokenHandler{
				db:              nil,
				settingsService: nil,
				secretKey:       testSecretKey,
				jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
				testQuerier:     mockQuerier,
				testSettings:    mockSettings,
			}

			req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/cryptex-token", nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", testUserID)
			c.Set("user_roles", []string{role})

			handler.handle(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tokenString, ok := response["token"].(string)
			require.True(t, ok, "Response should contain token")
			require.NotEmpty(t, tokenString, "Token should not be empty")

			// Parse and verify the token
			token, err := jwt.ParseWithClaims(tokenString, &CryptexClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(testSecretKey), nil
			})
			require.NoError(t, err)
			require.True(t, token.Valid, "Token should be valid")

			claims, ok := token.Claims.(*CryptexClaims)
			require.True(t, ok, "Claims should be of type CryptexClaims")

			// Verify claims (camelCase)
			assert.Equal(t, "Test User", claims.UserName)
			assert.Equal(t, testUserID, claims.UserID)
			assert.Equal(t, testChurchID, claims.ChurchID)
			assert.Equal(t, "Test Church", claims.ChurchName)
			assert.Equal(t, testProjectID, claims.ProjectID)

			// Verify audience
			require.Len(t, claims.Audience, 1)
			assert.Equal(t, "LADD-Cryptex", claims.Audience[0])

			// Verify issuer
			assert.Equal(t, "wayfarer", claims.Issuer)

			// Verify expiration is approximately 6 months in the future
			require.NotNil(t, claims.ExpiresAt)
			require.NotNil(t, claims.IssuedAt)

			issuedAt := claims.IssuedAt.Time
			expiresAt := claims.ExpiresAt.Time
			duration := expiresAt.Sub(issuedAt)

			// 6 months is approximately 180 days (can vary slightly depending on months)
			minDays := 175.0
			maxDays := 186.0
			actualDays := duration.Hours() / 24

			assert.True(t, actualDays >= minDays && actualDays <= maxDays,
				"Token should expire in approximately 6 months (175-186 days), got %.0f days", actualDays)
		})
	}
}

func TestHasAllowedRole(t *testing.T) {
	tests := []struct {
		name      string
		userRoles []string
		expected  bool
	}{
		{
			name:      "no roles",
			userRoles: []string{},
			expected:  false,
		},
		{
			name:      "user role only",
			userRoles: []string{"user"},
			expected:  false,
		},
		{
			name:      "church_admin role",
			userRoles: []string{"church_admin"},
			expected:  true,
		},
		{
			name:      "admin role",
			userRoles: []string{"admin"},
			expected:  true,
		},
		{
			name:      "superadmin role",
			userRoles: []string{"superadmin"},
			expected:  true,
		},
		{
			name:      "user and church_admin roles",
			userRoles: []string{"user", "church_admin"},
			expected:  true,
		},
		{
			name:      "multiple allowed roles",
			userRoles: []string{"church_admin", "admin"},
			expected:  true,
		},
		{
			name:      "unrelated roles",
			userRoles: []string{"viewer", "editor", "moderator"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasAllowedRole(tt.userRoles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCryptexClaims_JSONTags(t *testing.T) {
	claims := CryptexClaims{
		UserName:   "Test User",
		UserID:     "US123",
		ChurchID:   "CH456",
		ChurchName: "Test Church",
		ProjectID:  "PR789",
	}

	data, err := json.Marshal(claims)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify camelCase field names
	assert.Contains(t, decoded, "userName")
	assert.Contains(t, decoded, "userId")
	assert.Contains(t, decoded, "churchId")
	assert.Contains(t, decoded, "churchName")
	assert.Contains(t, decoded, "projectId")

	// Verify values
	assert.Equal(t, "Test User", decoded["userName"])
	assert.Equal(t, "US123", decoded["userId"])
	assert.Equal(t, "CH456", decoded["churchId"])
	assert.Equal(t, "Test Church", decoded["churchName"])
	assert.Equal(t, "PR789", decoded["projectId"])
}

// Mock implementations for testing

type mockCryptexQuerierImpl struct {
	user   *sqlc.GetUserByIDRow
	church *sqlc.GetChurchByIDRow
}

func (m *mockCryptexQuerierImpl) GetUserByID(_ context.Context, _ string) (*sqlc.GetUserByIDRow, error) {
	return m.user, nil
}

func (m *mockCryptexQuerierImpl) GetChurchByID(_ context.Context, _ string) (*sqlc.GetChurchByIDRow, error) {
	return m.church, nil
}

type mockSettingsServiceImpl struct {
	projectID string
}

func (m *mockSettingsServiceImpl) GetCurrentProjectID(_ context.Context) (string, error) {
	return m.projectID, nil
}
