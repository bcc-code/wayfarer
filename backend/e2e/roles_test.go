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
