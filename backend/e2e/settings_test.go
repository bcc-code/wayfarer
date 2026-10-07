package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const currentProjectQuery = `
	query CurrentProject {
		currentProject { id }
	}
`

const setCurrentProjectMutation = `
	mutation SetCurrentProject($projectId: ID!) {
		setCurrentProject(projectId: $projectId) { id name }
	}
`

func currentProjectID(t *testing.T, client *testutil.GraphQLClient) string {
	t.Helper()

	resp := client.MustExecute(t, currentProjectQuery, nil)
	require.Empty(t, resp.Errors)

	var result struct {
		CurrentProject struct{ ID string } `json:"currentProject"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &result))

	return result.CurrentProject.ID
}

func TestSetCurrentProject(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(data.ProjectIDs), 2, "the switch needs two projects to be meaningful")
	startingProjectID := data.ProjectIDs[0]
	targetProjectID := data.ProjectIDs[1]

	require.NoError(t, dbMgr.SetCurrentProject(ctx, startingProjectID))

	adminUserID := data.UserIDs[1]
	superadminUserID := data.UserIDs[2]
	regularUserID := data.UserIDs[3]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))
	require.NoError(t, dbMgr.AssignRole(ctx, superadminUserID, testutil.RoleSuperAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	userToken, err := testutil.GenerateUserToken(regularUserID)
	require.NoError(t, err)
	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)
	superadminToken, err := testutil.GenerateSuperAdminToken(superadminUserID)
	require.NoError(t, err)

	t.Run("a regular user cannot change it", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, setCurrentProjectMutation, map[string]any{
			"projectId": targetProjectID,
		})
		require.NotEmpty(t, resp.Errors)
		assert.Contains(t, resp.Errors[0].Message, "unauthorized")
	})

	t.Run("an admin cannot change it either", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, setCurrentProjectMutation, map[string]any{
			"projectId": targetProjectID,
		})
		require.NotEmpty(t, resp.Errors)
		assert.Contains(t, resp.Errors[0].Message, "unauthorized")
	})

	t.Run("an unknown project is rejected and the server stays up", func(t *testing.T) {
		resp := client.WithAuth(superadminToken).MustExecute(t, setCurrentProjectMutation, map[string]any{
			"projectId": "PR0000000000000000000000000",
		})
		require.NotEmpty(t, resp.Errors)

		// The old value must still be served: a rejected write may not leave
		// the in-memory map pointing at a project that does not exist.
		assert.Equal(t, startingProjectID, currentProjectID(t, client.WithAuth(superadminToken)))
	})

	t.Run("a superadmin switches the project and it takes effect immediately", func(t *testing.T) {
		resp := client.WithAuth(superadminToken).MustExecute(t, setCurrentProjectMutation, map[string]any{
			"projectId": targetProjectID,
		})
		require.Empty(t, resp.Errors)

		var result struct {
			SetCurrentProject struct {
				ID   string
				Name string
			} `json:"setCurrentProject"`
		}
		require.NoError(t, json.Unmarshal(resp.Data, &result))
		assert.Equal(t, targetProjectID, result.SetCurrentProject.ID)

		// Not after the response cache's TTL — now. This is what the project
		// invalidation in the resolver buys.
		assert.Equal(t, targetProjectID, currentProjectID(t, client.WithAuth(superadminToken)))
		assert.Equal(t, targetProjectID, currentProjectID(t, client.WithAuth(userToken)))
	})
}

func TestSetSetting(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.NoError(t, dbMgr.SetCurrentProject(ctx, data.ProjectIDs[0]))

	adminUserID := data.UserIDs[1]
	superadminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))
	require.NoError(t, dbMgr.AssignRole(ctx, superadminUserID, testutil.RoleSuperAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)
	superadminToken, err := testutil.GenerateSuperAdminToken(superadminUserID)
	require.NoError(t, err)

	const mutation = `
		mutation SetSetting($key: String!, $value: String!) {
			setSetting(key: $key, value: $value) { key value valueType }
		}
	`

	setSetting := func(t *testing.T, token, key, value string) *testutil.GraphQLResponse {
		t.Helper()
		return client.WithAuth(token).MustExecute(t, mutation, map[string]any{
			"key": key, "value": value,
		})
	}

	t.Run("an admin is refused", func(t *testing.T) {
		resp := setSetting(t, adminToken, "log_level", "debug")
		require.NotEmpty(t, resp.Errors)
		assert.Contains(t, resp.Errors[0].Message, "unauthorized")
	})

	t.Run("a superadmin writes each value type", func(t *testing.T) {
		for _, tc := range []struct{ key, value, valueType string }{
			{"log_level", "debug", "TEXT"},
			{"otel_enabled", "false", "BOOL"},
			{"otel_sampling_ratio", "0.25", "FLOAT"},
		} {
			resp := setSetting(t, superadminToken, tc.key, tc.value)
			require.Empty(t, resp.Errors, tc.key)

			var result struct {
				SetSetting struct {
					Key       string
					Value     string
					ValueType string
				} `json:"setSetting"`
			}
			require.NoError(t, json.Unmarshal(resp.Data, &result))
			assert.Equal(t, tc.value, result.SetSetting.Value, tc.key)
			assert.Equal(t, tc.valueType, result.SetSetting.ValueType, tc.key)
		}
	})

	t.Run("a value that does not parse is refused", func(t *testing.T) {
		resp := setSetting(t, superadminToken, "otel_sampling_ratio", "quite a lot")
		require.NotEmpty(t, resp.Errors)
	})

	t.Run("an unknown key cannot be invented", func(t *testing.T) {
		resp := setSetting(t, superadminToken, "brand_new_key", "1")
		require.NotEmpty(t, resp.Errors)

		// And it was not inserted: the API has no path that creates a row.
		listed := client.WithAuth(superadminToken).MustExecute(t, `
			query Settings { settings { key } }
		`, nil)
		require.Empty(t, listed.Errors)
		assert.NotContains(t, string(listed.Data), "brand_new_key")
	})
}

func TestSettingsQuery(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.NoError(t, dbMgr.SetCurrentProject(ctx, data.ProjectIDs[0]))

	adminUserID := data.UserIDs[1]
	superadminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))
	require.NoError(t, dbMgr.AssignRole(ctx, superadminUserID, testutil.RoleSuperAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)
	superadminToken, err := testutil.GenerateSuperAdminToken(superadminUserID)
	require.NoError(t, err)

	const settingsQuery = `
		query Settings {
			settings { key value valueType description requiresRestart envVar editable }
		}
	`

	t.Run("an admin is refused", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, settingsQuery, nil)
		require.NotEmpty(t, resp.Errors)
		assert.Contains(t, resp.Errors[0].Message, "superadmin")
	})

	t.Run("a superadmin sees the settings and what each one overrides", func(t *testing.T) {
		resp := client.WithAuth(superadminToken).MustExecute(t, settingsQuery, nil)
		require.Empty(t, resp.Errors)

		var result struct {
			Settings []struct {
				Key             string
				Value           string
				ValueType       string
				RequiresRestart bool
				EnvVar          *string
			} `json:"settings"`
		}
		require.NoError(t, json.Unmarshal(resp.Data, &result))
		require.NotEmpty(t, result.Settings)

		byKey := map[string]struct {
			restart bool
			envVar  *string
		}{}
		for _, setting := range result.Settings {
			byKey[setting.Key] = struct {
				restart bool
				envVar  *string
			}{setting.RequiresRestart, setting.EnvVar}
			if setting.Key == "current_project_id" {
				assert.Equal(t, data.ProjectIDs[0], setting.Value)
				assert.Equal(t, "TEXT", setting.ValueType)
			}
		}

		assert.False(t, byKey["log_level"].restart)
		require.NotNil(t, byKey["log_level"].envVar)
		assert.Equal(t, "LOG_LEVEL", *byKey["log_level"].envVar)

		assert.True(t, byKey["otel_enabled"].restart)

		// Application data, not configuration — it overrides no variable.
		assert.False(t, byKey["current_project_id"].restart)
		assert.Nil(t, byKey["current_project_id"].envVar)
	})
}
