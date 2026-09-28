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

func newAuthSessionFixture(t *testing.T) *authSessionFixture {
	t.Helper()
	cfg := config.JWTConfig{
		Secret:          "test-secret",
		Issuer:          "wayfarer",
		AccessTokenTTL:  7 * 24 * time.Hour,
		RefreshTokenTTL: 183 * 24 * time.Hour,
	}
	sessions := mocks.NewMockAuthSessionQuerier(t)
	roles := mocks.NewMockRoleQuerier(t)
	svc := NewAuthSessionService(sessions, NewRoleService(roles, newTestCache()), cfg)
	now := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	return &authSessionFixture{service: svc, sessions: sessions, roles: roles, cfg: cfg, now: now}
}

func (f *authSessionFixture) revokeUsers(revoked map[string]time.Time) {
	f.service.cfg.RevokedUsers = revoked
}

func (f *authSessionFixture) expectRoles(roles ...string) {
	rows := make([]*sqlc.UserRole, 0, len(roles))
	for _, r := range roles {
		rows = append(rows, &sqlc.UserRole{UserID: testSessionUserID, Role: r})
	}
	f.roles.On("GetUserRoles", mock.Anything, testSessionUserID).Return(rows, nil).Once()
}

// liveSession returns a stored session whose current token is currentToken.
func (f *authSessionFixture) liveSession(currentToken string) *sqlc.AuthSession {
	return &sqlc.AuthSession{
		ID:               testSessionID,
		UserID:           testSessionUserID,
		RefreshTokenHash: hashRefreshToken(currentToken),
		ExpiresAt:        pgtype.Timestamptz{Time: f.now.Add(f.cfg.RefreshTokenTTL), Valid: true},
		CreatedAt:        pgtype.Timestamptz{Time: f.now.Add(-24 * time.Hour), Valid: true},
	}
}

func (f *authSessionFixture) expectLoad(s *sqlc.AuthSession) {
	f.sessions.On("GetAuthSessionByID", mock.Anything, testSessionID).Return(s, nil).Once()
}

// expectRetired answers whether token is a retired token of the session.
func (f *authSessionFixture) expectRetired(token string, retired bool) {
	f.sessions.On("IsRetiredAuthSessionToken", mock.Anything, sqlc.IsRetiredAuthSessionTokenParams{
		TokenHash: hashRefreshToken(token),
		SessionID: testSessionID,
	}).Return(retired, nil).Once()
}

func mustToken(t *testing.T) string {
	t.Helper()
	token, _, err := newRefreshToken(testSessionID)
	require.NoError(t, err)
	return token
}

func assertValidPair(t *testing.T, f *authSessionFixture, pair *TokenPair) {
	t.Helper()
	sessionID, ok := parseRefreshToken(pair.RefreshToken)
	require.True(t, ok, "refresh token has the expected shape")
	assert.Equal(t, testSessionID, sessionID)
	assert.Equal(t, f.now.Add(f.cfg.RefreshTokenTTL), pair.RefreshExpiresAt)
	assert.Equal(t, f.now.Add(f.cfg.AccessTokenTTL).Truncate(time.Second), pair.AccessExpiresAt)

	claims := &authtoken.Claims{}
	_, _, err := jwt.NewParser().ParseUnverified(pair.AccessToken, claims)
	require.NoError(t, err)
	assert.Equal(t, testSessionUserID, claims.UserID)
	assert.Equal(t, testSessionID, claims.SessionID)
}

func TestStartSession_CreatesSessionAndTokens(t *testing.T) {
	f := newAuthSessionFixture(t)
	ctx := context.Background()

	var stored sqlc.CreateAuthSessionParams
	f.expectRoles("ADMIN")
	f.sessions.On("CreateAuthSession", ctx, mock.MatchedBy(func(p sqlc.CreateAuthSessionParams) bool {
		stored = p
		return true
	})).Return(&sqlc.AuthSession{}, nil)

	pair, err := f.service.StartSession(ctx, testSessionUserID, "Mozilla/5.0")
	require.NoError(t, err)

	sessionID, ok := parseRefreshToken(pair.RefreshToken)
	require.True(t, ok)
	assert.Equal(t, stored.ID, sessionID, "token carries the new session's ID")
	assert.True(t, strings.HasPrefix(stored.ID, "AS"))
	assert.Equal(t, testSessionUserID, stored.UserID)
	assert.Equal(t, hashRefreshToken(pair.RefreshToken), stored.RefreshTokenHash, "only the hash is stored")
	assert.Equal(t, f.now.Add(f.cfg.RefreshTokenTTL), stored.ExpiresAt.Time)
	require.NotNil(t, stored.UserAgent)
	assert.Equal(t, "Mozilla/5.0", *stored.UserAgent)
}

