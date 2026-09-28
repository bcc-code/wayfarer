package services

import (
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
	RotateAuthSession(ctx context.Context, arg sqlc.RotateAuthSessionParams) (*sqlc.AuthSession, error)
	GetAuthSessionByPrevHash(ctx context.Context, prevHash []byte) (*sqlc.AuthSession, error)
	RevokeAuthSession(ctx context.Context, id string) error
	RevokeAuthSessionByHash(ctx context.Context, refreshTokenHash []byte) error
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

	refreshToken, hash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshExpiresAt := now.Add(s.cfg.RefreshTokenTTL)

	session, err := s.queries.CreateAuthSession(ctx, sqlc.CreateAuthSessionParams{
		ID:               ulid.NewAuthSessionID(),
		UserID:           userID,
		RefreshTokenHash: hash,
		ExpiresAt:        timestamptz(refreshExpiresAt),
		UserAgent:        optionalString(userAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create auth session: %w", err)
	}

	return s.tokenPair(ctx, session, refreshToken, refreshExpiresAt, now)
}

// Refresh exchanges a refresh token for a new token pair. The presented
// token is rotated: it stops working and the returned one replaces it.
func (s *AuthSessionService) Refresh(ctx context.Context, refreshToken, userAgent string) (*TokenPair, error) {
	if !strings.HasPrefix(refreshToken, refreshTokenPrefix) {
		return nil, ErrRefreshTokenInvalid
	}
	now := s.now()
	currentHash := hashRefreshToken(refreshToken)

	newToken, newHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshExpiresAt := now.Add(s.cfg.RefreshTokenTTL)

	session, err := s.queries.RotateAuthSession(ctx, sqlc.RotateAuthSessionParams{
		NewHash:     newHash,
		ExpiresAt:   timestamptz(refreshExpiresAt),
		UserAgent:   optionalString(userAgent),
		CurrentHash: currentHash,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, s.classifyStaleToken(ctx, currentHash, now)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to rotate auth session: %w", err)
	}

	// A cutoff on the emergency list forces a fresh login: sessions opened
	// before it can no longer be refreshed.
	if s.cfg.IsRevoked(session.UserID, session.CreatedAt.Time) {
		if err := s.queries.RevokeAuthSession(ctx, session.ID); err != nil {
			slog.Error("auth: failed to revoke session of revoked user", "session_id", session.ID, "error", err)
		}
		return nil, ErrUserRevoked
	}

	return s.tokenPair(ctx, session, newToken, refreshExpiresAt, now)
}

// classifyStaleToken works out why a refresh token didn't match a live
// session: a benign race, a reuse of a stolen token, or just unknown.
func (s *AuthSessionService) classifyStaleToken(ctx context.Context, hash []byte, now time.Time) error {
	session, err := s.queries.GetAuthSessionByPrevHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRefreshTokenInvalid
	}
	if err != nil {
		return fmt.Errorf("failed to look up rotated auth session: %w", err)
	}
	if session.RevokedAt.Valid || !session.ExpiresAt.Time.After(now) {
		return ErrRefreshTokenInvalid
	}
	if session.RotatedAt.Valid && now.Sub(session.RotatedAt.Time) <= RefreshGracePeriod {
		return ErrRefreshTokenRotated
	}

	slog.Warn("auth: refresh token reused after rotation, revoking session",
		"session_id", session.ID,
		"user_id", session.UserID,
	)
	if err := s.queries.RevokeAuthSession(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to revoke reused auth session: %w", err)
	}
	return ErrRefreshTokenReused
}

// Logout revokes the session behind a refresh token. Unknown tokens are
// ignored so logout is idempotent.
func (s *AuthSessionService) Logout(ctx context.Context, refreshToken string) error {
	if !strings.HasPrefix(refreshToken, refreshTokenPrefix) {
		return nil
	}
	if err := s.queries.RevokeAuthSessionByHash(ctx, hashRefreshToken(refreshToken)); err != nil {
		return fmt.Errorf("failed to revoke auth session: %w", err)
	}
	return nil
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

func (s *AuthSessionService) tokenPair(ctx context.Context, session *sqlc.AuthSession, refreshToken string, refreshExpiresAt, now time.Time) (*TokenPair, error) {
	roles, err := s.roles.TokenRoleNames(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	claims := authtoken.NewClaims(s.cfg, session.UserID, session.ID, roles, now)
	accessToken, err := authtoken.Sign(s.cfg, claims)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      accessToken,
		AccessExpiresAt:  claims.ExpiresAt.Time,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// newRefreshToken returns a new random refresh token and its hash.
func newRefreshToken() (string, []byte, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := refreshTokenPrefix + base64.RawURLEncoding.EncodeToString(buf)
	return token, hashRefreshToken(token), nil
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
