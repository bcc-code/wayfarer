package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/cmd/seed/seeders"
	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup admin and superadmin users with database roles
	adminUserID := data.UserIDs[1]
	superadminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))
	require.NoError(t, dbMgr.AssignRole(ctx, superadminUserID, testutil.RoleSuperAdmin))

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	t.Run("unauthorized request returns error", func(t *testing.T) {
		resp := client.MustExecute(t, `query { me { id } }`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "401")
	})

	t.Run("expired token returns error", func(t *testing.T) {
		expiredToken, err := testutil.GenerateExpiredToken(data.UserIDs[0])
		require.NoError(t, err)

		resp := client.WithAuth(expiredToken).MustExecute(t, `query { me { id } }`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "401")
	})

	t.Run("wrong secret token returns error", func(t *testing.T) {
		wrongToken, err := testutil.GenerateTokenWithWrongSecret(data.UserIDs[0])
		require.NoError(t, err)

		resp := client.WithAuth(wrongToken).MustExecute(t, `query { me { id } }`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "401")
	})

	t.Run("valid user token can query me", func(t *testing.T) {
		userID := data.UserIDs[0]
		token, err := testutil.GenerateUserToken(userID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				me {
					id
					name
					church { id name }
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Me struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Church struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"church"`
			} `json:"me"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, userID, result.Me.ID)
		assert.NotEmpty(t, result.Me.Name)
		assert.NotEmpty(t, result.Me.Church.ID)
		assert.NotEmpty(t, result.Me.Church.Name)
	})

	t.Run("user role cannot access admin query users", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(data.UserIDs[0])
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 10) {
					edges { node { id name } }
					totalCount
				}
			}
		`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "permission denied")
	})

	t.Run("admin role can access admin query users", func(t *testing.T) {
		// Use adminUserID which has ADMIN role assigned in database
		token, err := testutil.GenerateAdminToken(adminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 10) {
					edges { node { id name } }
					totalCount
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Greater(t, result.Users.TotalCount, 0)
		assert.NotEmpty(t, result.Users.Edges)
	})

	t.Run("superadmin role can access admin queries", func(t *testing.T) {
		// Use superadminUserID which has SUPERADMIN role assigned in database
		token, err := testutil.GenerateSuperAdminToken(superadminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				projects(first: 10) {
					edges { node { id name } }
					totalCount
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Projects struct {
				TotalCount int `json:"totalCount"`
			} `json:"projects"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, 2, result.Projects.TotalCount) // DefaultSeedConfig has 2 projects
	})
}

func TestLanguageSync(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	userID := data.UserIDs[0]
	token, err := testutil.GenerateUserToken(userID)
	require.NoError(t, err)

	t.Run("language is updated when Accept-Language header differs from stored", func(t *testing.T) {
		// Set initial language in database
		require.NoError(t, dbMgr.SetUserLanguage(ctx, userID, "no"))

		// Verify initial language
		initialLang, err := dbMgr.GetUserLanguage(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "no", initialLang)

		// Make request with different language
		resp := client.WithAuth(token).WithLanguage("en-US").MustExecute(t, `
			query {
				me { id language }
			}
		`, nil)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		// Verify language was updated in database
		updatedLang, err := dbMgr.GetUserLanguage(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "en", updatedLang, "language should be updated to 'en' from 'en-US' header")
	})

	t.Run("language is not updated when Accept-Language header matches stored", func(t *testing.T) {
		// Set language to match what we'll send
		require.NoError(t, dbMgr.SetUserLanguage(ctx, userID, "de"))

		// Make request with same language
		resp := client.WithAuth(token).WithLanguage("de-DE").MustExecute(t, `
			query {
				me { id }
			}
		`, nil)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		// Verify language is still the same
		lang, err := dbMgr.GetUserLanguage(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "de", lang, "language should remain 'de'")
	})

	t.Run("language defaults to Norwegian when no Accept-Language header", func(t *testing.T) {
		// Set initial language to something else
		require.NoError(t, dbMgr.SetUserLanguage(ctx, userID, "fr"))

		// Make request without language header
		resp := client.WithAuth(token).MustExecute(t, `
			query {
				me { id }
			}
		`, nil)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		// Verify language was updated to default Norwegian
		lang, err := dbMgr.GetUserLanguage(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "no", lang, "language should be updated to default 'no'")
	})

	t.Run("language returned in me query reflects stored value", func(t *testing.T) {
		// Set a specific language
		require.NoError(t, dbMgr.SetUserLanguage(ctx, userID, "sv"))

		// Make request with same language (so it doesn't get updated)
		resp := client.WithAuth(token).WithLanguage("sv").MustExecute(t, `
			query {
				me { id language }
			}
		`, nil)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Me struct {
				ID       string `json:"id"`
				Language string `json:"language"`
			} `json:"me"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, userID, result.Me.ID)
		assert.Equal(t, "sv", result.Me.Language)
	})
}

func TestAuthWithMinimalSeed(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with minimal data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 123, seeders.SeedConfig{
		NumUsers:                  5,
		NumProjects:               1,
		NumChurches:               1,
		NumSuperTeams:             1,
		NumAchievements:           3,
		TeamSize:                  3,
		ProjectParticipationRate:  1.0,
		AchievementCompletionRate: 0.5,
	})
	require.NoError(t, err)

	// Assign admin role to first user for admin query access
	adminUserID := data.UserIDs[0]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	t.Run("seeded data is deterministic", func(t *testing.T) {
		// Verify we have exactly what we seeded
		token, err := testutil.GenerateAdminToken(adminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 100) { totalCount }
				projects(first: 100) { totalCount }
				churches(first: 100) { totalCount }
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users     struct{ TotalCount int } `json:"users"`
			Projects  struct{ TotalCount int } `json:"projects"`
			Churches  struct{ TotalCount int } `json:"churches"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, 5, result.Users.TotalCount)
		assert.Equal(t, 1, result.Projects.TotalCount)
		assert.Equal(t, 1, result.Churches.TotalCount)
	})
}