func TestStartSession_RoleFailureCreatesNoSession(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.roles.On("GetUserRoles", mock.Anything, testSessionUserID).Return(nil, errors.New("db down"))

	_, err := f.service.StartSession(context.Background(), testSessionUserID, "")
	assert.Error(t, err)
	f.sessions.AssertNotCalled(t, "CreateAuthSession", mock.Anything, mock.Anything)
}

func TestStartSession_RevokedUser(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.revokeUsers(map[string]time.Time{testSessionUserID: {}})

	_, err := f.service.StartSession(context.Background(), testSessionUserID, "")
	assert.ErrorIs(t, err, ErrUserRevoked)
}

func TestStartSession_CutoffInPastAllowsNewLogin(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.revokeUsers(map[string]time.Time{testSessionUserID: f.now.Add(-time.Hour)})
	f.expectRoles()
	f.sessions.On("CreateAuthSession", mock.Anything, mock.Anything).Return(&sqlc.AuthSession{}, nil)

	_, err := f.service.StartSession(context.Background(), testSessionUserID, "")
	assert.NoError(t, err)
}

func TestRefresh_RotatesToken(t *testing.T) {
	f := newAuthSessionFixture(t)
	ctx := context.Background()
	oldToken := mustToken(t)
	f.expectLoad(f.liveSession(oldToken))
	f.expectRoles()

	var params sqlc.RotateAuthSessionParams
	f.sessions.On("RotateAuthSession", ctx, mock.MatchedBy(func(p sqlc.RotateAuthSessionParams) bool {
		params = p
		return true
	})).Return(testSessionID, nil)

	pair, err := f.service.Refresh(ctx, oldToken, "")
	require.NoError(t, err)
	assertValidPair(t, f, pair)

	assert.Equal(t, testSessionID, params.ID)
	assert.Equal(t, hashRefreshToken(oldToken), params.CurrentHash, "swap is conditional on the validated hash")
	assert.Equal(t, hashRefreshToken(pair.RefreshToken), params.NewHash)
	assert.NotEqual(t, oldToken, pair.RefreshToken)
	assert.Equal(t, f.now.Add(f.cfg.RefreshTokenTTL), params.ExpiresAt.Time, "expiry slides forward")
}

func TestRefresh_RejectsMalformedTokens(t *testing.T) {
	f := newAuthSessionFixture(t)
	valid := mustToken(t)

	for name, token := range map[string]string{
		"jwt":              "eyJhbGciOi...",
		"empty":            "",
		"wrong prefix":     "xyz_" + valid[4:],
		"truncated":        valid[:len(valid)-1],
		"too long":         valid + "A",
		"bad separator":    valid[:32] + "-" + valid[33:],
		"not a session id": "wfr_US01ARZ3NDEKTSV4RRFFQ69G5FAV" + valid[32:],
		"huge":             "wfr_" + strings.Repeat("A", 1<<20),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := f.service.Refresh(context.Background(), token, "")
			assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
		})
	}
	// None of these reach the database.
	f.sessions.AssertNotCalled(t, "GetAuthSessionByID", mock.Anything, mock.Anything)
}

func TestRefresh_UnknownSession(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.sessions.On("GetAuthSessionByID", mock.Anything, testSessionID).Return(nil, pgx.ErrNoRows)

	_, err := f.service.Refresh(context.Background(), mustToken(t), "")
	assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
}

func TestRefresh_RevokedOrExpiredSession(t *testing.T) {
	for name, mutate := range map[string]func(*sqlc.AuthSession, time.Time){
		"revoked": func(s *sqlc.AuthSession, now time.Time) {
			s.RevokedAt = pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}
		},
		"expired": func(s *sqlc.AuthSession, now time.Time) {
			s.ExpiresAt = pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := newAuthSessionFixture(t)
			token := mustToken(t)
			s := f.liveSession(token)
			mutate(s, f.now)
			f.expectLoad(s)

			_, err := f.service.Refresh(context.Background(), token, "")
			assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
			f.sessions.AssertNotCalled(t, "RotateAuthSession", mock.Anything, mock.Anything)
		})
	}
}

func TestRefresh_JustRotatedTokenWithinGracePeriod(t *testing.T) {
	f := newAuthSessionFixture(t)
	oldToken := mustToken(t)
	s := f.liveSession(mustToken(t))
	s.PrevRefreshTokenHash = hashRefreshToken(oldToken)
	s.RotatedAt = pgtype.Timestamptz{Time: f.now.Add(-5 * time.Second), Valid: true}
	f.expectLoad(s)

	_, err := f.service.Refresh(context.Background(), oldToken, "")
	assert.ErrorIs(t, err, ErrRefreshTokenRotated)
	f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
}

