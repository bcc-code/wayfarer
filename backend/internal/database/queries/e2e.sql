-- Queries only the e2e tests use: state checks and race setups the API has no
-- reason to expose. Kept in sqlc so schema changes break them at compile time.

-- name: SnapshotLeaderboardRows :many
-- Every row of every leaderboard table the apply worker maintains, as JSON
-- without updated_at (which only says when a row was written)
SELECT t.table_name, t.row_json
FROM (
    SELECT 'leaderboard_project_persons'::text AS table_name, (to_jsonb(l) - 'updated_at')::text AS row_json FROM leaderboard_project_persons l
    UNION ALL
    SELECT 'leaderboard_project_teams', (to_jsonb(l) - 'updated_at')::text FROM leaderboard_project_teams l
    UNION ALL
    SELECT 'leaderboard_project_superteams', (to_jsonb(l) - 'updated_at')::text FROM leaderboard_project_superteams l
    UNION ALL
    SELECT 'leaderboard_project_churches', (to_jsonb(l) - 'updated_at')::text FROM leaderboard_project_churches l
    UNION ALL
    SELECT 'leaderboard_event_persons', (to_jsonb(l) - 'updated_at')::text FROM leaderboard_event_persons l
    UNION ALL
    SELECT 'leaderboard_event_teams', (to_jsonb(l) - 'updated_at')::text FROM leaderboard_event_teams l
    UNION ALL
    SELECT 'leaderboard_event_superteams', (to_jsonb(l) - 'updated_at')::text FROM leaderboard_event_superteams l
    UNION ALL
    SELECT 'leaderboard_event_churches', (to_jsonb(l) - 'updated_at')::text FROM leaderboard_event_churches l
) t
ORDER BY t.table_name, t.row_json;

-- name: ApplyLeaderboardApplyQueuePerDelta :one
-- Applies every queued delta on its own, in queue order, without dequeuing
-- them: the drain before migration 00110 grouped deltas, as a reference
SELECT count(*) AS applied
FROM (
    SELECT apply_score_journal_delta(q.user_id, q.project_id, q.event_id, q.points_delta, q.score_at)
    FROM leaderboard_apply_queue q
    ORDER BY q.id
) a;

-- name: LockLeaderboardApplyQueue :one
-- Row-locks every queued delta (for the caller's transaction), so a drain
-- skips them (SKIP LOCKED) like a read that beats the apply worker
SELECT count(*) AS locked
FROM (SELECT id FROM leaderboard_apply_queue FOR UPDATE) q;
