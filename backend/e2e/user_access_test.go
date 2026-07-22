package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserAccessScoping exercises the Query.user access-control rules through the
// GraphQL API. It builds a controlled two-church / two-project fixture so the
// cross-church and cross-project cases are deterministic (the default seed spreads
// users randomly across churches).
//
// Test kinds:
//   - LOCK:   legitimate behavior that must NOT change across the refactor.
//   - TARGET: encodes the DESIRED post-fix behavior. Tagged with the fix that turns
//     it green; expected to FAIL against the current (vulnerable) code.
func TestUserAccessScoping(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	// Seed to populate settings (current_project_id) and baseline infra.
	_, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// --- Controlled fixture ---------------------------------------------------
	churchA := ulid.NewChurchID()
	churchB := ulid.NewChurchID()
	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchA, "Church A", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchB, "Church B", "NO", "S"))

	projectA := ulid.NewProjectID()
	projectB := ulid.NewProjectID()
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectA, "Project A"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectB, "Project B"))

	birth := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	// alice: church A, member of project A
	alice := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, alice, "Alice", "FEMALE", birth, churchA))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, alice, projectA))

	// bob: church B, member of project B only
	bob := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, bob, "Bob", "MALE", birth, churchB))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, bob, projectB))

	// Actors
	projAdmin := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, projAdmin, "Proj Admin", "MALE", birth, churchA))
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, projAdmin, testutil.RoleProjectAdmin, nil, &projectA, nil))

	churchAdmin := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, churchAdmin, "Church Admin", "MALE", birth, churchA))
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdmin, testutil.RoleChurchAdmin, &churchA, nil, nil))

	adminUser := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, adminUser, "Admin", "MALE", birth, churchA))
	require.NoError(t, dbMgr.AssignRole(ctx, adminUser, testutil.RoleAdmin))

	regularUser := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, regularUser, "Regular", "MALE", birth, churchA))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// userQuery selects a sensitive field (email) to prove full-User access.
	const userQuery = `query GetUser($id: ID!) { user(id: $id) { id email } }`

	getUser := func(t *testing.T, token, targetID string) *testutil.GraphQLResponse {
		t.Helper()
		return client.WithAuth(token).MustExecute(t, userQuery, map[string]any{"id": targetID})
	}

	t.Run("LOCK self can read own user", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(alice)
		require.NoError(t, err)
		resp := getUser(t, token, alice)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})

	t.Run("LOCK admin can read any user", func(t *testing.T) {
		token, err := testutil.GenerateAdminToken(adminUser)
		require.NoError(t, err)
		resp := getUser(t, token, bob)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})

	t.Run("LOCK m2m can read any user", func(t *testing.T) {
		token, err := testutil.GenerateM2MToken()
		require.NoError(t, err)
		resp := getUser(t, token, bob)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})

	t.Run("LOCK church admin can read same-church user", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(churchAdmin)
		require.NoError(t, err)
		resp := getUser(t, token, alice) // alice is in church A
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})

	t.Run("LOCK church admin cannot read cross-church user", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(churchAdmin)
		require.NoError(t, err)
		resp := getUser(t, token, bob) // bob is in church B
		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "permission denied")
	})

	t.Run("LOCK regular user cannot read another user", func(t *testing.T) {
		token, err := testutil.GenerateUserToken(regularUser)
		require.NoError(t, err)
		resp := getUser(t, token, alice)
		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "permission denied")
	})

	t.Run("LOCK project admin can read user in their project", func(t *testing.T) {
		// alice is a member of project A, which projAdmin administers.
		token, err := testutil.GenerateUserToken(projAdmin)
		require.NoError(t, err)
		resp := getUser(t, token, alice)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})

	t.Run("TARGET(Fix1) project admin cannot read user outside their project", func(t *testing.T) {
		// bob is only in project B; projAdmin administers project A. Currently
		// allowed via the unscoped IsProjectAdmin branch — must become denied.
		token, err := testutil.GenerateUserToken(projAdmin)
		require.NoError(t, err)
		resp := getUser(t, token, bob)
		require.True(t, resp.HasErrors(), "project admin must NOT access users outside their project")
		assert.Contains(t, resp.ErrorMessage(), "permission denied")
	})
}
