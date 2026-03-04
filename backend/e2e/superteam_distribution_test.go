package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DistributionResponse matches the handler response
type DistributionResponse struct {
	Superteams []SuperteamResult `json:"superteams"`
	Variance   float64           `json:"variance"`
}

type SuperteamResult struct {
	SuperTeamID string     `json:"super_team_id"`
	Name        string     `json:"name"`
	TotalScore  int64      `json:"total_score"`
	TeamCount   int        `json:"team_count"`
	MemberCount int32      `json:"member_count"`
	Teams       []TeamInfo `json:"teams"`
	Churches    []string   `json:"churches"`
}

type TeamInfo struct {
	TeamID      string `json:"team_id"`
	TeamName    string `json:"team_name"`
	ChurchID    string `json:"church_id"`
	TotalScore  int64  `json:"total_score"`
	MemberCount int32  `json:"member_count"`
}

func TestSuperTeamDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup admin user
	adminUserID := data.UserIDs[1]
	userID := data.UserIDs[0]

	// Assign admin role
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	// Create HTTP test server
	server := httptest.NewServer(router)
	defer server.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	userToken, err := testutil.GenerateUserToken(userID)
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]

	t.Run("preview requires authentication", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/plugins/ladder-to-heaven/preview-superteams?project_id="+projectID, nil)
		require.NoError(t, err)
		// No auth header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("preview requires admin role", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/plugins/ladder-to-heaven/preview-superteams?project_id="+projectID, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+userToken) // User, not admin

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("preview requires project_id", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/plugins/ladder-to-heaven/preview-superteams", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("admin can preview distribution", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/plugins/ladder-to-heaven/preview-superteams?project_id="+projectID, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result DistributionResponse
		require.NoError(t, json.Unmarshal(body, &result))

		// Should have 4 superteams in preview
		assert.Len(t, result.Superteams, 4)

		// Superteam names should be Purple, Green, Red, Yellow
		names := make([]string, 4)
		for i, st := range result.Superteams {
			names[i] = st.Name
		}
		assert.Contains(t, names, "Purple")
		assert.Contains(t, names, "Green")
		assert.Contains(t, names, "Red")
		assert.Contains(t, names, "Yellow")

		// Preview should not assign IDs
		for _, st := range result.Superteams {
			assert.Empty(t, st.SuperTeamID, "Preview should not assign IDs")
		}
	})

	t.Run("distribute requires authentication", func(t *testing.T) {
		reqBody := map[string]string{"project_id": projectID}
		bodyBytes, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", server.URL+"/plugins/ladder-to-heaven/distribute-superteams", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		// No auth header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("distribute requires admin role", func(t *testing.T) {
		reqBody := map[string]string{"project_id": projectID}
		bodyBytes, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", server.URL+"/plugins/ladder-to-heaven/distribute-superteams", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken) // User, not admin

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("admin can execute distribution", func(t *testing.T) {
		reqBody := map[string]string{"project_id": projectID}
		bodyBytes, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", server.URL+"/plugins/ladder-to-heaven/distribute-superteams", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result DistributionResponse
		require.NoError(t, json.Unmarshal(body, &result))

		// Should have 4 superteams
		assert.Len(t, result.Superteams, 4)

		// Execute should assign IDs
		for _, st := range result.Superteams {
			assert.NotEmpty(t, st.SuperTeamID, "Execute should assign IDs")
			assert.Len(t, st.SuperTeamID, 28, "SuperTeamID should be 28 characters")
			assert.Equal(t, "ST", st.SuperTeamID[:2], "SuperTeamID should have ST prefix")
		}

		// Calculate total teams distributed
		totalTeams := 0
		for _, st := range result.Superteams {
			totalTeams += st.TeamCount
		}

		// Variance should be non-negative
		assert.GreaterOrEqual(t, result.Variance, 0.0)
	})

	t.Run("distribution can be verified via graphql", func(t *testing.T) {
		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		// Query superteams for the project
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSuperTeams($filter: SuperTeamFilter) {
				superteams(filter: $filter, first: 100) {
					edges {
						node {
							id
							name
							teams {
								id
								name
							}
						}
					}
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"projectId": projectID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Superteams struct {
				Edges []struct {
					Node struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						Teams []struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"teams"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"superteams"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// After distribution, should have exactly 4 superteams
		assert.Equal(t, 4, result.Superteams.TotalCount)

		// Verify superteam names
		names := make([]string, len(result.Superteams.Edges))
		for i, edge := range result.Superteams.Edges {
			names[i] = edge.Node.Name
		}
		assert.Contains(t, names, "Purple")
		assert.Contains(t, names, "Green")
		assert.Contains(t, names, "Red")
		assert.Contains(t, names, "Yellow")
	})

	t.Run("running distribution again replaces existing superteams", func(t *testing.T) {
		// First run
		reqBody := map[string]string{"project_id": projectID}
		bodyBytes, _ := json.Marshal(reqBody)

		req1, err := http.NewRequest("POST", server.URL+"/plugins/ladder-to-heaven/distribute-superteams", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+adminToken)

		resp1, err := http.DefaultClient.Do(req1)
		require.NoError(t, err)
		resp1.Body.Close()
		require.Equal(t, http.StatusOK, resp1.StatusCode)

		// Second run should still succeed (replaces old superteams)
		bodyBytes, _ = json.Marshal(reqBody)
		req2, err := http.NewRequest("POST", server.URL+"/plugins/ladder-to-heaven/distribute-superteams", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+adminToken)

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		// Verify only 4 superteams exist (not 8)
		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		gqlResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSuperTeams($filter: SuperTeamFilter) {
				superteams(filter: $filter, first: 100) {
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"projectId": projectID,
			},
		})

		require.False(t, gqlResp.HasErrors())

		var result struct {
			Superteams struct {
				TotalCount int `json:"totalCount"`
			} `json:"superteams"`
		}
		require.NoError(t, gqlResp.UnmarshalData(&result))
		assert.Equal(t, 4, result.Superteams.TotalCount)
	})
}
