package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/loadtestauth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

// auth0User holds the DB fields needed to mint a simulated Auth0 token.
type auth0User struct {
	UserID           string
	PersonID         int
	PersonUUID       string
	ChurchExternalID int32
}

// resolveAuth0Key parses the shared load-test key (PEM or base64 PEM), or
// generates a fresh one when keySpec is empty. The returned string is the
// base64 PEM to put in the server's AUTH0_LOADTEST_PRIVATE_KEY.
func resolveAuth0Key(keySpec string) (*rsa.PrivateKey, string, error) {
	if keySpec != "" {
		key, err := loadtestauth.ParsePrivateKey(keySpec)
		if err != nil {
			return nil, "", err
		}
		encoded, err := loadtestauth.EncodePrivateKey(key)
		if err != nil {
			return nil, "", err
		}
		return key, encoded, nil
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate RSA key: %w", err)
	}
	encoded, err := loadtestauth.EncodePrivateKey(key)
	if err != nil {
		return nil, "", err
	}
	return key, encoded, nil
}

// mintAuth0Token signs an RS256 JWT shaped like a login.bcc.no token so the
// server's /token callback accepts it when the server holds the same key via
// AUTH0_LOADTEST_PRIVATE_KEY. Claim names mirror handlers.Auth0Claims.
func mintAuth0Token(key *rsa.PrivateKey, issuer string, u auth0User, now time.Time, validity time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"iss":                                   issuer,
		"iat":                                   jwt.NewNumericDate(now),
		"exp":                                   jwt.NewNumericDate(now.Add(validity)),
		"https://login.bcc.no/claims/churchId":  u.ChurchExternalID,
		"https://login.bcc.no/claims/personId":  u.PersonID,
		"https://login.bcc.no/claims/personUid": u.PersonUUID,
		"https://members.bcc.no/app_metadata": map[string]any{
			"hasMembership": true,
			"personId":      u.PersonID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = loadtestauth.KeyID
	return token.SignedString(key)
}

// generateAuth0Tokens selects up to count users eligible for the auth-dance
// scenario (numeric members_id, person_uuid set, church with an external_id so
// the callback's church lookup succeeds without the Members API) and mints an
// RS256 token for each with the given key.
func generateAuth0Tokens(ctx context.Context, db *database.DB, count int, issuer string, key *rsa.PrivateKey, now time.Time, validity time.Duration) ([]UserToken, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT u.id, u.members_id, u.person_uuid::text, c.external_id
		FROM users u
		JOIN churches c ON c.id = u.church_id
		WHERE c.external_id IS NOT NULL AND u.person_uuid IS NOT NULL
		ORDER BY u.id
		LIMIT @limit`, pgx.NamedArgs{"limit": count})
	if err != nil {
		return nil, fmt.Errorf("failed to query auth0-eligible users: %w", err)
	}
	defer rows.Close()

	var users []auth0User
	for rows.Next() {
		var (
			userID, membersID, personUUID string
			churchExternalID              int32
		)
		if err := rows.Scan(&userID, &membersID, &personUUID, &churchExternalID); err != nil {
			return nil, fmt.Errorf("failed to scan auth0-eligible user: %w", err)
		}
		personID, convErr := strconv.Atoi(membersID)
		if convErr != nil {
			slog.Warn("Skipping user with non-numeric members_id for Auth0 token", "userId", userID, "membersId", membersID)
			continue
		}
		users = append(users, auth0User{
			UserID:           userID,
			PersonID:         personID,
			PersonUUID:       personUUID,
			ChurchExternalID: churchExternalID,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating auth0-eligible users: %w", err)
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("no auth0-eligible users found (need numeric members_id, person_uuid and a church with external_id)")
	}

	tokens := make([]UserToken, 0, len(users))
	for _, u := range users {
		tokenString, err := mintAuth0Token(key, issuer, u, now, validity)
		if err != nil {
			return nil, fmt.Errorf("failed to sign auth0 token for %s: %w", u.UserID, err)
		}
		tokens = append(tokens, UserToken{UserID: u.UserID, Token: tokenString})
	}

	return tokens, nil
}
