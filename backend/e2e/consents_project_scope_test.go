package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectScopedConsents covers consents.project_id: a consent applies to
// one project, or to every project when project_id is NULL.
func TestProjectScopedConsents(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data.ProjectIDs), 2, "scoping needs two projects")

	projectA, projectB := data.ProjectIDs[0], data.ProjectIDs[1]
	userID := data.UserIDs[0]
	adminUserID := data.UserIDs[1]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	_, err = dbMgr.CreateTestProjectConsent(ctx, "scoped_to_project_a", &projectA)
	require.NoError(t, err)
	_, err = dbMgr.CreateTestProjectConsent(ctx, "applies_everywhere", nil)
	require.NoError(t, err)

	userToken, err := testutil.GenerateUserToken(userID)
	require.NoError(t, err)

	// A fresh server per project: SettingsService reads current_project_id into
	// memory at startup, so switching it needs a new server to take effect.
	pendingKeys := func(t *testing.T, currentProject string) []string {
		require.NoError(t, dbMgr.SetCurrentProject(ctx, currentProject))

		router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
		require.NoError(t, err)
		defer cleanup()

		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		resp := client.WithAuth(userToken).MustExecute(t, `
			query PendingConsents {
				pendingConsents {
					key
					project { id }
				}
			}
		`, map[string]any{})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			PendingConsents []struct {
				Key     string `json:"key"`
				Project *struct {
					ID string `json:"id"`
				} `json:"project"`
			} `json:"pendingConsents"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		keys := make([]string, 0, len(result.PendingConsents))
		for _, consent := range result.PendingConsents {
			keys = append(keys, consent.Key)

			switch consent.Key {
			case "scoped_to_project_a":
				require.NotNil(t, consent.Project, "a scoped consent resolves its project")
				assert.Equal(t, projectA, consent.Project.ID)
			case "applies_everywhere":
				assert.Nil(t, consent.Project, "a global consent has no project")
			}
		}

		return keys
	}

	t.Run("scoped consent is pending for its own project", func(t *testing.T) {
		keys := pendingKeys(t, projectA)

		assert.Contains(t, keys, "scoped_to_project_a")
		assert.Contains(t, keys, "applies_everywhere")
	})

	t.Run("scoped consent is not pending for another project", func(t *testing.T) {
		keys := pendingKeys(t, projectB)

		assert.NotContains(t, keys, "scoped_to_project_a")
		assert.Contains(t, keys, "applies_everywhere", "global consents apply to every project")
	})

	t.Run("a new version cannot move a consent to another project", func(t *testing.T) {
		require.NoError(t, dbMgr.SetCurrentProject(ctx, projectA))

		router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
		require.NoError(t, err)
		defer cleanup()

		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		adminToken, err := testutil.GenerateAdminToken(adminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateConsent($key: String!, $title: String!, $body: String!, $projectId: ID) {
				createConsent(key: $key, title: $title, body: $body, projectId: $projectId) {
					id
				}
			}
		`, map[string]any{
			"key":       "scoped_to_project_a",
			"title":     "Version 2",
			"body":      "Updated body",
			"projectId": projectB,
		})

		require.True(t, resp.HasErrors(), "re-scoping should be rejected")
		assert.Contains(t, resp.ErrorMessage(), "cannot change")
	})
}
