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

-- name: RegenerateLeaderboards :exec
-- Rebuilds all leaderboard tables from score_journal (the full rebuild the
-- incremental apply must agree with)
SELECT count(*) FROM regenerate_leaderboards();

-- The group leaderboard queries as they read the stored group tables, before
-- the boards were computed from person totals: the reference the e2e tests
-- hold the new queries to. Removed together with the stored tables.

-- name: StoredProjectTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        lpt.score,
        lpt.last_score_at,
        DENSE_RANK() OVER (ORDER BY lpt.score DESC) AS rank
    FROM leaderboard_project_teams lpt
    INNER JOIN teams t ON lpt.team_id = t.id
    WHERE lpt.project_id = @projectid::char(28)
      AND t.leaderboard_excluded = false
      AND lpt.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpt.score <= @maxscore::int)
      -- Church filter: team has ANY member from specified church
      AND (@churchid::char(28) = '' OR EXISTS (
          SELECT 1 FROM team_members tm
          INNER JOIN users u ON tm.user_id = u.id
          WHERE tm.team_id = t.id AND u.church_id = @churchid::char(28)
      ))
)
SELECT entity_id, name, image, score, rank, last_score_at
FROM ranked_scores
ORDER BY rank ASC, last_score_at DESC NULLS LAST, name ASC;

-- name: StoredProjectSuperTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        lpst.score,
        lpst.last_score_at,
        DENSE_RANK() OVER (ORDER BY lpst.score DESC) AS rank
    FROM leaderboard_project_superteams lpst
    INNER JOIN super_teams st ON lpst.super_team_id = st.id
    WHERE lpst.project_id = @projectid::char(28)
      AND lpst.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpst.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank, last_score_at
FROM ranked_scores
ORDER BY rank ASC, last_score_at DESC NULLS LAST, name ASC;

-- name: StoredProjectChurchLeaderboard :many
WITH ranked_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        lpc.score,
        lpc.last_score_at,
        DENSE_RANK() OVER (ORDER BY lpc.score DESC) AS rank
    FROM leaderboard_project_churches lpc
    INNER JOIN churches c ON lpc.church_id = c.id
    WHERE lpc.project_id = @projectid::char(28)
      AND lpc.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpc.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank, last_score_at
FROM ranked_scores
ORDER BY rank ASC, last_score_at DESC NULLS LAST, name ASC;

-- name: StoredEventTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        let.score,
        let.last_score_at,
        DENSE_RANK() OVER (ORDER BY let.score DESC) AS rank
    FROM leaderboard_event_teams let
    INNER JOIN teams t ON let.team_id = t.id
    WHERE let.event_id = @eventid::char(28)
      AND t.leaderboard_excluded = false
      AND let.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR let.score <= @maxscore::int)
      -- Church filter: team has ANY member from specified church
      AND (@churchid::char(28) = '' OR EXISTS (
          SELECT 1 FROM team_members tm
          INNER JOIN users u ON tm.user_id = u.id
          WHERE tm.team_id = t.id AND u.church_id = @churchid::char(28)
      ))
)
SELECT entity_id, name, image, score, rank, last_score_at
FROM ranked_scores
ORDER BY rank ASC, last_score_at DESC NULLS LAST, name ASC;

-- name: StoredEventSuperTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        lest.score,
        lest.last_score_at,
        DENSE_RANK() OVER (ORDER BY lest.score DESC) AS rank
    FROM leaderboard_event_superteams lest
    INNER JOIN super_teams st ON lest.super_team_id = st.id
    WHERE lest.event_id = @eventid::char(28)
      AND lest.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lest.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank, last_score_at
FROM ranked_scores
ORDER BY rank ASC, last_score_at DESC NULLS LAST, name ASC;

-- name: StoredEventChurchLeaderboard :many
WITH ranked_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        lec.score,
        lec.last_score_at,
        DENSE_RANK() OVER (ORDER BY lec.score DESC) AS rank
    FROM leaderboard_event_churches lec
    INNER JOIN churches c ON lec.church_id = c.id
    WHERE lec.event_id = @eventid::char(28)
      AND lec.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lec.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank, last_score_at
FROM ranked_scores
ORDER BY rank ASC, last_score_at DESC NULLS LAST, name ASC;
