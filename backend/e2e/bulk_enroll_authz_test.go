package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
)

// TestBulkEnrollAuthorization covers who a church admin may bulk-enroll in a
// challenge: members of the challenge's project who are in their church, or
// on a team in that project with a member from their church. Admins may
// enroll anyone.
func TestBulkEnrollAuthorization(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data.ProjectIDs), 2)
	require.GreaterOrEqual(t, len(data.ChurchIDs), 2)

	projectA, projectB := data.ProjectIDs[0], data.ProjectIDs[1]
	churchA, churchB := data.ChurchIDs[0], data.ChurchIDs[1]
	require.NotEmpty(t, data.ChallengeIDs[projectA])
	challengeID := data.ChallengeIDs[projectA][0]

	var (
		churchAdminID = data.UserIDs[0] // church admin of church A
		systemAdminID = data.UserIDs[1] // global admin
		ownMember     = data.UserIDs[2] // church A, project A
		teamMate      = data.UserIDs[3] // church A, project A, on the mixed team
		mixedLead     = data.UserIDs[4] // church B, project A, on the mixed team
		otherMember   = data.UserIDs[5] // church B, project A, no shared team
		otherProject  = data.UserIDs[6] // church A, project B only
		fixtureUsers  = []string{ownMember, teamMate, mixedLead, otherMember, otherProject}
		inProjectA    = []string{ownMember, teamMate, mixedLead, otherMember}
		exec          = func(query string, args ...any) {
			t.Helper()
			_, err := dbMgr.DB.Pool.Exec(ctx, query, args...)
			require.NoError(t, err)
		}
	)

	// Fixture: exact churches, project memberships and one mixed team, so the
	// random seed relationships don't matter
	exec(`UPDATE users SET church_id = $1 WHERE id = ANY($2)`, churchA, []string{ownMember, teamMate, otherProject})
	exec(`UPDATE users SET church_id = $1 WHERE id = ANY($2)`, churchB, []string{mixedLead, otherMember})
	exec(`DELETE FROM team_members WHERE user_id = ANY($1)`, fixtureUsers)
	exec(`DELETE FROM user_projects WHERE user_id = ANY($1)`, fixtureUsers)
	exec(`DELETE FROM user_challenge_enrollments WHERE user_id = ANY($1)`, fixtureUsers)
	for _, u := range inProjectA {
		exec(`INSERT INTO user_projects (user_id, project_id) VALUES ($1, $2)`, u, projectA)
	}
	exec(`INSERT INTO user_projects (user_id, project_id) VALUES ($1, $2)`, otherProject, projectB)
	teamID := ulid.NewTeamID()
	exec(`INSERT INTO teams (id, project_id, name, join_code, created_by_user_id) VALUES ($1, $2, 'Mixed team', $3, $4)`,
		teamID, projectA, "MIX"+teamID[len(teamID)-5:], teamMate)
	exec(`INSERT INTO team_members (team_id, user_id) VALUES ($1, $2), ($1, $3)`, teamID, teamMate, mixedLead)

	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdminID, testutil.RoleChurchAdmin, &churchA, nil, nil))
	require.NoError(t, dbMgr.AssignRole(ctx, systemAdminID, testutil.RoleAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()
	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	churchAdminToken, err := testutil.GenerateChurchAdminToken(churchAdminID)
	require.NoError(t, err)
	systemAdminToken, err := testutil.GenerateAdminToken(systemAdminID)
	require.NoError(t, err)

	enroll := func(token string, target map[string]any) string {
		t.Helper()
		resp, err := client.WithAuth(token).Execute(ctx, `
			mutation($target: EnrollmentTargetInput!, $challengeId: ID!) {
				bulkEnrollUsersInChallenge(target: $target, challengeId: $challengeId) { id }
			}`, map[string]any{"target": target, "challengeId": challengeID})
		require.NoError(t, err)
		return resp.ErrorMessage()
	}
	enrolled := func(userID string) bool {
		t.Helper()
		var ok bool
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM user_challenge_enrollments WHERE user_id = $1 AND challenge_id = $2)`,
			userID, challengeID).Scan(&ok))
		return ok
	}

	t.Run("church admin may enroll own church members and leads of their church teams", func(t *testing.T) {
		assert.Empty(t, enroll(churchAdminToken, map[string]any{"userIds": []string{ownMember, mixedLead}}))
		assert.True(t, enrolled(ownMember))
		assert.True(t, enrolled(mixedLead), "a lead from another church on a team with a church member")
	})

	t.Run("church admin is rejected as a whole when one user is out of scope", func(t *testing.T) {
		msg := enroll(churchAdminToken, map[string]any{"userIds": []string{teamMate, otherMember}})
		assert.Contains(t, msg, "not allowed to enroll 1 user(s) outside your church")
		assert.False(t, enrolled(teamMate), "nobody is enrolled when the request is rejected")
		assert.False(t, enrolled(otherMember))
	})

	t.Run("church admin may not target another church", func(t *testing.T) {
		msg := enroll(churchAdminToken, map[string]any{
			"churchInProject": map[string]any{"churchId": churchB, "projectId": projectA},
		})
		assert.Contains(t, msg, "outside your church")
	})

	t.Run("church admin may not enroll members of another project", func(t *testing.T) {
		msg := enroll(churchAdminToken, map[string]any{"allProjectMembers": projectB})
		assert.Contains(t, msg, "outside your church")
		assert.False(t, enrolled(otherProject))
	})

	t.Run("church admin is rejected on the async mutation before a job is created", func(t *testing.T) {
		resp, err := client.WithAuth(churchAdminToken).Execute(ctx, `
			mutation($target: EnrollmentTargetInput!, $challengeId: ID!) {
				bulkEnrollUsersInChallengeAsync(target: $target, challengeId: $challengeId) { id }
			}`, map[string]any{"target": map[string]any{"userIds": []string{otherMember}}, "challengeId": challengeID})
		require.NoError(t, err)
		assert.Contains(t, resp.ErrorMessage(), "outside your church")

		var jobs int
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM bulk_jobs WHERE created_by = $1`, churchAdminID).Scan(&jobs))
		assert.Zero(t, jobs)
		assert.False(t, enrolled(otherMember))
	})

	t.Run("admin may enroll anyone", func(t *testing.T) {
		assert.Empty(t, enroll(systemAdminToken, map[string]any{"userIds": []string{otherMember, otherProject}}))
		assert.True(t, enrolled(otherMember))
		assert.True(t, enrolled(otherProject))
	})

	t.Run("m2m may enroll anyone", func(t *testing.T) {
		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)
		exec(`DELETE FROM user_challenge_enrollments WHERE user_id = $1 AND challenge_id = $2`, otherMember, challengeID)

		assert.Empty(t, enroll(m2mToken, map[string]any{"userIds": []string{otherMember}}))
		assert.True(t, enrolled(otherMember))
	})

	t.Run("bulk enroll returns only newly enrolled users", func(t *testing.T) {
		users := []string{teamMate, ownMember} // ownMember is already enrolled
		first, err := dbMgr.DB.Queries.BulkEnrollUsersInChallenge(ctx, sqlc.BulkEnrollUsersInChallengeParams{
			Userids: users, Challengeid: challengeID,
		})
		require.NoError(t, err)
		assert.Equal(t, []string{teamMate}, first)

		second, err := dbMgr.DB.Queries.BulkEnrollUsersInChallenge(ctx, sqlc.BulkEnrollUsersInChallengeParams{
			Userids: users, Challengeid: challengeID,
		})
		require.NoError(t, err)
		assert.Empty(t, second, "already enrolled users are not returned, so they are not notified again")
	})
}
