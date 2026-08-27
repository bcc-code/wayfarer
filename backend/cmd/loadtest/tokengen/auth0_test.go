package main

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/handlers"
	"github.com/bcc-media/wayfarer/internal/loadtestauth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMintAuth0Token_ClaimsMatchServerShape(t *testing.T) {
	key, _, err := resolveAuth0Key("")
	require.NoError(t, err)

	now := time.Now()
	user := auth0User{
		UserID:           "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
		PersonID:         12345,
		PersonUUID:       "a2f7c1de-3b4a-4f4f-9c1d-0e2b3c4d5e6f",
		ChurchExternalID: 42,
	}

	tokenString, err := mintAuth0Token(key, "https://login.bcc.no/", user, now, 24*time.Hour)
	require.NoError(t, err)

	// Parse with the same claims struct the server's callback handler uses
	var claims handlers.Auth0Claims
	parsed, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		assert.Equal(t, loadtestauth.KeyID, token.Header["kid"])
		return key.Public(), nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	assert.Equal(t, "https://login.bcc.no/", claims.Issuer)
	assert.Equal(t, 42, claims.ChurchID)
	assert.Equal(t, 12345, claims.PersonID)
	assert.Equal(t, user.PersonUUID, claims.PersonUUID)
	assert.True(t, claims.AppMetadata.HasMembership)
	assert.Equal(t, 12345, claims.AppMetadata.PersonID)
}

func TestResolveAuth0Key_RoundTrip(t *testing.T) {
	// Fresh key when no spec given
	key, encoded, err := resolveAuth0Key("")
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	// Feeding the printed env value back yields the same key — a token minted
	// with one verifies against the other (tokengen rerun with -auth0-key).
	key2, encoded2, err := resolveAuth0Key(encoded)
	require.NoError(t, err)
	assert.True(t, key.Equal(key2))
	assert.Equal(t, encoded, encoded2)
}
