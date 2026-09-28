package services

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/authtoken"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
)

const (
	testSessionUserID = "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testSessionID     = "AS01ARZ3NDEKTSV4RRFFQ69G5FAV"
)

type authSessionFixture struct {
	service  *AuthSessionService
	sessions *mocks.MockAuthSessionQuerier
	roles    *mocks.MockRoleQuerier
	cfg      config.JWTConfig
	now      time.Time
}

func newAuthSessionFixture(t *testing.T, revoked map[string]time.Time) *authSessionFixture {
	t.Helper()
	cfg := config.JWTConfig{
		Secret:          "test-secret",
		Issuer:          "wayfarer",
		AccessTokenTTL:  7 * 24 * time.Hour,
		RefreshTokenTTL: 183 * 24 * time.Hour,
		RevokedUsers:    revoked,
	}
	sessions := mocks.NewMockAuthSessionQuerier(t)
	roles := mocks.NewMockRoleQuerier(t)
	svc := NewAuthSessionService(sessions, NewRoleService(roles, newTestCache()), cfg)
	now := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	return &authSessionFixture{service: svc, sessions: sessions, roles: roles, cfg: cfg, now: now}
}

func (f *authSessionFixture) expectRoles(roles ...string) {
	rows := make([]*sqlc.UserRole, 0, len(roles))
	for _, r := range roles {
		rows = append(rows, &sqlc.UserRole{UserID: testSessionUserID, Role: r})
	}
	f.roles.On("GetUserRoles", mock.Anything, testSessionUserID).Return(rows, nil).Once()
}

func (f *authSessionFixture) session(createdAt time.Time) *sqlc.AuthSession {
	return &sqlc.AuthSession{
		ID:        testSessionID,
		UserID:    testSessionUserID,
		ExpiresAt: pgtype.Timestamptz{Time: f.now.Add(f.cfg.RefreshTokenTTL), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
}

func assertValidPair(t *testing.T, f *authSessionFixture, pair *TokenPair) {
	t.Helper()
	assert.True(t, strings.HasPrefix(pair.RefreshToken, refreshTokenPrefix))
	assert.Equal(t, f.now.Add(f.cfg.RefreshTokenTTL), pair.RefreshExpiresAt)
	assert.Equal(t, f.now.Add(f.cfg.AccessTokenTTL).Truncate(time.Second), pair.AccessExpiresAt)

	// Parse at the fixture's clock: the claims are stamped with f.now.
	claims := &authtoken.Claims{}
	_, _, err := jwt.NewParser().ParseUnverified(pair.AccessToken, claims)
	require.NoError(t, err)
	assert.Equal(t, testSessionUserID, claims.UserID)
	assert.Equal(t, testSessionID, claims.SessionID)
}

func TestStartSession_CreatesSessionAndTokens(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	ctx := context.Background()

	var stored sqlc.CreateAuthSessionParams
	f.sessions.On("CreateAuthSession", ctx, mock.MatchedBy(func(p sqlc.CreateAuthSessionParams) bool {
		stored = p
		return true
	})).Return(f.session(f.now), nil)
	f.expectRoles("ADMIN")

	pair, err := f.service.StartSession(ctx, testSessionUserID, "Mozilla/5.0")
	require.NoError(t, err)
	assertValidPair(t, f, pair)

	assert.Equal(t, testSessionUserID, stored.UserID)
	assert.True(t, strings.HasPrefix(stored.ID, "AS"))
	assert.Equal(t, hashRefreshToken(pair.RefreshToken), stored.RefreshTokenHash, "only the hash is stored")
	assert.NotContains(t, string(stored.RefreshTokenHash), pair.RefreshToken)
	assert.Equal(t, f.now.Add(f.cfg.RefreshTokenTTL), stored.ExpiresAt.Time)
	require.NotNil(t, stored.UserAgent)
	assert.Equal(t, "Mozilla/5.0", *stored.UserAgent)
}

func TestStartSession_RevokedUser(t *testing.T) {
	f := newAuthSessionFixture(t, map[string]time.Time{testSessionUserID: {}})

	_, err := f.service.StartSession(context.Background(), testSessionUserID, "")
	assert.ErrorIs(t, err, ErrUserRevoked)
}

func TestStartSession_CutoffInPastAllowsNewLogin(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	f.service.cfg.RevokedUsers = map[string]time.Time{testSessionUserID: f.now.Add(-time.Hour)}
	f.sessions.On("CreateAuthSession", mock.Anything, mock.Anything).Return(f.session(f.now), nil)
	f.expectRoles()

	_, err := f.service.StartSession(context.Background(), testSessionUserID, "")
	assert.NoError(t, err)
}

func TestRefresh_RotatesToken(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	ctx := context.Background()
	oldToken, oldHash, err := newRefreshToken()
	require.NoError(t, err)

	var params sqlc.RotateAuthSessionParams
	f.sessions.On("RotateAuthSession", ctx, mock.MatchedBy(func(p sqlc.RotateAuthSessionParams) bool {
		params = p
		return true
	})).Return(f.session(f.now.Add(-24*time.Hour)), nil)
	f.expectRoles()

	pair, err := f.service.Refresh(ctx, oldToken, "")
	require.NoError(t, err)
	assertValidPair(t, f, pair)

	assert.Equal(t, oldHash, params.CurrentHash)
	assert.Equal(t, hashRefreshToken(pair.RefreshToken), params.NewHash)
	assert.NotEqual(t, oldToken, pair.RefreshToken)
	assert.Equal(t, f.now.Add(f.cfg.RefreshTokenTTL), params.ExpiresAt.Time, "expiry slides forward")
	assert.Nil(t, params.UserAgent)
}

func TestRefresh_RejectsMalformedToken(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	_, err := f.service.Refresh(context.Background(), "eyJhbGciOi...", "")
	assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
}

func TestRefresh_UnknownToken(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
	f.sessions.On("GetAuthSessionByPrevHash", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)

	token, _, _ := newRefreshToken()
	_, err := f.service.Refresh(context.Background(), token, "")
	assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
}

func TestRefresh_WithinGracePeriod(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	s := f.session(f.now.Add(-time.Hour))
	s.RotatedAt = pgtype.Timestamptz{Time: f.now.Add(-5 * time.Second), Valid: true}
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
	f.sessions.On("GetAuthSessionByPrevHash", mock.Anything, mock.Anything).Return(s, nil)

	token, _, _ := newRefreshToken()
	_, err := f.service.Refresh(context.Background(), token, "")
	assert.ErrorIs(t, err, ErrRefreshTokenRotated)
	f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
}

func TestRefresh_ReuseAfterGraceRevokesSession(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	s := f.session(f.now.Add(-time.Hour))
	s.RotatedAt = pgtype.Timestamptz{Time: f.now.Add(-RefreshGracePeriod - time.Second), Valid: true}
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
	f.sessions.On("GetAuthSessionByPrevHash", mock.Anything, mock.Anything).Return(s, nil)
	f.sessions.On("RevokeAuthSession", mock.Anything, testSessionID).Return(nil).Once()

	token, _, _ := newRefreshToken()
	_, err := f.service.Refresh(context.Background(), token, "")
	assert.ErrorIs(t, err, ErrRefreshTokenReused)
}

func TestRefresh_PrevHashOfRevokedOrExpiredSession(t *testing.T) {
	for name, mutate := range map[string]func(*sqlc.AuthSession, time.Time){
		"revoked": func(s *sqlc.AuthSession, now time.Time) {
			s.RevokedAt = pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}
		},
		"expired": func(s *sqlc.AuthSession, now time.Time) {
			s.ExpiresAt = pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := newAuthSessionFixture(t, nil)
			s := f.session(f.now.Add(-time.Hour))
			s.RotatedAt = pgtype.Timestamptz{Time: f.now.Add(-time.Hour), Valid: true}
			mutate(s, f.now)
			f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)
			f.sessions.On("GetAuthSessionByPrevHash", mock.Anything, mock.Anything).Return(s, nil)

			token, _, _ := newRefreshToken()
			_, err := f.service.Refresh(context.Background(), token, "")
			assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
			f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
		})
	}
}