func TestRefresh_PreviousTokenAfterGraceRevokesSession(t *testing.T) {
	f := newAuthSessionFixture(t)
	oldToken := mustToken(t)
	s := f.liveSession(mustToken(t))
	s.PrevRefreshTokenHash = hashRefreshToken(oldToken)
	s.RotatedAt = pgtype.Timestamptz{Time: f.now.Add(-RefreshGracePeriod - time.Second), Valid: true}
	f.expectLoad(s)
	f.expectRetired(oldToken, true)
	f.sessions.On("RevokeAuthSession", mock.Anything, testSessionID).Return(nil).Once()

	_, err := f.service.Refresh(context.Background(), oldToken, "")
	assert.ErrorIs(t, err, ErrRefreshTokenReused)
}

// An attacker holding token A rotates A→B→C. The legitimate client's later
// use of A must still revoke the session, even though A is neither the
// current nor the previous hash.
func TestRefresh_OlderTokenAfterDoubleRotationRevokesSession(t *testing.T) {
	f := newAuthSessionFixture(t)
	tokenA, tokenB, tokenC := mustToken(t), mustToken(t), mustToken(t)
	s := f.liveSession(tokenC)
	s.PrevRefreshTokenHash = hashRefreshToken(tokenB)
	s.RotatedAt = pgtype.Timestamptz{Time: f.now.Add(-time.Second), Valid: true}
	f.expectLoad(s)
	f.expectRetired(tokenA, true)
	f.sessions.On("RevokeAuthSession", mock.Anything, testSessionID).Return(nil).Once()

	_, err := f.service.Refresh(context.Background(), tokenA, "")
	assert.ErrorIs(t, err, ErrRefreshTokenReused, "even within the grace window, only the immediately previous token is a benign race")
}

// The session ID in a refresh token is not secret: it is also in the access
// token's sid claim. A token fabricated around a known session ID must not
// be able to revoke that session.
func TestRefresh_FabricatedTokenDoesNotRevoke(t *testing.T) {
	f := newAuthSessionFixture(t)
	s := f.liveSession(mustToken(t))
	s.PrevRefreshTokenHash = hashRefreshToken(mustToken(t))
	s.RotatedAt = pgtype.Timestamptz{Time: f.now.Add(-time.Hour), Valid: true}
	forged := "wfr_" + testSessionID + "_" + strings.Repeat("A", 43)
	f.expectLoad(s)
	f.expectRetired(forged, false)

	_, err := f.service.Refresh(context.Background(), forged, "")
	assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
	f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
}

func TestRefresh_RetiredCheckFailure(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.expectLoad(f.liveSession(mustToken(t)))
	dbErr := errors.New("db down")
	f.sessions.On("IsRetiredAuthSessionToken", mock.Anything, mock.Anything).Return(false, dbErr)

	_, err := f.service.Refresh(context.Background(), mustToken(t), "")
	assert.ErrorIs(t, err, dbErr)
	f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
}

func TestRefresh_ConcurrentRotationLosesSwap(t *testing.T) {
	f := newAuthSessionFixture(t)
	token := mustToken(t)
	f.expectLoad(f.liveSession(token))
	f.expectRoles()
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return("", pgx.ErrNoRows)

	_, err := f.service.Refresh(context.Background(), token, "")
	assert.ErrorIs(t, err, ErrRefreshTokenRotated)
	f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
}

// A failure while preparing the new pair must not commit the rotation,
// otherwise the replacement token never reaches the client and its retry
// with the old token would later be treated as reuse.
func TestRefresh_RoleFailureDoesNotRotate(t *testing.T) {
	f := newAuthSessionFixture(t)
	token := mustToken(t)
	f.expectLoad(f.liveSession(token))
	f.roles.On("GetUserRoles", mock.Anything, testSessionUserID).Return(nil, errors.New("db down")).Once()

	_, err := f.service.Refresh(context.Background(), token, "")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrRefreshTokenInvalid)
	f.sessions.AssertNotCalled(t, "RotateAuthSession", mock.Anything, mock.Anything)

	// The retry with the same token succeeds.
	f.expectLoad(f.liveSession(token))
	f.expectRoles()
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(testSessionID, nil).Once()
	_, err = f.service.Refresh(context.Background(), token, "")
	assert.NoError(t, err)
}

func TestRefresh_EmergencyCutoffRevokesOlderSession(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.revokeUsers(map[string]time.Time{testSessionUserID: f.now.Add(-time.Hour)})
	token := mustToken(t)
	f.expectLoad(f.liveSession(token)) // created 24h ago, before the cutoff
	f.sessions.On("RevokeAuthSession", mock.Anything, testSessionID).Return(nil).Once()

	_, err := f.service.Refresh(context.Background(), token, "")
	assert.ErrorIs(t, err, ErrUserRevoked)
	f.sessions.AssertNotCalled(t, "RotateAuthSession", mock.Anything, mock.Anything)
}

