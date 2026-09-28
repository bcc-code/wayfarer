package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSessionTestRouter(t *testing.T) (*gin.Engine, *mocks.MockAuthSessionQuerier, *mocks.MockRoleQuerier) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	sessions := mocks.NewMockAuthSessionQuerier(t)
	roles := mocks.NewMockRoleQuerier(t)
	cfg := config.JWTConfig{
		Secret:          "test-secret",
		Issuer:          "wayfarer",
		AccessTokenTTL:  7 * 24 * time.Hour,
		RefreshTokenTTL: 183 * 24 * time.Hour,
	}
	h := &AuthHandler{
		Cfg:            &config.Config{JWT: cfg},
		SessionService: services.NewAuthSessionService(sessions, services.NewRoleService(roles, newTestCache()), cfg),
	}
	r := gin.New()
	g := r.Group("/auth", middleware.MaxBodyBytes(AuthRequestBodyLimit))
	g.POST("/exchange", h.Exchange)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", h.Logout)
	return r, sessions, roles
}

const handlerTestSessionID = "AS01ARZ3NDEKTSV4RRFFQ69G5FAV"

// testRefreshToken is a well-formed refresh token for handlerTestSessionID.
var testRefreshToken = "wfr_" + handlerTestSessionID + "_" + strings.Repeat("A", 43)

func refreshBody(token string) string {
	b, _ := json.Marshal(map[string]string{"refresh_token": token})
	return string(b)
}

func liveSessionFor(token string) *sqlc.AuthSession {
	sum := sha256.Sum256([]byte(token))
	return &sqlc.AuthSession{
		ID:               handlerTestSessionID,
		UserID:           "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
		RefreshTokenHash: sum[:],
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		CreatedAt:        pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
	}
}

func doPost(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body["code"]
}

func TestRefreshHandler_Success(t *testing.T) {
	r, sessions, roles := newSessionTestRouter(t)
	sessions.On("GetAuthSessionByID", mock.Anything, handlerTestSessionID).Return(liveSessionFor(testRefreshToken), nil)
	roles.On("GetUserRoles", mock.Anything, "US01ARZ3NDEKTSV4RRFFQ69G5FAV").Return([]*sqlc.UserRole{}, nil)
	sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(handlerTestSessionID, nil)

	rec := doPost(r, "/auth/refresh", refreshBody(testRefreshToken))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var pair services.TokenPair
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pair))
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEqual(t, testRefreshToken, pair.RefreshToken)
	assert.Len(t, pair.RefreshToken, services.RefreshTokenLength)
}

func TestRefreshHandler_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(*mocks.MockAuthSessionQuerier)
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing token",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "token longer than any refresh token",
			body:       refreshBody("wfr_" + strings.Repeat("A", 200)),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed token",
			body:       refreshBody("wfr_abc"),
			wantStatus: http.StatusUnauthorized,
			wantCode:   authCodeInvalidRefreshToken,
		},
		{
			name: "unknown session",
			body: refreshBody(testRefreshToken),
			setup: func(m *mocks.MockAuthSessionQuerier) {
				m.On("GetAuthSessionByID", mock.Anything, handlerTestSessionID).Return(nil, pgx.ErrNoRows)
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   authCodeInvalidRefreshToken,
		},
		{
			name: "just rotated",
			body: refreshBody(testRefreshToken),
			setup: func(m *mocks.MockAuthSessionQuerier) {
				s := liveSessionFor("wfr_" + handlerTestSessionID + "_" + strings.Repeat("B", 43))
				sum := sha256.Sum256([]byte(testRefreshToken))
				s.PrevRefreshTokenHash = sum[:]
				s.RotatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
				m.On("GetAuthSessionByID", mock.Anything, handlerTestSessionID).Return(s, nil)
			},
			wantStatus: http.StatusConflict,
			wantCode:   authCodeTokenRotated,
		},
		{
			name: "database failure",
			body: refreshBody(testRefreshToken),
			setup: func(m *mocks.MockAuthSessionQuerier) {
				m.On("GetAuthSessionByID", mock.Anything, handlerTestSessionID).Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, sessions, _ := newSessionTestRouter(t)
			if tt.setup != nil {
				tt.setup(sessions)
			}
			rec := doPost(r, "/auth/refresh", tt.body)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantCode != "" {
				assert.Equal(t, tt.wantCode, errorCode(t, rec))
			}
		})
	}
}

func TestAuthEndpoints_RejectOversizedBodies(t *testing.T) {
	r, _, _ := newSessionTestRouter(t)
	huge := `{"token":"` + strings.Repeat("A", AuthRequestBodyLimit) + `"}`

	for _, path := range []string{"/auth/exchange", "/auth/refresh", "/auth/logout"} {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, http.StatusRequestEntityTooLarge, doPost(r, path, huge).Code)

			// Without a Content-Length the limit applies while reading.
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(huge))
			req.ContentLength = -1
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
		})
	}
}

func TestExchangeHandler_RejectsOverlongToken(t *testing.T) {
	r, _, _ := newSessionTestRouter(t)
	body := `{"token":"` + strings.Repeat("A", 8193) + `"}`
	assert.Equal(t, http.StatusBadRequest, doPost(r, "/auth/exchange", body).Code)
}

func TestLogoutHandler(t *testing.T) {
	r, sessions, _ := newSessionTestRouter(t)
	sessions.On("GetAuthSessionByID", mock.Anything, handlerTestSessionID).Return(liveSessionFor(testRefreshToken), nil).Once()
	sessions.On("RevokeAuthSession", mock.Anything, handlerTestSessionID).Return(nil).Once()

	assert.Equal(t, http.StatusNoContent, doPost(r, "/auth/logout", refreshBody(testRefreshToken)).Code)
	assert.Equal(t, http.StatusNoContent, doPost(r, "/auth/logout", refreshBody("wfr_abc")).Code, "malformed token is a no-op")
	assert.Equal(t, http.StatusBadRequest, doPost(r, "/auth/logout", `{}`).Code)
}

func TestFabricatedTokenCannotRevokeSession(t *testing.T) {
	forged := "wfr_" + handlerTestSessionID + "_" + strings.Repeat("Z", 43)
	for _, path := range []string{"/auth/refresh", "/auth/logout"} {
		t.Run(path, func(t *testing.T) {
			r, sessions, _ := newSessionTestRouter(t)
			sessions.On("GetAuthSessionByID", mock.Anything, handlerTestSessionID).Return(liveSessionFor(testRefreshToken), nil)
			sessions.On("IsRetiredAuthSessionToken", mock.Anything, mock.Anything).Return(false, nil)

			doPost(r, path, refreshBody(forged))
			sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
		})
	}
}

func TestExchangeHandler_RejectsBadInput(t *testing.T) {
	r, _, _ := newSessionTestRouter(t)

	assert.Equal(t, http.StatusBadRequest, doPost(r, "/auth/exchange", `{}`).Code)
	// No JWKS configured, so any token is invalid.
	assert.Equal(t, http.StatusUnauthorized, doPost(r, "/auth/exchange", `{"token":"eyJ.invalid.token"}`).Code)
}