func TestRefresh_EmergencyCutoffRevokesOlderSession(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	f.service.cfg.RevokedUsers = map[string]time.Time{testSessionUserID: f.now.Add(-time.Hour)}
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(f.session(f.now.Add(-48*time.Hour)), nil)
	f.sessions.On("RevokeAuthSession", mock.Anything, testSessionID).Return(nil).Once()

	token, _, _ := newRefreshToken()
	_, err := f.service.Refresh(context.Background(), token, "")
	assert.ErrorIs(t, err, ErrUserRevoked)
}

func TestRefresh_EmergencyCutoffAllowsNewerSession(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	f.service.cfg.RevokedUsers = map[string]time.Time{testSessionUserID: f.now.Add(-time.Hour)}
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(f.session(f.now.Add(-time.Minute)), nil)
	f.expectRoles()

	token, _, _ := newRefreshToken()
	_, err := f.service.Refresh(context.Background(), token, "")
	assert.NoError(t, err)
}

func TestRefresh_DatabaseError(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	dbErr := errors.New("connection refused")
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(nil, dbErr)

	token, _, _ := newRefreshToken()
	_, err := f.service.Refresh(context.Background(), token, "")
	assert.ErrorIs(t, err, dbErr)
	assert.NotErrorIs(t, err, ErrRefreshTokenInvalid)
}

func TestLogout(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	token, hash, _ := newRefreshToken()
	f.sessions.On("RevokeAuthSessionByHash", mock.Anything, hash).Return(nil).Once()

	require.NoError(t, f.service.Logout(context.Background(), token))
	require.NoError(t, f.service.Logout(context.Background(), "not-a-refresh-token"), "malformed token is a no-op")
}

func TestRevokeUserSessions(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	f.sessions.On("RevokeUserAuthSessions", mock.Anything, testSessionUserID).Return(int64(3), nil)

	n, err := f.service.RevokeUserSessions(context.Background(), testSessionUserID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestDeleteStaleSessions(t *testing.T) {
	f := newAuthSessionFixture(t, nil)
	retention := 30 * 24 * time.Hour
	f.sessions.On("DeleteStaleAuthSessions", mock.Anything, pgtype.Timestamptz{Time: f.now.Add(-retention), Valid: true}).Return(int64(7), nil)

	n, err := f.service.DeleteStaleSessions(context.Background(), retention)
	require.NoError(t, err)
	assert.Equal(t, int64(7), n)
}

func TestNewRefreshToken_UniqueAndHashed(t *testing.T) {
	a, hashA, err := newRefreshToken()
	require.NoError(t, err)
	b, hashB, err := newRefreshToken()
	require.NoError(t, err)

	assert.NotEqual(t, a, b)
	assert.False(t, bytes.Equal(hashA, hashB))
	assert.Len(t, hashA, 32)
	// 4-char prefix + 43 chars of unpadded base64url for 32 bytes
	assert.Len(t, a, len(refreshTokenPrefix)+43)
}