func TestRefresh_EmergencyCutoffAllowsNewerSession(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.revokeUsers(map[string]time.Time{testSessionUserID: f.now.Add(-48 * time.Hour)})
	token := mustToken(t)
	f.expectLoad(f.liveSession(token))
	f.expectRoles()
	f.sessions.On("RotateAuthSession", mock.Anything, mock.Anything).Return(testSessionID, nil)

	_, err := f.service.Refresh(context.Background(), token, "")
	assert.NoError(t, err)
}

func TestRefresh_DatabaseError(t *testing.T) {
	f := newAuthSessionFixture(t)
	dbErr := errors.New("connection refused")
	f.sessions.On("GetAuthSessionByID", mock.Anything, testSessionID).Return(nil, dbErr)

	_, err := f.service.Refresh(context.Background(), mustToken(t), "")
	assert.ErrorIs(t, err, dbErr)
	assert.NotErrorIs(t, err, ErrRefreshTokenInvalid)
}

func TestLogout_CurrentToken(t *testing.T) {
	f := newAuthSessionFixture(t)
	token := mustToken(t)
	f.expectLoad(f.liveSession(token))
	f.sessions.On("RevokeAuthSession", mock.Anything, testSessionID).Return(nil).Once()

	require.NoError(t, f.service.Logout(context.Background(), token))
}

func TestLogout_RotatedToken(t *testing.T) {
	f := newAuthSessionFixture(t)
	rotated := mustToken(t)
	f.expectLoad(f.liveSession(mustToken(t)))
	f.expectRetired(rotated, true)
	f.sessions.On("RevokeAuthSession", mock.Anything, testSessionID).Return(nil).Once()

	require.NoError(t, f.service.Logout(context.Background(), rotated), "a logout racing a refresh still ends the session")
}

func TestLogout_FabricatedTokenIsIgnored(t *testing.T) {
	f := newAuthSessionFixture(t)
	forged := "wfr_" + testSessionID + "_" + strings.Repeat("A", 43)
	f.expectLoad(f.liveSession(mustToken(t)))
	f.expectRetired(forged, false)

	require.NoError(t, f.service.Logout(context.Background(), forged))
	f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
}

func TestLogout_NoOps(t *testing.T) {
	f := newAuthSessionFixture(t)

	require.NoError(t, f.service.Logout(context.Background(), "not-a-refresh-token"), "malformed")

	f.sessions.On("GetAuthSessionByID", mock.Anything, testSessionID).Return(nil, pgx.ErrNoRows).Once()
	require.NoError(t, f.service.Logout(context.Background(), mustToken(t)), "unknown session")

	token := mustToken(t)
	s := f.liveSession(token)
	s.RevokedAt = pgtype.Timestamptz{Time: f.now, Valid: true}
	f.expectLoad(s)
	require.NoError(t, f.service.Logout(context.Background(), token), "already revoked")

	f.sessions.AssertNotCalled(t, "RevokeAuthSession", mock.Anything, mock.Anything)
}

func TestRevokeUserSessions(t *testing.T) {
	f := newAuthSessionFixture(t)
	f.sessions.On("RevokeUserAuthSessions", mock.Anything, testSessionUserID).Return(int64(3), nil)

	n, err := f.service.RevokeUserSessions(context.Background(), testSessionUserID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestDeleteStaleSessions(t *testing.T) {
	f := newAuthSessionFixture(t)
	retention := 30 * 24 * time.Hour
	f.sessions.On("DeleteStaleAuthSessions", mock.Anything, pgtype.Timestamptz{Time: f.now.Add(-retention), Valid: true}).Return(int64(7), nil)

	n, err := f.service.DeleteStaleSessions(context.Background(), retention)
	require.NoError(t, err)
	assert.Equal(t, int64(7), n)
}

func TestNewRefreshToken_ShapeAndUniqueness(t *testing.T) {
	a, hashA, err := newRefreshToken(testSessionID)
	require.NoError(t, err)
	b, hashB, err := newRefreshToken(testSessionID)
	require.NoError(t, err)

	assert.NotEqual(t, a, b)
	assert.False(t, bytes.Equal(hashA, hashB))
	assert.Len(t, hashA, 32)
	assert.Len(t, a, RefreshTokenLength)
	assert.Equal(t, 76, RefreshTokenLength)

	sessionID, ok := parseRefreshToken(a)
	require.True(t, ok)
	assert.Equal(t, testSessionID, sessionID)
}
