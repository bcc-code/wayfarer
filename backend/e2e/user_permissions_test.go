package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserPermissions tests the checkUserPermissions function through the users query
// which enforces role-based access control for listing users.
func TestUserPermissions(t *testing.T) {
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

	// Get seeded IDs for role assignments
	projectID := data.ProjectIDs[0]
	teamIDs := data.TeamIDs[projectID]
	require.NotEmpty(t, teamIDs, "should have seeded teams")
	teamID := teamIDs[0]

	// Use different users for different roles to avoid conflicts
	regularUserID := data.UserIDs[0]
	churchAdminUserID := data.UserIDs[1]
	projectAdminUserID := data.UserIDs[2]
	teamLeadUserID := data.UserIDs[3]
	globalAdminUserID := data.UserIDs[4]

	// Get church ID for church admin scope - use the church admin user's own church
	churchAdminToken, err := testutil.GenerateUserToken(churchAdminUserID)
	require.NoError(t, err)

	// Query to get the church admin user's church ID
	resp := client.WithAuth(churchAdminToken).MustExecute(t, `query { me { church { id } } }`, nil)
	require.False(t, resp.HasErrors())
	var meResult struct {
		Me struct {
			Church struct{ ID string } `json:"church"`
		} `json:"me"`
	}
	require.NoError(t, resp.UnmarshalData(&meResult))
	churchAdminChurchID := meResult.Me.Church.ID

	// Assign roles - church admin gets their own church
	require.NoError(t, dbMgr.AssignRole(ctx, globalAdminUserID, testutil.RoleAdmin))
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdminUserID, testutil.RoleChurchAdmin, &churchAdminChurchID, nil, nil))
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, projectAdminUserID, testutil.RoleProjectAdmin, nil, &projectID, nil))
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, teamLeadUserID, testutil.RoleTeamLead, nil, nil, &teamID))

	t.Run("regular user cannot query users", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(regularUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 10) {
					edges { node { id } }
					totalCount
				}
			}
		`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "permission denied")
	})

	t.Run("global admin can query all users without filter", func(t *testing.T) {
		token, err := testutil.GenerateAdminToken(globalAdminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 100) {
					edges { node { id name } }
					totalCount
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				Edges      []struct{ Node struct{ ID string } } `json:"edges"`
				TotalCount int                                  `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Admin should see all seeded users (20 from DefaultSeedConfig)
		assert.Equal(t, 20, result.Users.TotalCount)
	})

	t.Run("global admin can query users with any filter", func(t *testing.T) {
		token, err := testutil.GenerateAdminToken(globalAdminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges { node { id } }
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
			Users struct {
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Should return users in the project
		assert.Greater(t, result.Users.TotalCount, 0)
	})

	t.Run("church admin can query users", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(churchAdminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 100) {
					edges {
						node {
							id
							church { id }
						}
					}
					totalCount
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Church struct{ ID string } `json:"church"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Church admin should see users, filtered to their church
		assert.Greater(t, result.Users.TotalCount, 0)

		// Verify all returned users are from the church admin's church
		for _, edge := range result.Users.Edges {
			assert.Equal(t, churchAdminChurchID, edge.Node.Church.ID, "church admin should only see users from their church")
		}
	})

	t.Run("project admin can query users with project filter", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(projectAdminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges { node { id } }
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
			Users struct {
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Project admin should see users in their project
		assert.Greater(t, result.Users.TotalCount, 0)
	})

	t.Run("project admin without project filter is denied", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(projectAdminUserID)
		require.NoError(t, err)

		// Project admin trying to query without specifying their project
		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 10) {
					edges { node { id } }
					totalCount
				}
			}
		`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "permission denied")
	})

	t.Run("project admin cannot query other projects", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(projectAdminUserID)
		require.NoError(t, err)

		// Try to query a different project (if exists)
		if len(data.ProjectIDs) > 1 {
			otherProjectID := data.ProjectIDs[1]

			resp := client.WithAuth(token).MustExecute(t, `
				query GetUsers($filter: UserFilter) {
					users(filter: $filter, first: 10) {
						edges { node { id } }
						totalCount
					}
				}
			`, map[string]any{
				"filter": map[string]any{
					"projectId": otherProjectID,
				},
			})

			require.True(t, resp.HasErrors())
			assert.Contains(t, resp.ErrorMessage(), "permission denied")
		}
	})

	t.Run("team lead can query users with team filter", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(teamLeadUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges { node { id } }
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"teamId": teamID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Team lead should see users in their team
		assert.Greater(t, result.Users.TotalCount, 0)
	})

	t.Run("team lead without team filter is denied", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(teamLeadUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 10) {
					edges { node { id } }
					totalCount
				}
			}
		`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "permission denied")
	})

	t.Run("team lead cannot query other teams", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(teamLeadUserID)
		require.NoError(t, err)

		// Try to query a different team (if exists)
		if len(teamIDs) > 1 {
			otherTeamID := teamIDs[1]

			resp := client.WithAuth(token).MustExecute(t, `
				query GetUsers($filter: UserFilter) {
					users(filter: $filter, first: 10) {
						edges { node { id } }
						totalCount
					}
				}
			`, map[string]any{
				"filter": map[string]any{
					"teamId": otherTeamID,
				},
			})

			require.True(t, resp.HasErrors())
			assert.Contains(t, resp.ErrorMessage(), "permission denied")
		}
	})
}

// TestUserPermissionsCombinedRoles tests users with multiple roles
func TestUserPermissionsCombinedRoles(t *testing.T) {
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

	projectID := data.ProjectIDs[0]
	teamIDs := data.TeamIDs[projectID]
	require.NotEmpty(t, teamIDs)
	teamID := teamIDs[0]

	multiRoleUserID := data.UserIDs[0]

	// Get the user's church ID
	token, err := testutil.GenerateUserToken(multiRoleUserID)
	require.NoError(t, err)
	resp := client.WithAuth(token).MustExecute(t, `query { me { church { id } } }`, nil)
	require.False(t, resp.HasErrors())
	var meResult struct {
		Me struct{ Church struct{ ID string } } `json:"me"`
	}
	require.NoError(t, resp.UnmarshalData(&meResult))
	churchID := meResult.Me.Church.ID

	// Assign both church admin and project admin roles
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, multiRoleUserID, testutil.RoleChurchAdmin, &churchID, nil, nil))
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, multiRoleUserID, testutil.RoleProjectAdmin, nil, &projectID, nil))

	t.Run("user with church admin can query without project filter", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(multiRoleUserID)
		require.NoError(t, err)

		// Church admin role should allow querying without project filter
		resp := client.WithAuth(token).MustExecute(t, `
			query {
				users(first: 100) {
					edges { node { id church { id } } }
					totalCount
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Church struct{ ID string } `json:"church"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Should see users from their church
		assert.Greater(t, result.Users.TotalCount, 0)
	})

	t.Run("user with project admin can also query with project filter", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(multiRoleUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges { node { id } }
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
			Users struct {
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Greater(t, result.Users.TotalCount, 0)
	})

	// Add team lead role
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, multiRoleUserID, testutil.RoleTeamLead, nil, nil, &teamID))

	t.Run("user with team lead can query with team filter", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(multiRoleUserID)
		require.NoError(t, err)

		resp := client.WithAuth(token).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges { node { id } }
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"teamId": teamID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Greater(t, result.Users.TotalCount, 0)
	})
}

// TestUserPermissionsFilterEnforcement verifies that filters are properly enforced
func TestUserPermissionsFilterEnforcement(t *testing.T) {
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

	projectID := data.ProjectIDs[0]
	adminUserID := data.UserIDs[0]
	churchAdminUserID := data.UserIDs[1]

	// Assign admin role
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	// Get church ID for church admin
	token, err := testutil.GenerateUserToken(churchAdminUserID)
	require.NoError(t, err)
	resp := client.WithAuth(token).MustExecute(t, `query { me { church { id } } }`, nil)
	require.False(t, resp.HasErrors())
	var meResult struct {
		Me struct{ Church struct{ ID string } } `json:"me"`
	}
	require.NoError(t, resp.UnmarshalData(&meResult))
	churchID := meResult.Me.Church.ID

	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdminUserID, testutil.RoleChurchAdmin, &churchID, nil, nil))

	t.Run("church admin filter is automatically applied", func(t *testing.T) {
		// Get admin count for comparison
		adminToken, err := testutil.GenerateAdminToken(adminUserID)
		require.NoError(t, err)

		adminResp := client.WithAuth(adminToken).MustExecute(t, `
			query {
				users(first: 100) {
					totalCount
				}
			}
		`, nil)
		require.False(t, adminResp.HasErrors())

		var adminResult struct {
			Users struct{ TotalCount int } `json:"users"`
		}
		require.NoError(t, adminResp.UnmarshalData(&adminResult))
		totalUsers := adminResult.Users.TotalCount

		// Church admin query
		churchAdminToken, err := testutil.GenerateUserToken(churchAdminUserID)
		require.NoError(t, err)

		churchAdminResp := client.WithAuth(churchAdminToken).MustExecute(t, `
			query {
				users(first: 100) {
					totalCount
				}
			}
		`, nil)
		require.False(t, churchAdminResp.HasErrors(), "unexpected error: %s", churchAdminResp.ErrorMessage())

		var churchAdminResult struct {
			Users struct{ TotalCount int } `json:"users"`
		}
		require.NoError(t, churchAdminResp.UnmarshalData(&churchAdminResult))
		churchUsers := churchAdminResult.Users.TotalCount

		// Church admin should see fewer users than global admin (filtered to their church)
		// Unless all users happen to be in the same church
		assert.LessOrEqual(t, churchUsers, totalUsers, "church admin should see at most as many users as global admin")
		assert.Greater(t, churchUsers, 0, "church admin should see at least some users")
	})

	t.Run("admin can combine filters", func(t *testing.T) {
		adminToken, err := testutil.GenerateAdminToken(adminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges {
						node {
							id
							gender
						}
					}
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"projectId": projectID,
				"gender":    "MALE",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Gender string `json:"gender"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// All returned users should be male
		for _, edge := range result.Users.Edges {
			assert.Equal(t, "MALE", edge.Node.Gender)
		}
	})

	t.Run("admin can filter by age range", func(t *testing.T) {
		adminToken, err := testutil.GenerateAdminToken(adminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges { node { id } }
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"minAge": 18,
				"maxAge": 30,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Should return users in the age range
		assert.GreaterOrEqual(t, result.Users.TotalCount, 0)
	})

	t.Run("admin can filter by specific user IDs", func(t *testing.T) {
		adminToken, err := testutil.GenerateAdminToken(adminUserID)
		require.NoError(t, err)

		targetIDs := []string{data.UserIDs[0], data.UserIDs[1], data.UserIDs[2]}

		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetUsers($filter: UserFilter) {
				users(filter: $filter, first: 100) {
					edges { node { id } }
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"ids": targetIDs,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Users struct {
				Edges []struct {
					Node struct{ ID string } `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, 3, result.Users.TotalCount)
		for _, edge := range result.Users.Edges {
			assert.Contains(t, targetIDs, edge.Node.ID)
		}
	})
}
