package e2e

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/plugins"
	"github.com/bcc-media/wayfarer/internal/plugins/teams"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTeamsPluginWebhook checks that the team-name-changed receiver only acts
// on signed requests and takes the project from the team, not the payload.
func TestTeamsPluginWebhook(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data.ProjectIDs), 2)

	projectA, projectB := data.ProjectIDs[0], data.ProjectIDs[1]
	require.NotEmpty(t, data.TeamIDs[projectA])
	require.NotEmpty(t, data.ChallengeIDs[projectA])
	teamID := data.TeamIDs[projectA][0]
	challengeID := data.ChallengeIDs[projectA][0]

	// Make sure the team has a member, so a processed request writes journal rows
	_, err = dbMgr.DB.Pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		teamID, data.UserIDs[0])
	require.NoError(t, err)

	testCache, err := testutil.NewTestCache()
	require.NoError(t, err)
	deps := plugins.Dependencies{DB: dbMgr.DB, Cache: testCache}

	const secret = "teams-test-secret"
	newRouter := func(secretKey string) *gin.Engine {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		plugins.RegisterPlugin(router, deps, nil, teams.NewPlugin(teams.Config{
			SecretKey:         secretKey,
			RenameChallengeID: challengeID,
		}))
		return router
	}
	router := newRouter(secret)

	payload := func(projectID string) []byte {
		body, err := json.Marshal(map[string]any{
			"event_type": "team_name_changed",
			"timestamp":  time.Now().Format(time.RFC3339),
			"project_id": projectID,
			"data":       map[string]any{"team_id": teamID, "old_name": "Old", "new_name": "New"},
		})
		require.NoError(t, err)
		return body
	}
	sign := func(body []byte) string {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		return "sha256=" + hex.EncodeToString(mac.Sum(nil))
	}
	post := func(r *gin.Engine, body []byte, signature string) int {
		req := httptest.NewRequest(http.MethodPost, "/plugins/teams/team-name-changed", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if signature != "" {
			req.Header.Set("X-Webhook-Signature", signature)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}
	journalRows := func(projectID string) int {
		var n int
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx, `
			SELECT count(*) FROM score_journal
			WHERE source_type = 'PLUGIN' AND source_id = $1 AND project_id = $2`,
			teamID, projectID).Scan(&n))
		return n
	}

	t.Run("rejects every request when no secret is configured", func(t *testing.T) {
		body := payload(projectA)
		assert.Equal(t, http.StatusServiceUnavailable, post(newRouter(""), body, ""))
		assert.Zero(t, journalRows(projectA))
	})

	t.Run("rejects an unsigned request", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, post(router, payload(projectA), ""))
		assert.Zero(t, journalRows(projectA))
	})

	t.Run("rejects a project_id that is not the team's project", func(t *testing.T) {
		body := payload(projectB)
		assert.Equal(t, http.StatusBadRequest, post(router, body, sign(body)))
		assert.Zero(t, journalRows(projectB))
		assert.Zero(t, journalRows(projectA))
	})

	t.Run("awards points in the team's project for a signed request", func(t *testing.T) {
		body := payload(projectA)
		assert.Equal(t, http.StatusCreated, post(router, body, sign(body)))
		assert.Positive(t, journalRows(projectA))
	})
}
