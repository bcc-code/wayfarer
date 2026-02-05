package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoles(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs
	userID := data.UserIDs[0]
	adminUserID := data.UserIDs[1]
	superadminUserID := data.UserIDs[2]
	targetUserID := data.UserIDs[3] // User to assign roles to

	// Assign admin role and superadmin role
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))
	require.NoError(t, dbMgr.AssignRole(ctx, superadminUserID, testutil.RoleSuperAdmin))

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	userToken, err := testutil.GenerateUserToken(userID)
	require.NoError(t, err)

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	superadminToken, err := testutil.GenerateSuperAdminToken(superadminUserID)
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]
	churchID := data.ChurchIDs[0]
	teamIDs := data.TeamIDs[projectID]

	t.Run("superadmin can assign global admin role", func(t *testing.T) {
		// Only superadmin can assign global ADMIN role
		resp := client.WithAuth(superadminToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) {
					id
					role
					scope {
						type
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"userId": targetUserID,
				"role":   "ADMIN",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AssignRole struct {
				ID    string `json:"id"`
				Role  string `json:"role"`
				Scope *struct {
					Type string `json:"type"`
				} `json:"scope"`
			} `json:"assignRole"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.AssignRole.ID)
		assert.Equal(t, "ADMIN", result.AssignRole.Role)
		assert.Nil(t, result.AssignRole.Scope) // Global role has no scope
	})

	t.Run("admin can assign church admin role", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) {
					id
					role
					scope {
						type
						id
						church {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"userId":    targetUserID,
				"role":      "CHURCH_ADMIN",
				"scopeType": "CHURCH",
				"scopeId":   churchID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AssignRole struct {
				ID    string `json:"id"`
				Role  string `json:"role"`
				Scope struct {
					Type   string `json:"type"`
					ID     string `json:"id"`
					Church *struct {
						ID string `json:"id"`
					} `json:"church"`
				} `json:"scope"`
			} `json:"assignRole"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, "CHURCH_ADMIN", result.AssignRole.Role)
		assert.Equal(t, "CHURCH", result.AssignRole.Scope.Type)
		assert.NotNil(t, result.AssignRole.Scope.Church)
	})

	t.Run("church admin can assign church admin role to user in same church", func(t *testing.T) {
		// Setup: Get two users from the same church
		churchAdminUserID := data.UserIDs[6]
		targetUserInSameChurch := data.UserIDs[7]

		// Get the church admin user's church ID via GraphQL
		tempToken, err := testutil.GenerateUserToken(churchAdminUserID)
		require.NoError(t, err)

		meResp := client.WithAuth(tempToken).MustExecute(t, `query { me { church { id } } }`, nil)
		require.False(t, meResp.HasErrors())

		var meResult struct {
			Me struct {
				Church struct{ ID string } `json:"church"`
			} `json:"me"`
		}
		require.NoError(t, meResp.UnmarshalData(&meResult))
		churchAdminChurchID := meResult.Me.Church.ID

		// Assign church admin role to the first user
		require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdminUserID, testutil.RoleChurchAdmin, &churchAdminChurchID, nil, nil))

		// Generate token for church admin
		churchAdminToken, err := testutil.GenerateChurchAdminToken(churchAdminUserID)
		require.NoError(t, err)

		// Church admin assigns church admin role to another user in the same church
		resp := client.WithAuth(churchAdminToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) {
					id
					role
					scope {
						type
						id
						church {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"userId":    targetUserInSameChurch,
				"role":      "CHURCH_ADMIN",
				"scopeType": "CHURCH",
				"scopeId":   churchAdminChurchID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AssignRole struct {
				ID    string `json:"id"`
				Role  string `json:"role"`
				Scope struct {
					Type   string `json:"type"`
					ID     string `json:"id"`
					Church *struct {
						ID string `json:"id"`
					} `json:"church"`
				} `json:"scope"`
			} `json:"assignRole"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, "CHURCH_ADMIN", result.AssignRole.Role)
		assert.Equal(t, "CHURCH", result.AssignRole.Scope.Type)
		assert.NotNil(t, result.AssignRole.Scope.Church)
		assert.Equal(t, churchAdminChurchID, result.AssignRole.Scope.Church.ID)
	})

	t.Run("admin can assign project admin role", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) {
					id
					role
					scope {
						type
						id
						project {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"userId":    targetUserID,
				"role":      "PROJECT_ADMIN",
				"scopeType": "PROJECT",
				"scopeId":   projectID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AssignRole struct {
				ID    string `json:"id"`
				Role  string `json:"role"`
				Scope struct {
					Type    string `json:"type"`
					ID      string `json:"id"`
					Project *struct {
						ID string `json:"id"`
					} `json:"project"`
				} `json:"scope"`
			} `json:"assignRole"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, "PROJECT_ADMIN", result.AssignRole.Role)
		assert.Equal(t, "PROJECT", result.AssignRole.Scope.Type)
		assert.NotNil(t, result.AssignRole.Scope.Project)
	})

	t.Run("admin can assign team lead role", func(t *testing.T) {
		if len(teamIDs) == 0 {
			t.Skip("No teams available")
		}

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) {
					id
					role
					scope {
						type
						id
						team {
							id
						}
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"userId":    targetUserID,
				"role":      "TEAM_LEAD",
				"scopeType": "TEAM",
				"scopeId":   teamIDs[0],
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AssignRole struct {
				ID    string `json:"id"`
				Role  string `json:"role"`
				Scope struct {
					Type string `json:"type"`
					ID   string `json:"id"`
					Team *struct {
						ID string `json:"id"`
					} `json:"team"`
				} `json:"scope"`
			} `json:"assignRole"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, "TEAM_LEAD", result.AssignRole.Role)
		assert.Equal(t, "TEAM", result.AssignRole.Scope.Type)
		assert.NotNil(t, result.AssignRole.Scope.Team)
	})

	t.Run("admin can revoke role", func(t *testing.T) {
		// First assign a role to revoke
		assignResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"userId":    data.UserIDs[4],
				"role":      "PROJECT_ADMIN",
				"scopeType": "PROJECT",
				"scopeId":   projectID,
			},
		})
		require.False(t, assignResp.HasErrors())

		// Now revoke it
		revokeResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation RevokeRole($input: RevokeRoleInput!) {
				revokeRole(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"userId":    data.UserIDs[4],
				"role":      "PROJECT_ADMIN",
				"scopeType": "PROJECT",
				"scopeId":   projectID,
			},
		})

		require.False(t, revokeResp.HasErrors(), "unexpected error: %s", revokeResp.ErrorMessage())

		var result struct {
			RevokeRole bool `json:"revokeRole"`
		}
		require.NoError(t, revokeResp.UnmarshalData(&result))
		assert.True(t, result.RevokeRole)
	})

	t.Run("user cannot assign roles", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"userId": data.UserIDs[5],
				"role":   "ADMIN",
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("query userRoles", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetUserRoles($userId: ID!) {
				userRoles(userId: $userId) {
					id
					role
					scope {
						type
						id
					}
				}
			}
		`, map[string]any{
			"userId": targetUserID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			UserRoles []struct {
				ID    string `json:"id"`
				Role  string `json:"role"`
				Scope *struct {
					Type string `json:"type"`
					ID   string `json:"id"`
				} `json:"scope"`
			} `json:"userRoles"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		// Target user should have roles we assigned
		assert.Greater(t, len(result.UserRoles), 0)
	})

	t.Run("query usersWithRole", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetUsersWithRole($role: RoleType!) {
				usersWithRole(role: $role) {
					id
					name
				}
			}
		`, map[string]any{
			"role": "ADMIN",
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			UsersWithRole []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"usersWithRole"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		// Should have at least the admin user and target user
		assert.GreaterOrEqual(t, len(result.UsersWithRole), 1)
	})

	t.Run("query usersWithRole with scope filter", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetUsersWithRole($role: RoleType!, $scopeType: ScopeType, $scopeId: ID) {
				usersWithRole(role: $role, scopeType: $scopeType, scopeId: $scopeId) {
					id
					name
				}
			}
		`, map[string]any{
			"role":      "PROJECT_ADMIN",
			"scopeType": "PROJECT",
			"scopeId":   projectID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			UsersWithRole []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"usersWithRole"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		// Should have at least the target user we assigned project admin to
		assert.GreaterOrEqual(t, len(result.UsersWithRole), 1)
	})
}

// TestRoles_ImmediateRoleChange tests that role changes take effect immediately
// without requiring the user to re-authenticate (get a new token).
func TestRoles_ImmediateRoleChange(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	superadminUserID := data.UserIDs[0]
	testUserID := data.UserIDs[1]

	// Assign superadmin role to first user
	require.NoError(t, dbMgr.AssignRole(ctx, superadminUserID, testutil.RoleSuperAdmin))

	router, testCache, cleanup, err := testutil.SetupTestServerWithCache(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// Generate token for test user (has only "user" role in token, no admin)
	testUserToken, err := testutil.GenerateUserToken(testUserID)
	require.NoError(t, err)

	superadminToken, err := testutil.GenerateSuperAdminToken(superadminUserID)
	require.NoError(t, err)

	// Use clearAllCache mutation which requires admin role and has no complex inputs
	clearCacheMutation := `mutation { clearAllCache }`

	t.Run("user cannot perform admin mutation before role assignment", func(t *testing.T) {
		resp := client.WithAuth(testUserToken).MustExecute(t, clearCacheMutation, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("user can perform admin mutation immediately after role assignment", func(t *testing.T) {
		// Superadmin assigns admin role to test user
		assignResp := client.WithAuth(superadminToken).MustExecute(t, `
			mutation AssignRole($input: AssignRoleInput!) {
				assignRole(input: $input) { id role }
			}
		`, map[string]any{
			"input": map[string]any{
				"userId": testUserID,
				"role":   "ADMIN",
			},
		})
		require.False(t, assignResp.HasErrors(), "assign error: %s", assignResp.ErrorMessage())

		// Clear cache to ensure fresh lookup
		testCache.InvalidateUser(testUserID)

		// Now test user should be able to clear cache using the SAME token
		// (token still has "user" role, but DB has "admin")
		resp := client.WithAuth(testUserToken).MustExecute(t, clearCacheMutation, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})

	t.Run("user loses access immediately after role revocation", func(t *testing.T) {
		// Revoke the admin role
		revokeResp := client.WithAuth(superadminToken).MustExecute(t, `
			mutation RevokeRole($input: RevokeRoleInput!) {
				revokeRole(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"userId": testUserID,
				"role":   "ADMIN",
			},
		})
		require.False(t, revokeResp.HasErrors())

		// Clear cache
		testCache.InvalidateUser(testUserID)

		// Now test user should NOT be able to clear cache (role revoked)
		resp := client.WithAuth(testUserToken).MustExecute(t, clearCacheMutation, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})
}

// TestRoles_M2MBypassesDBLookup tests that M2M users use token-based roles
// and don't require a database lookup.
func TestRoles_M2MBypassesDBLookup(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// Generate M2M token (has m2m role in token, user ID is "M2M_SERVICE")
	m2mToken, err := testutil.GenerateM2MToken()
	require.NoError(t, err)

	// M2M should be able to perform mutations based on token roles
	// even if user doesn't exist in database
	t.Run("M2M can revoke achievement based on token roles", func(t *testing.T) {
		// Use revokeAchievement mutation which requires m2m role
		// This will succeed (return false) even if the achievement wasn't awarded
		projectID := data.ProjectIDs[0]
		achievementIDs := data.AchievementIDs[projectID]
		if len(achievementIDs) == 0 {
			t.Skip("No achievements available")
		}

		// Use a user that wasn't seeded with this achievement (last user)
		targetUserID := data.UserIDs[len(data.UserIDs)-1]

		resp := client.WithAuth(m2mToken).MustExecute(t, `
			mutation RevokeAchievement($userId: ID!, $achievementId: ID!) {
				revokeAchievement(userId: $userId, achievementId: $achievementId)
			}
		`, map[string]any{
			"userId":        targetUserID,
			"achievementId": achievementIDs[0],
		})

		// Should succeed based on m2m token role (returns false since not awarded)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})
}
