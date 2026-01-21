package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminSetUserConsent(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs
	targetUserID := data.UserIDs[0]
	adminUserID := data.UserIDs[1]
	superadminUserID := data.UserIDs[2]
	regularUserID := data.UserIDs[3]

	// Assign admin role and superadmin role
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))
	require.NoError(t, dbMgr.AssignRole(ctx, superadminUserID, testutil.RoleSuperAdmin))

	// Create test consents
	localConsentID, err := dbMgr.CreateTestConsent(ctx, "test_local_consent", false)
	require.NoError(t, err)

	remoteConsentID, err := dbMgr.CreateTestConsent(ctx, "test_remote_consent", true)
	require.NoError(t, err)

	// Setup test server
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

	t.Run("admin can accept consent for user", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
					id
					action
					source
					consent {
						id
						key
					}
				}
			}
		`, map[string]any{
			"userId":    targetUserID,
			"consentId": localConsentID,
			"action":    "ACCEPTED",
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AdminSetUserConsent struct {
				ID      string  `json:"id"`
				Action  string  `json:"action"`
				Source  *string `json:"source"`
				Consent struct {
					ID  string `json:"id"`
					Key string `json:"key"`
				} `json:"consent"`
			} `json:"adminSetUserConsent"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.AdminSetUserConsent.ID)
		assert.Equal(t, "ACCEPTED", result.AdminSetUserConsent.Action)
		assert.NotNil(t, result.AdminSetUserConsent.Source)
		assert.Equal(t, adminUserID, *result.AdminSetUserConsent.Source) // Admin ID recorded in source
		assert.Equal(t, localConsentID, result.AdminSetUserConsent.Consent.ID)
		assert.Equal(t, "test_local_consent", result.AdminSetUserConsent.Consent.Key)

		// Verify in database
		action, source, err := dbMgr.GetLatestUserConsentAction(ctx, targetUserID, "test_local_consent")
		require.NoError(t, err)
		assert.Equal(t, "ACCEPTED", action)
		assert.NotNil(t, source)
		assert.Equal(t, adminUserID, *source)
	})

	t.Run("admin can reject consent for user", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
					id
					action
					source
				}
			}
		`, map[string]any{
			"userId":    targetUserID,
			"consentId": localConsentID,
			"action":    "REJECTED",
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AdminSetUserConsent struct {
				ID     string  `json:"id"`
				Action string  `json:"action"`
				Source *string `json:"source"`
			} `json:"adminSetUserConsent"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, "REJECTED", result.AdminSetUserConsent.Action)
		assert.NotNil(t, result.AdminSetUserConsent.Source)
		assert.Equal(t, adminUserID, *result.AdminSetUserConsent.Source)

		// Verify in database - should have 2 entries now (ACCEPTED + REJECTED)
		count, err := dbMgr.GetUserConsentHistoryCount(ctx, targetUserID, "test_local_consent")
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// Latest should be REJECTED
		action, _, err := dbMgr.GetLatestUserConsentAction(ctx, targetUserID, "test_local_consent")
		require.NoError(t, err)
		assert.Equal(t, "REJECTED", action)
	})

	t.Run("superadmin can set consent for user", func(t *testing.T) {
		resp := client.WithAuth(superadminToken).MustExecute(t, `
			mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
					id
					action
					source
				}
			}
		`, map[string]any{
			"userId":    regularUserID,
			"consentId": localConsentID,
			"action":    "ACCEPTED",
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AdminSetUserConsent struct {
				ID     string  `json:"id"`
				Action string  `json:"action"`
				Source *string `json:"source"`
			} `json:"adminSetUserConsent"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, "ACCEPTED", result.AdminSetUserConsent.Action)
		assert.NotNil(t, result.AdminSetUserConsent.Source)
		assert.Equal(t, superadminUserID, *result.AdminSetUserConsent.Source)
	})

	t.Run("regular user cannot set consent for others", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
					id
				}
			}
		`, map[string]any{
			"userId":    targetUserID,
			"consentId": localConsentID,
			"action":    "ACCEPTED",
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("admin cannot set consent for remote-managed consent", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
					id
				}
			}
		`, map[string]any{
			"userId":    targetUserID,
			"consentId": remoteConsentID,
			"action":    "ACCEPTED",
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "remote-managed")
	})

	t.Run("admin cannot set consent for non-existent user", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
					id
				}
			}
		`, map[string]any{
			"userId":    "US01NONEXISTENTUSER12345678",
			"consentId": localConsentID,
			"action":    "ACCEPTED",
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "user not found")
	})

	t.Run("admin cannot set consent for non-existent consent", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
					id
				}
			}
		`, map[string]any{
			"userId":    targetUserID,
			"consentId": "CN01NONEXISTENTCONS12345678",
			"action":    "ACCEPTED",
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "consent not found")
	})
}
