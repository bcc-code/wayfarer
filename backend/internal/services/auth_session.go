package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/bcc-media/wayfarer/internal/authtoken"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
)

const (
	// refreshTokenPrefix marks Wayfarer refresh tokens so they are easy to
	// recognise in logs and secret scanners.
	refreshTokenPrefix = "wfr_"
	// refreshTokenBytes is the amount of randomness in a refresh token.
	refreshTokenBytes = 32
	// sessionIDLength is the length of the session ID embedded in a token.
	sessionIDLength = 28
	// RefreshTokenLength is the exact length of a refresh token:
	// "wfr_" + session ID + "_" + unpadded base64url of refreshTokenBytes.
	RefreshTokenLength = len(refreshTokenPrefix) + sessionIDLength + 1 + (refreshTokenBytes*8+5)/6
	// RefreshGracePeriod is how long a just-rotated refresh token is still
	// recognised as a benign race (e.g. two tabs refreshing at once) instead
	// of as token theft.
	RefreshGracePeriod = 30 * time.Second
)

var (
	// ErrRefreshTokenInvalid means the refresh token is unknown, expired or
	// revoked. The client must sign in again.
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
	// ErrRefreshTokenRotated means the token was rotated moments ago by
	// another request. The client should pick up the tokens that request
	// stored instead of signing in again.
	ErrRefreshTokenRotated = errors.New("refresh token already rotated")
	// ErrRefreshTokenReused means an old refresh token was presented after
	// the grace period. The session has been revoked.
	ErrRefreshTokenReused = errors.New("refresh token reused")
	// ErrUserRevoked means the user is on the emergency revocation list.
	ErrUserRevoked = errors.New("user revoked")
)

// AuthSessionQuerier defines the database operations needed for auth sessions
type AuthSessionQuerier interface {
	CreateAuthSession(ctx context.Context, arg sqlc.CreateAuthSessionParams) (*sqlc.AuthSession, error)
	RotateAuthSession(ctx context.Context, arg sqlc.RotateAuthSessionParams) (string, error)
	GetAuthSessionByID(ctx context.Context, id string) (*sqlc.AuthSession, error)
	IsRetiredAuthSessionToken(ctx context.Context, arg sqlc.IsRetiredAuthSessionTokenParams) (bool, error)
	RevokeAuthSession(ctx context.Context, id string) error
	RevokeUserAuthSessions(ctx context.Context, userID string) (int64, error)
	DeleteStaleAuthSessions(ctx context.Context, cutoff pgtype.Timestamptz) (int64, error)
}

// TokenPair is what a successful login or refresh returns to the client.
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// AuthSessionService issues Wayfarer access/refresh token pairs and manages
// the auth_sessions table behind them.
type AuthSessionService struct {
	queries AuthSessionQuerier
	roles   *RoleService
	cfg     config.JWTConfig
	now     func() time.Time
}

// NewAuthSessionService creates a new auth session service
func NewAuthSessionService(queries AuthSessionQuerier, roles *RoleService, cfg config.JWTConfig) *AuthSessionService {
	return &AuthSessionService{
		queries: queries,
		roles:   roles,
		cfg:     cfg,
		now:     time.Now,
	}
}

