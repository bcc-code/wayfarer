package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postJSON(t *testing.T, router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodePair(t *testing.T, rec *httptest.ResponseRecorder) services.TokenPair {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var pair services.TokenPair
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pair))
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
	return pair
}

func TestAuthSessions(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	userID := data.UserIDs[0]

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	testCache, err := testutil.NewTestCache()
	require.NoError(t, err)
	defer testCache.Close()
	sessions := services.NewAuthSessionService(dbMgr.DB.Queries, services.NewRoleService(dbMgr.DB.Queries, testCache), testutil.TestJWTConfig())

	start, err := sessions.StartSession(ctx, userID, "e2e")
	require.NoError(t, err)

	t.Run("access token authenticates GraphQL", func(t *testing.T) {
		resp := client.WithAuth(start.AccessToken).MustExecute(t, `query { me { id } }`, nil)
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
	})

	var rotated services.TokenPair
	t.Run("refresh rotates the token", func(t *testing.T) {
		rotated = decodePair(t, postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": start.RefreshToken}))
		assert.NotEqual(t, start.RefreshToken, rotated.RefreshToken)

		resp := client.WithAuth(rotated.AccessToken).MustExecute(t, `query { me { id } }`, nil)
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
	})

	t.Run("old token within grace period reports rotation", func(t *testing.T) {
		rec := postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": start.RefreshToken})
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "token_rotated")

		// The session survives a benign race.
		rotated = decodePair(t, postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": rotated.RefreshToken}))
	})

	t.Run("unknown token is rejected", func(t *testing.T) {
		rec := postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": "wfr_doesnotexist"})
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("missing body is a bad request", func(t *testing.T) {
		rec := postJSON(t, router, "/auth/refresh", gin.H{})
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("logout revokes the session", func(t *testing.T) {
		rec := postJSON(t, router, "/auth/logout", gin.H{"refresh_token": rotated.RefreshToken})
		assert.Equal(t, http.StatusNoContent, rec.Code)

		rec = postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": rotated.RefreshToken})
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		// Logout is idempotent.
		rec = postJSON(t, router, "/auth/logout", gin.H{"refresh_token": rotated.RefreshToken})
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("older token after two rotations revokes the session", func(t *testing.T) {
		a, err := sessions.StartSession(ctx, userID, "victim")
		require.NoError(t, err)
		// An attacker holding A rotates A→B→C.
		b := decodePair(t, postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": a.RefreshToken}))
		c := decodePair(t, postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": b.RefreshToken}))

		// The legitimate client's A is neither current nor previous.
		rec := postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": a.RefreshToken})
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		// The attacker's chain is dead too.
		rec = postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": c.RefreshToken})
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("logout with an already-rotated token ends the session", func(t *testing.T) {
		a, err := sessions.StartSession(ctx, userID, "racing-tabs")
		require.NoError(t, err)
		b := decodePair(t, postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": a.RefreshToken}))

		rec := postJSON(t, router, "/auth/logout", gin.H{"refresh_token": a.RefreshToken})
		assert.Equal(t, http.StatusNoContent, rec.Code)

		rec = postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": b.RefreshToken})
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("fabricated token with a known session ID cannot revoke it", func(t *testing.T) {
		a, err := sessions.StartSession(ctx, userID, "target")
		require.NoError(t, err)
		// Rotate once so the session has a retired token and a prev hash.
		live := decodePair(t, postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": a.RefreshToken}))

		// The session ID is public via the access token's sid claim.
		sessionID := live.RefreshToken[len("wfr_") : len("wfr_")+28]
		forged := "wfr_" + sessionID + "_" + strings.Repeat("A", 43)

		rec := postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": forged})
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		rec = postJSON(t, router, "/auth/logout", gin.H{"refresh_token": forged})
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// The real session is untouched.
		decodePair(t, postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": live.RefreshToken}))
	})

	t.Run("oversized body is rejected", func(t *testing.T) {
		huge := make([]byte, 64<<10)
		for i := range huge {
			huge[i] = 'A'
		}
		rec := postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": string(huge)})
		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})

	t.Run("revoking all user sessions", func(t *testing.T) {
		a, err := sessions.StartSession(ctx, userID, "device-a")
		require.NoError(t, err)
		b, err := sessions.StartSession(ctx, userID, "device-b")
		require.NoError(t, err)

		n, err := sessions.RevokeUserSessions(ctx, userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, n, int64(2))

		for _, tok := range []string{a.RefreshToken, b.RefreshToken} {
			rec := postJSON(t, router, "/auth/refresh", gin.H{"refresh_token": tok})
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	})
}
