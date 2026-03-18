package ladder_to_heaven

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
	testExcaliburSecretKey = "test-secret-key-for-excalibur-token"
	testExcaliburBaseURL   = "https://dev.excalibur.bcc.media"
)

func TestExcaliburUserURLHandler_MissingPathParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &excaliburUserURLHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       testExcaliburSecretKey,
		baseURL:         testExcaliburBaseURL,
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/excalibur-user-url", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.handle(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "path parameter is required", response["error"])
}

func TestExcaliburUserURLHandler_FeatureDisabled_MissingSecretKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &excaliburUserURLHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       "", // Feature disabled
		baseURL:         testExcaliburBaseURL,
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/excalibur-user-url?path=/callback", nil)
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

func TestExcaliburUserURLHandler_FeatureDisabled_MissingBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &excaliburUserURLHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       testExcaliburSecretKey,
		baseURL:         "", // Feature disabled - missing base URL
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/excalibur-user-url?path=/callback", nil)
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

func TestExcaliburUserURLHandler_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &excaliburUserURLHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       testExcaliburSecretKey,
		baseURL:         testExcaliburBaseURL,
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/excalibur-user-url?path=/callback", nil)
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

func TestExcaliburUserURLHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock querier
	mockQuerier := &mockExcaliburQuerierImpl{
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
	mockSettings := &mockExcaliburSettingsServiceImpl{
		projectID: testProjectID,
	}

	handler := &excaliburUserURLHandler{
		db:              nil,
		settingsService: nil,
		secretKey:       testExcaliburSecretKey,
		baseURL:         testExcaliburBaseURL,
		jwtConfig:       config.JWTConfig{Issuer: "wayfarer"},
		testQuerier:     mockQuerier,
		testSettings:    mockSettings,
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/excalibur-user-url?path=/callback", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", testUserID)

	handler.handle(c)

	// Should return 302 Found redirect
	assert.Equal(t, http.StatusFound, w.Code)

	// Get the redirect location
	userURL := w.Header().Get("Location")
	require.NotEmpty(t, userURL, "Location header should be set")

	// URL should have the correct format
	assert.True(t, strings.HasPrefix(userURL, testExcaliburBaseURL+"/callback?token="),
		"URL should start with base URL + /callback?token=")

	// Extract and parse the token from the URL
	parsedURL, err := url.Parse(userURL)
	require.NoError(t, err)
	tokenString := parsedURL.Query().Get("token")
	require.NotEmpty(t, tokenString, "Token should be present in URL")

	// Parse and verify the token
	token, err := jwt.ParseWithClaims(tokenString, &ExcaliburClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(testExcaliburSecretKey), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid, "Token should be valid")

	claims, ok := token.Claims.(*ExcaliburClaims)
	require.True(t, ok, "Claims should be of type ExcaliburClaims")

	// Verify claims (camelCase)
	assert.Equal(t, "Test User", claims.UserName)
	assert.Equal(t, testUserID, claims.UserID)
	assert.Equal(t, testChurchID, claims.ChurchID)
	assert.Equal(t, "Test Church", claims.ChurchName)
	assert.Equal(t, testProjectID, claims.ProjectID)

	// Verify audience is "excalibur"
	require.Len(t, claims.Audience, 1)
	assert.Equal(t, "excalibur", claims.Audience[0])

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
}

func TestExcaliburClaims_JSONTags(t *testing.T) {
	claims := ExcaliburClaims{
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

type mockExcaliburQuerierImpl struct {
	user   *sqlc.GetUserByIDRow
	church *sqlc.GetChurchByIDRow
}

func (m *mockExcaliburQuerierImpl) GetUserByID(_ context.Context, _ string) (*sqlc.GetUserByIDRow, error) {
	return m.user, nil
}

func (m *mockExcaliburQuerierImpl) GetChurchByID(_ context.Context, _ string) (*sqlc.GetChurchByIDRow, error) {
	return m.church, nil
}

type mockExcaliburSettingsServiceImpl struct {
	projectID string
}

func (m *mockExcaliburSettingsServiceImpl) GetCurrentProjectID(_ context.Context) (string, error) {
	return m.projectID, nil
}
