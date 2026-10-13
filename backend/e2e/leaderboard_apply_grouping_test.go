package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// leaderboardTables are all tables the apply worker maintains
var leaderboardTables = []string{
	"leaderboard_project_persons", "leaderboard_project_teams",
	"leaderboard_project_superteams", "leaderboard_project_churches",
	"leaderboard_event_persons", "leaderboard_event_teams",
	"leaderboard_event_superteams", "leaderboard_event_churches",
}

// snapshotLeaderboards returns every leaderboard row as JSON (without
// updated_at, which only says when a row was written), keyed by table
func snapshotLeaderboards(t *testing.T, ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}) map[string][]string {
	t.Helper()
	snap := map[string][]string{}
	for _, table := range leaderboardTables {
		rows, err := q.Query(ctx, `SELECT (to_jsonb(l) - 'updated_at')::text FROM `+table+` l ORDER BY 1`)
		require.NoError(t, err)
		for rows.Next() {
			var row string
			require.NoError(t, rows.Scan(&row))
			snap[table] = append(snap[table], row)
		}
		require.NoError(t, rows.Err())
	}
	return snap
}

// TestLeaderboardApplyGroupsDeltas checks that draining the queue, which sums
// a batch's deltas per user, project and event before applying them
// (migration 00110), ends in exactly the state of applying every queued
// delta on its own (the drain before 00110): repeated deltas per user, event
// and project-only deltas, negative deltas, and batches that split a user's
// deltas.
func TestLeaderboardApplyGroupsDeltas(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.NoError(t, dbMgr.DrainLeaderboardApplyQueue(ctx))

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]
	users := data.UserIDs[:20]

	// Like a bet settlement: per user a stake and a winnings entry per bet,
	// interleaved with event-scoped entries and other users
	type entry struct {
		user    string
		event   *string
		points  int
		created time.Time
	}
	var entries []entry
	base := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	for i, u := range users {
		for bet := 0; bet < 5; bet++ {
			at := base.Add(time.Duration(i*10+bet) * time.Second)
			entries = append(entries,
				entry{u, nil, -(10 + bet), at},
				entry{u, nil, (10 + bet) * (i % 3), at}, // 0x, 1x or 2x payout
			)
		}
		entries = append(entries, entry{u, &eventID, 7 - i, base.Add(time.Duration(i) * time.Minute)})
	}
	for _, e := range entries {
		_, err := dbMgr.DB.Pool.Exec(ctx,
			`INSERT INTO score_journal (id, project_id, user_id, event_id, points, source_type, reason, created_at)
			 VALUES ($1, $2, $3, $4, $5, 'MANUAL', 'grouping test', $6)`,
			ulid.NewScoreJournalID(), projectID, e.user, e.event, e.points, e.created)
		require.NoError(t, err)
	}

	var queued int
	require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx, `SELECT count(*) FROM leaderboard_apply_queue`).Scan(&queued))
	require.Equal(t, len(entries), queued)

	// Reference: every queued delta applied on its own, in queue order, in a
	// transaction that is rolled back afterwards
	tx, err := dbMgr.DB.Pool.Begin(ctx)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `SELECT count(*) FROM (
		SELECT apply_score_journal_delta(user_id, project_id, event_id, points_delta, score_at)
		FROM leaderboard_apply_queue ORDER BY id) applied`)
	require.NoError(t, err)
	perDelta := snapshotLeaderboards(t, ctx, tx)
	require.NoError(t, tx.Rollback(ctx))

	// Small batches split users' deltas across drains; the return value is
	// the number of queue rows consumed, which the worker's loop relies on
	consumed := 0
	for {
		n, err := dbMgr.DB.Queries.DrainLeaderboardApplyQueue(ctx, 7)
		require.NoError(t, err)
		consumed += int(n)
		if n < 7 {
			break
		}
	}
	assert.Equal(t, len(entries), consumed, "every queued row is reported as applied")
	require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx, `SELECT count(*) FROM leaderboard_apply_queue`).Scan(&queued))
	assert.Zero(t, queued)

	grouped := snapshotLeaderboards(t, ctx, dbMgr.DB.Pool)
	for _, table := range leaderboardTables {
		assert.NotEmpty(t, perDelta[table], "%s has rows", table)
		assert.Equal(t, perDelta[table], grouped[table], "%s: grouped drain equals applying each delta", table)
	}
}
