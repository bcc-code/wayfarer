package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
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
	r.POST("/auth/exchange", h.Exchange)
	r.POST("/auth/refresh", h.Refresh)
	r.POST("/auth/logout", h.Logout)
	return r, sessions, roles
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
	sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(&sqlc.AuthSession{
		ID:        "AS01ARZ3NDEKTSV4RRFFQ69G5FAV",
		UserID:    "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil)
	roles.On("GetUserRoles", mock.Anything, "US01ARZ3NDEKTSV4RRFFQ69G5FAV").Return([]*sqlc.UserRole{}, nil)

	rec := doPost(r, "/auth/refresh", `{"refresh_token":"wfr_abc"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var pair services.TokenPair
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pair))
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEqual(t, "wfr_abc", pair.RefreshToken)
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
			name: "unknown token",
			body: `{"refresh_token":"wfr_abc"}`,
			setup: func(m *mocks.MockAuthSessionQuerier) {
				m.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
				m.On("GetAuthSessionByPrevHash", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   authCodeInvalidRefreshToken,
		},
		{
			name: "just rotated",
			body: `{"refresh_token":"wfr_abc"}`,
			setup: func(m *mocks.MockAuthSessionQuerier) {
				m.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
				m.On("GetAuthSessionByPrevHash", mock.Anything, mock.Anything).Return(&sqlc.AuthSession{
					ID:        "AS01ARZ3NDEKTSV4RRFFQ69G5FAV",
					ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					RotatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				}, nil)
			},
			wantStatus: http.StatusConflict,
			wantCode:   authCodeTokenRotated,
		},
		{
			name: "database failure",
			body: `{"refresh_token":"wfr_abc"}`,
			setup: func(m *mocks.MockAuthSessionQuerier) {
				m.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, errors.New("db down"))
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

func TestLogoutHandler(t *testing.T) {
	r, sessions, _ := newSessionTestRouter(t)
	sessions.On("RevokeAuthSessionByHash", mock.Anything, mock.Anything).Return(nil).Once()

	assert.Equal(t, http.StatusNoContent, doPost(r, "/auth/logout", `{"refresh_token":"wfr_abc"}`).Code)
	assert.Equal(t, http.StatusBadRequest, doPost(r, "/auth/logout", `{}`).Code)
}

func TestExchangeHandler_RejectsBadInput(t *testing.T) {
	r, _, _ := newSessionTestRouter(t)

	assert.Equal(t, http.StatusBadRequest, doPost(r, "/auth/exchange", `{}`).Code)
	// No JWKS configured, so any token is invalid.
	assert.Equal(t, http.StatusUnauthorized, doPost(r, "/auth/exchange", `{"token":"eyJ.invalid.token"}`).Code)
}
