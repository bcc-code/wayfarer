package main

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/authtoken"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMintWayfarerToken_AcceptedByServer(t *testing.T) {
	cfg := config.JWTConfig{Secret: "loadtest-secret", Issuer: "wayfarer"}

	tokenString, err := mintWayfarerToken(cfg.Secret, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", []string{"user", "admin"}, time.Now(), time.Hour)
	require.NoError(t, err)

	// Parsed the way the server's JWT middleware parses access tokens
	claims, err := authtoken.Parse(tokenString, cfg)
	require.NoError(t, err)
	assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", claims.UserID)
	assert.Equal(t, []string{"user", "admin"}, claims.UserRoles)
}

func TestMintWayfarerToken_Expired(t *testing.T) {
	cfg := config.JWTConfig{Secret: "loadtest-secret", Issuer: "wayfarer"}

	tokenString, err := mintWayfarerToken(cfg.Secret, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", []string{"user"}, time.Now().Add(-2*time.Hour), time.Hour)
	require.NoError(t, err)

	_, err = authtoken.Parse(tokenString, cfg)
	assert.ErrorIs(t, err, authtoken.ErrInvalidToken)
}
