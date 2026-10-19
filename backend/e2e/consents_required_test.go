package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequiredConsentField covers consents.is_required through the GraphQL API.
func TestRequiredConsentField(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	adminUserID := data.UserIDs[0]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	const createMutation = `
		mutation CreateConsent($key: String!, $title: String!, $body: String!, $isRequired: Boolean) {
			createConsent(key: $key, title: $title, body: $body, isRequired: $isRequired) {
				id
				isRequired
			}
		}
	`

	create := func(t *testing.T, key string, isRequired any) (string, bool) {
		t.Helper()
		vars := map[string]any{"key": key, "title": "T", "body": "B"}
		if isRequired != nil {
			vars["isRequired"] = isRequired
		}

		resp := client.WithAuth(adminToken).MustExecute(t, createMutation, vars)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateConsent struct {
				ID         string `json:"id"`
				IsRequired bool   `json:"isRequired"`
			} `json:"createConsent"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		return result.CreateConsent.ID, result.CreateConsent.IsRequired
	}

	readIsRequired := func(t *testing.T, id string) bool {
		t.Helper()
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query Consent($id: ID!) {
				consent(id: $id) { isRequired }
			}
		`, map[string]any{"id": id})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Consent struct {
				IsRequired bool `json:"isRequired"`
			} `json:"consent"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		return result.Consent.IsRequired
	}

	t.Run("defaults to false when not given", func(t *testing.T) {
		id, isRequired := create(t, "not_required_by_default", nil)
		assert.False(t, isRequired)
		assert.False(t, readIsRequired(t, id))
	})

	t.Run("round-trips when set", func(t *testing.T) {
		id, isRequired := create(t, "required_consent", true)
		assert.True(t, isRequired)
		assert.True(t, readIsRequired(t, id))
	})

	t.Run("an update that omits required keeps the stored value", func(t *testing.T) {
		id, _ := create(t, "required_then_edited", true)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateConsent($id: ID!, $title: String) {
				updateConsent(id: $id, title: $title) { isRequired }
			}
		`, map[string]any{"id": id, "title": "Edited title"})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		assert.True(t, readIsRequired(t, id), "editing the title must not clear required")
	})

	t.Run("an update can change required", func(t *testing.T) {
		id, _ := create(t, "required_then_relaxed", true)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateConsent($id: ID!, $isRequired: Boolean) {
				updateConsent(id: $id, isRequired: $isRequired) { isRequired }
			}
		`, map[string]any{"id": id, "isRequired": false})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		assert.False(t, readIsRequired(t, id))
	})
}