// StartSession opens a new session for a user who has just authenticated
// with an external identity provider.
func (s *AuthSessionService) StartSession(ctx context.Context, userID, userAgent string) (*TokenPair, error) {
	now := s.now()
	if s.cfg.IsRevoked(userID, now) {
		return nil, ErrUserRevoked
	}

	sessionID := ulid.NewAuthSessionID()
	refreshToken, hash, err := newRefreshToken(sessionID)
	if err != nil {
		return nil, err
	}
	refreshExpiresAt := now.Add(s.cfg.RefreshTokenTTL)

	// Prepare the access token before writing anything, so a failure leaves
	// no orphaned session behind.
	accessToken, accessExpiresAt, err := s.accessToken(ctx, userID, sessionID, now)
	if err != nil {
		return nil, err
	}

	_, err = s.queries.CreateAuthSession(ctx, sqlc.CreateAuthSessionParams{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hash,
		ExpiresAt:        timestamptz(refreshExpiresAt),
		UserAgent:        optionalString(userAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create auth session: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// Refresh exchanges a refresh token for a new token pair. The presented
// token is rotated: it stops working and the returned one replaces it.
//
// Everything that can fail (role lookup, signing) happens before the
// rotation is committed, so an error leaves the presented token valid and
// the client can simply retry.
func (s *AuthSessionService) Refresh(ctx context.Context, refreshToken, userAgent string) (*TokenPair, error) {
	sessionID, ok := parseRefreshToken(refreshToken)
	if !ok {
		return nil, ErrRefreshTokenInvalid
	}
	now := s.now()
	presentedHash := hashRefreshToken(refreshToken)

	session, err := s.queries.GetAuthSessionByID(ctx, sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRefreshTokenInvalid
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load auth session: %w", err)
	}
	if session.RevokedAt.Valid || !session.ExpiresAt.Time.After(now) {
		return nil, ErrRefreshTokenInvalid
	}
	if !bytes.Equal(session.RefreshTokenHash, presentedHash) {
		return nil, s.handleStaleToken(ctx, session, presentedHash, now)
	}

	// A cutoff on the emergency list forces a fresh login: sessions opened
	// before it can no longer be refreshed.
	if s.cfg.IsRevoked(session.UserID, session.CreatedAt.Time) {
		s.revoke(ctx, session.ID, "emergency revocation list")
		return nil, ErrUserRevoked
	}

	accessToken, accessExpiresAt, err := s.accessToken(ctx, session.UserID, session.ID, now)
	if err != nil {
		return nil, err
	}
	newToken, newHash, err := newRefreshToken(session.ID)
	if err != nil {
		return nil, err
	}
	refreshExpiresAt := now.Add(s.cfg.RefreshTokenTTL)

	_, err = s.queries.RotateAuthSession(ctx, sqlc.RotateAuthSessionParams{
		NewHash:     newHash,
		ExpiresAt:   timestamptz(refreshExpiresAt),
		UserAgent:   optionalString(userAgent),
		ID:          session.ID,
		CurrentHash: presentedHash,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		// A concurrent request rotated (or revoked) the session between our
		// read and the swap. Same situation as the grace-period race.
		return nil, ErrRefreshTokenRotated
	}
	if err != nil {
		return nil, fmt.Errorf("failed to rotate auth session: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshToken:     newToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// handleStaleToken deals with a token that names a live session but is not
// its current one. The session ID in a token is not secret (it is also in
// access-token claims), so a stale token only counts as evidence of theft
// once its hash proves the server really issued it:
//   - the token rotated out moments ago is a benign race (two tabs);
//   - any other retired token of the session means a copy is in someone
//     else's hands, so the whole session is revoked;
//   - anything else is forged or garbage and is rejected without side effects.
func (s *AuthSessionService) handleStaleToken(ctx context.Context, session *sqlc.AuthSession, presentedHash []byte, now time.Time) error {
	if bytes.Equal(session.PrevRefreshTokenHash, presentedHash) &&
		session.RotatedAt.Valid && now.Sub(session.RotatedAt.Time) <= RefreshGracePeriod {
		return ErrRefreshTokenRotated
	}

	issued, err := s.wasIssued(ctx, session, presentedHash)
	if err != nil {
		return err
	}
	if !issued {
		return ErrRefreshTokenInvalid
	}

	slog.Warn("auth: retired refresh token reused, revoking session",
		"session_id", session.ID,
		"user_id", session.UserID,
	)
	if err := s.queries.RevokeAuthSession(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to revoke reused auth session: %w", err)
	}
	return ErrRefreshTokenReused
}

// wasIssued reports whether hash is the session's current refresh token or
// one it has rotated out.
func (s *AuthSessionService) wasIssued(ctx context.Context, session *sqlc.AuthSession, hash []byte) (bool, error) {
	if bytes.Equal(session.RefreshTokenHash, hash) {
		return true, nil
	}
	retired, err := s.queries.IsRetiredAuthSessionToken(ctx, sqlc.IsRetiredAuthSessionTokenParams{
		TokenHash: hash,
		SessionID: session.ID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check retired refresh token: %w", err)
	}
	return retired, nil
}

// Logout revokes the session a refresh token belongs to. Any token the
// session was really issued works, including one rotated moments ago, so a
// logout racing a refresh still ends the session. Unknown or forged tokens
// are ignored, which also keeps logout idempotent.
func (s *AuthSessionService) Logout(ctx context.Context, refreshToken string) error {
	sessionID, ok := parseRefreshToken(refreshToken)
	if !ok {
		return nil
	}
	session, err := s.queries.GetAuthSessionByID(ctx, sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to load auth session: %w", err)
	}
	if session.RevokedAt.Valid {
		return nil
	}

	issued, err := s.wasIssued(ctx, session, hashRefreshToken(refreshToken))
	if err != nil {
		return err
	}
	if !issued {
		slog.Warn("auth: logout with a token the session never issued, ignoring", "session_id", session.ID)
		return nil
	}
	if err := s.queries.RevokeAuthSession(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to revoke auth session: %w", err)
	}
	return nil
}

func (s *AuthSessionService) revoke(ctx context.Context, sessionID, reason string) {
	if err := s.queries.RevokeAuthSession(ctx, sessionID); err != nil {
		slog.Error("auth: failed to revoke session", "session_id", sessionID, "reason", reason, "error", err)
	}
}

// RevokeUserSessions revokes every session of a user. Access tokens already
// issued stay valid until they expire; use AUTH_REVOKED_USERS to cut them off.
func (s *AuthSessionService) RevokeUserSessions(ctx context.Context, userID string) (int64, error) {
	n, err := s.queries.RevokeUserAuthSessions(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to revoke user auth sessions: %w", err)
	}
	return n, nil
}

// DeleteStaleSessions removes sessions that expired or were revoked more
// than retention ago.
func (s *AuthSessionService) DeleteStaleSessions(ctx context.Context, retention time.Duration) (int64, error) {
	n, err := s.queries.DeleteStaleAuthSessions(ctx, timestamptz(s.now().Add(-retention)))
	if err != nil {
		return 0, fmt.Errorf("failed to delete stale auth sessions: %w", err)
	}
	return n, nil
}

// accessToken signs an access token for a session. It loads the user's
// roles, so it can fail and must run before any state is committed.
func (s *AuthSessionService) accessToken(ctx context.Context, userID, sessionID string, now time.Time) (string, time.Time, error) {
	roles, err := s.roles.TokenRoleNames(ctx, userID)
	if err != nil {
		return "", time.Time{}, err
	}
	claims := authtoken.NewClaims(s.cfg, userID, sessionID, roles, now)
	signed, err := authtoken.Sign(s.cfg, claims)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, claims.ExpiresAt.Time, nil
}

// newRefreshToken returns a new random refresh token for a session and its
// hash. The session ID lets refresh and logout find the session by primary
// key and recognise every older token of it.
func newRefreshToken(sessionID string) (string, []byte, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := refreshTokenPrefix + sessionID + "_" + base64.RawURLEncoding.EncodeToString(buf)
	return token, hashRefreshToken(token), nil
}

// parseRefreshToken validates the shape of a refresh token and returns the
// session ID it carries.
func parseRefreshToken(token string) (string, bool) {
	if len(token) != RefreshTokenLength || !strings.HasPrefix(token, refreshTokenPrefix) {
		return "", false
	}
	rest := token[len(refreshTokenPrefix):]
	sessionID := rest[:sessionIDLength]
	if rest[sessionIDLength] != '_' || !ulid.IsAuthSessionID(sessionID) {
		return "", false
	}
	return sessionID, true
}

// hashRefreshToken hashes a refresh token for storage. SHA-256 is enough
// because the token is 256 bits of randomness; a slow password hash would
// only add CPU cost to every refresh.
func hashRefreshToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
