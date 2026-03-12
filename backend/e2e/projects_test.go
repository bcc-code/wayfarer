package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjects(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs - use different users for different roles
	userID := data.UserIDs[0]
	adminUserID := data.UserIDs[1]
	superadminUserID := data.UserIDs[2]

	// Assign database roles
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

	t.Run("query myProjects returns user projects", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query {
				myProjects {
					id
					name
					description
					startDate
					endDate
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			MyProjects []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				StartDate   string `json:"startDate"`
				EndDate     string `json:"endDate"`
			} `json:"myProjects"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// User should be participating in at least one project due to seeding
		assert.GreaterOrEqual(t, len(result.MyProjects), 1)
	})

	t.Run("query project by id returns single project", func(t *testing.T) {
		projectID := data.ProjectIDs[0]

		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetProject($id: ID!) {
				project(id: $id) {
					id
					name
					description
				}
			}
		`, map[string]any{"id": projectID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Project struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"project"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, projectID, result.Project.ID)
		assert.NotEmpty(t, result.Project.Name)
	})

	t.Run("admin can list all projects via filter", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query {
				projects(first: 100) {
					edges {
						node {
							id
							name
							startDate
							endDate
						}
					}
					totalCount
					pageInfo {
						hasNextPage
						hasPreviousPage
					}
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Projects struct {
				Edges []struct {
					Node struct {
						ID        string `json:"id"`
						Name      string `json:"name"`
						StartDate string `json:"startDate"`
						EndDate   string `json:"endDate"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
				PageInfo   struct {
					HasNextPage     bool `json:"hasNextPage"`
					HasPreviousPage bool `json:"hasPreviousPage"`
				} `json:"pageInfo"`
			} `json:"projects"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Should have seeded projects
		assert.Equal(t, 2, result.Projects.TotalCount)
		assert.Len(t, result.Projects.Edges, 2)
	})

	t.Run("user can query projects list", func(t *testing.T) {
		// The projects query is publicly accessible (no @requireRole directive)
		resp := client.WithAuth(userToken).MustExecute(t, `
			query {
				projects(first: 10) {
					totalCount
					edges {
						node { id name }
					}
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Projects struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"projects"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Greater(t, result.Projects.TotalCount, 0)
	})

	// Helper to create valid branding input
	validBranding := map[string]any{
		"logo":     "",
		"rounding": 8,
		"colors": map[string]any{
			"light": map[string]any{
				"accent":            "#FF0000",
				"accentContrast":    "#938636",
				"onAccent":          "#01121a",
				"backgroundDefault": "#f3ede5",
				"backgroundRaised":  "#ffffff",
				"backgroundIndent":  "rgb(99 56 1 / 0.05)",
				"textDefault":       "#282521",
				"textMuted":         "rgb(40 37 33 / 0.65)",
				"textHint":          "rgb(40 37 33 / 0.4)",
				"shadowDefault":     "rgb(40 37 33 / 0.1)",
				"shadowBlank":       "rgb(40 37 33 / 0)",
				"borderDefault":     "rgb(40 37 33 / 0.15)",
			},
			"dark": map[string]any{
				"accent":            "#FF0000",
				"accentContrast":    "#e8dfa7",
				"onAccent":          "#1a1401",
				"backgroundDefault": "#122026",
				"backgroundRaised":  "#0a3644",
				"backgroundIndent":  "rgb(0 9 13 / 0.25)",
				"textDefault":       "#f3ede5",
				"textMuted":         "rgb(243 237 229 / 0.7)",
				"textHint":          "rgb(243 237 229 / 0.4)",
				"shadowDefault":     "rgb(18 32 38 / 0.3)",
				"shadowBlank":       "rgb(18 32 38 / 0)",
				"borderDefault":     "rgb(156 214 243 / 0.09)",
			},
		},
	}

	t.Run("user cannot create project", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation CreateProject($input: CreateProjectInput!) {
				createProject(input: $input) {
					id
					name
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":        "Test Project",
				"description": "Test description",
				"startDate":   "2025-01-01T00:00:00Z",
				"endDate":     "2025-12-31T23:59:59Z",
				"branding":    validBranding,
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("admin can create project", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateProject($input: CreateProjectInput!) {
				createProject(input: $input) {
					id
					name
					description
					startDate
					endDate
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":        "E2E Test Project",
				"description": "Created by E2E test",
				"startDate":   "2025-01-01T00:00:00Z",
				"endDate":     "2025-12-31T23:59:59Z",
				"branding":    validBranding,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateProject struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				StartDate   string `json:"startDate"`
				EndDate     string `json:"endDate"`
			} `json:"createProject"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateProject.ID)
		assert.Equal(t, "E2E Test Project", result.CreateProject.Name)
		assert.Equal(t, "Created by E2E test", result.CreateProject.Description)
	})

	t.Run("superadmin can create project", func(t *testing.T) {
		superadminToken, err := testutil.GenerateSuperAdminToken(superadminUserID)
		require.NoError(t, err)

		resp := client.WithAuth(superadminToken).MustExecute(t, `
			mutation CreateProject($input: CreateProjectInput!) {
				createProject(input: $input) {
					id
					name
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":        "Superadmin Project",
				"description": "Created by superadmin",
				"startDate":   "2025-02-01T00:00:00Z",
				"endDate":     "2025-11-30T23:59:59Z",
				"branding":    validBranding,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateProject struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"createProject"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateProject.ID)
		assert.Equal(t, "Superadmin Project", result.CreateProject.Name)
	})

	t.Run("user can query myPoints on currentProject", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query {
				currentProject {
					id
					name
					myPoints
				}
			}
		`, nil)

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CurrentProject struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				MyPoints int    `json:"myPoints"`
			} `json:"currentProject"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Points should be a non-negative integer
		assert.GreaterOrEqual(t, result.CurrentProject.MyPoints, 0)
	})

	t.Run("user can query myPoints on project by ID", func(t *testing.T) {
		projectID := data.ProjectIDs[0]

		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetProject($id: ID!) {
				project(id: $id) {
					id
					myPoints
				}
			}
		`, map[string]any{"id": projectID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Project struct {
				ID       string `json:"id"`
				MyPoints int    `json:"myPoints"`
			} `json:"project"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, projectID, result.Project.ID)
		// Points should be a non-negative integer
		assert.GreaterOrEqual(t, result.Project.MyPoints, 0)
	})

	t.Run("m2m user gets 0 for myPoints", func(t *testing.T) {
		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)

		projectID := data.ProjectIDs[0]

		resp := client.WithAuth(m2mToken).MustExecute(t, `
			query GetProject($id: ID!) {
				project(id: $id) {
					id
					myPoints
				}
			}
		`, map[string]any{"id": projectID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Project struct {
				ID       string `json:"id"`
				MyPoints int    `json:"myPoints"`
			} `json:"project"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// M2M users should always get 0 points
		assert.Equal(t, 0, result.Project.MyPoints)
	})
}
