// Package authtoken signs and verifies Wayfarer access tokens (HS256 JWTs).
//
// Verification happens on every authenticated request, so it is kept free of
// I/O: the signing secret is picked by the token's kid header, and the
// emergency revocation check is a single map lookup.
package authtoken

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// Audience is the aud claim set on every token Wayfarer issues.
const Audience = "wayfarer"

var (
	// ErrInvalidToken is returned for tokens that fail signature, expiry,
	// issuer or audience checks.
	ErrInvalidToken = errors.New("invalid token")
	// ErrRevoked is returned when the user is on the emergency revocation list.
	ErrRevoked = errors.New("token revoked")
)

// Claims are the claims carried by a Wayfarer access token.
type Claims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"` // All roles the user has
	// SessionID links the token to an auth_sessions row. Empty for legacy
	// tokens from GET /token and for m2m tokens.
	SessionID string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

var keyIDs sync.Map // secret -> kid

// KeyID derives a short, non-reversible identifier for a signing secret.
// Tokens carry it in their kid header so verification can pick the right
// secret during rotation without trying each one.
func KeyID(secret string) string {
	if kid, ok := keyIDs.Load(secret); ok {
		return kid.(string)
	}
	sum := sha256.Sum256([]byte(secret))
	kid := hex.EncodeToString(sum[:4])
	keyIDs.Store(secret, kid)
	return kid
}

// NewClaims builds access-token claims for a user session, valid for the
// configured access-token lifetime starting at now.
func NewClaims(cfg config.JWTConfig, userID, sessionID string, roles []string, now time.Time) Claims {
	return Claims{
		UserID:    userID,
		UserRoles: roles,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTokenTTL)),
		},
	}
}

// Sign signs claims with the current secret and stamps its kid.
func Sign(cfg config.JWTConfig, claims Claims) (string, error) {
	if cfg.Secret == "" {
		return "", errors.New("JWT secret not configured")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = KeyID(cfg.Secret)
	signed, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

// Parse verifies a Wayfarer access token and returns its claims.
//
// Tokens without a kid (legacy tokens and m2m tokens) are verified with the
// current secret. Tokens without an aud are accepted for the same reason; a
// token that does carry an aud must include Audience.
func Parse(tokenString string, cfg config.JWTConfig) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return secretFor(t, cfg)
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if cfg.Issuer != "" && claims.Issuer != cfg.Issuer {
		return nil, fmt.Errorf("%w: issuer", ErrInvalidToken)
	}
	if len(claims.Audience) > 0 && !slices.Contains(claims.Audience, Audience) {
		return nil, fmt.Errorf("%w: audience", ErrInvalidToken)
	}

	var issuedAt time.Time
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}
	if cfg.IsRevoked(claims.UserID, issuedAt) {
		return nil, ErrRevoked
	}
	return claims, nil
}

func secretFor(t *jwt.Token, cfg config.JWTConfig) ([]byte, error) {
	kid, _ := t.Header["kid"].(string)
	switch {
	case kid == "" || kid == KeyID(cfg.Secret):
		return []byte(cfg.Secret), nil
	case cfg.PreviousSecret != "" && kid == KeyID(cfg.PreviousSecret):
		return []byte(cfg.PreviousSecret), nil
	default:
		return nil, errors.New("unknown kid")
	}
}
