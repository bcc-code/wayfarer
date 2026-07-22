package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hasFieldValidationError reports whether the response contains a GraphQL
// field-validation error (raised before execution when a selected field does
// not exist on the resolved type).
func hasFieldValidationError(resp *testutil.GraphQLResponse, field string) bool {
	for _, e := range resp.Errors {
		if strings.Contains(e.Message, "Cannot query field") && strings.Contains(e.Message, field) {
			return true
		}
	}
	return false
}

// TestNestedUserExposure verifies that sensitive User PII cannot be selected
// through nested, non-privileged GraphQL paths. These paths return the full User
// object today (protected only by clients not selecting PII); after the PublicUser
// split they expose id/name/image only.
//
//   - LOCK:        id/name remain selectable through the nested path.
//   - TARGET(Item6): selecting email through the nested path becomes a field-validation error.
func TestNestedUserExposure(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	_, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	churchA := ulid.NewChurchID()
	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchA, "Church A", "NO", "S"))

	projectA := ulid.NewProjectID()
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectA, "Project A"))

	birth := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	teamA := ulid.NewTeamID()
	require.NoError(t, dbMgr.CreateTestTeam(ctx, teamA, "Team A", projectA))

	member := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, member, "Member", "MALE", birth, churchA))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, member, teamA))

	adminUser := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, adminUser, "Admin", "MALE", birth, churchA))
	require.NoError(t, dbMgr.AssignRole(ctx, adminUser, testutil.RoleAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUser)
	require.NoError(t, err)

	t.Run("LOCK TeamMember.user id/name selectable", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query TeamMembers($id: ID!) {
				team(id: $id) { members { user { id name } } }
			}
		`, map[string]any{"id": teamA})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Team struct {
				Members []struct {
					User struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"user"`
				} `json:"members"`
			} `json:"team"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.Team.Members, 1)
		assert.Equal(t, member, result.Team.Members[0].User.ID)
	})

	t.Run("TARGET(Item6) TeamMember.user email not selectable", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query TeamMembers($id: ID!) {
				team(id: $id) { members { user { email } } }
			}
		`, map[string]any{"id": teamA})
		assert.True(t, hasFieldValidationError(resp, "email"),
			"email must not be selectable through TeamMember.user; got: %s", resp.ErrorMessage())
	})

	t.Run("TARGET(Item6) QuizSession.createdBy email not selectable", func(t *testing.T) {
		// Uses a non-existent id; we assert only on the schema-level validation error,
		// which is raised before execution regardless of whether the session exists.
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query CreatedBy($id: ID!) {
				quizSession(id: $id) { createdBy { email } }
			}
		`, map[string]any{"id": "QS00000000000000000000000000"})
		assert.True(t, hasFieldValidationError(resp, "email"),
			"email must not be selectable through QuizSession.createdBy; got: %s", resp.ErrorMessage())
	})
}
