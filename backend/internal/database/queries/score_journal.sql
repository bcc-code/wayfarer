-- name: CreateScoreJournalEntry :one
INSERT INTO score_journal (
    id,
    project_id,
    user_id,
    event_id,
    challenge_id,
    points,
    source_type,
    source_id,
    reason,
    awarded_by,
    created_at
) VALUES (
    @id::text,
    @project_id::text,
    @user_id::text,
    sqlc.narg('event_id')::text,
    sqlc.narg('challenge_id')::text,
    @points::int,
    @source_type::text,
    sqlc.narg('source_id')::text,
    sqlc.narg('reason')::text,
    sqlc.narg('awarded_by')::text,
    COALESCE(sqlc.narg('created_at')::timestamptz, now())
) RETURNING *;

-- name: GetScoreJournalFiltered :many
SELECT
    id,
    project_id,
    user_id,
    event_id,
    challenge_id,
    points,
    source_type,
    source_id,
    reason,
    awarded_by,
    created_at
FROM score_journal
WHERE
    (@project_id::char(28) = '' OR project_id = @project_id::char(28))
    AND (@user_id::char(28) = '' OR user_id = @user_id::char(28))
    AND (@event_id::char(28) = '' OR event_id = @event_id::char(28))
    AND (@challenge_id::char(28) = '' OR challenge_id = @challenge_id::char(28))
    AND (@source_type::text = '' OR source_type = @source_type::text)
    AND (@aftercursor::char(28) = '' OR id > @aftercursor::char(28))
    AND (@beforecursor::char(28) = '' OR id < @beforecursor::char(28))
ORDER BY
    CASE WHEN @isbackward::bool = true THEN created_at END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN created_at END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountScoreJournalFiltered :one
SELECT COUNT(*)
FROM score_journal
WHERE
    (@project_id::char(28) = '' OR project_id = @project_id::char(28))
    AND (@user_id::char(28) = '' OR user_id = @user_id::char(28))
    AND (@event_id::char(28) = '' OR event_id = @event_id::char(28))
    AND (@challenge_id::char(28) = '' OR challenge_id = @challenge_id::char(28))
    AND (@source_type::text = '' OR source_type = @source_type::text);

-- name: GetScoreJournalByIDs :many
SELECT
    id,
    project_id,
    user_id,
    event_id,
    challenge_id,
    points,
    source_type,
    source_id,
    reason,
    awarded_by,
    created_at
FROM score_journal
WHERE id = ANY(@ids::char(28)[])
ORDER BY created_at DESC;

-- name: GetUserProjectScore :one
SELECT COALESCE(SUM(points), 0)::bigint AS total_score
FROM score_journal
WHERE user_id = @user_id::char(28)
    AND project_id = @project_id::char(28);

-- name: GetBulkUserProjectScores :many
-- Batch variant of GetUserProjectScore for the dataloader; pairs are matched by index
SELECT user_id, project_id, COALESCE(SUM(points), 0)::bigint AS total_score
FROM score_journal
WHERE (user_id, project_id) IN (
    SELECT unnest(@user_ids::char(28)[]), unnest(@project_ids::char(28)[])
)
GROUP BY user_id, project_id;

-- name: GetUserEventScore :one
SELECT COALESCE(SUM(points), 0)::bigint AS total_score
FROM score_journal
WHERE user_id = @user_id::char(28)
    AND project_id = @project_id::char(28)
    AND event_id = @event_id::char(28);

-- name: DeleteScoreJournalEntry :exec
DELETE FROM score_journal
WHERE id = @id::char(28);

-- name: DeleteScoreJournalByAchievement :exec
DELETE FROM score_journal
WHERE user_id = @user_id::char(28)
    AND source_type = 'ACHIEVEMENT'
    AND source_id = @achievement_id::char(28);

-- name: CheckScoreJournalEntryExists :one
-- Check if a score journal entry already exists for a specific source (e.g., achievement)
SELECT EXISTS(
    SELECT 1 FROM score_journal
    WHERE user_id = @user_id::char(28)
      AND source_type = @source_type::text
      AND source_id = @source_id::char(28)
) AS exists;

-- name: CheckScoreJournalEntryExistsBySource :one
-- Check if any score journal entry exists for a specific source (without user constraint)
SELECT EXISTS(
    SELECT 1 FROM score_journal
    WHERE source_type = @source_type::text
      AND source_id = @source_id::char(28)
) AS exists;

-- name: CreateTeamScoreAdjustmentBatch :many
-- Creates score journal entries for multiple team members at once
-- Points array must have the same length as user_ids array
INSERT INTO score_journal (
    id,
    project_id,
    user_id,
    event_id,
    points,
    source_type,
    reason,
    awarded_by,
    created_at
)
SELECT
    unnest(@ids::text[]),
    @project_id::text,
    unnest(@user_ids::text[]),
    sqlc.narg('event_id')::text,
    unnest(@points::int[]),
    'MANUAL',
    sqlc.narg('reason')::text,
    sqlc.narg('awarded_by')::text,
    now()
RETURNING *;

-- name: CreateBulkScoreAdjustmentBatch :many
-- Creates score journal entries for multiple users in one batch
-- Each adjustment can have different points and reason
-- All arrays must have the same length
INSERT INTO score_journal (
    id,
    project_id,
    user_id,
    event_id,
    points,
    source_type,
    reason,
    awarded_by,
    created_at
)
SELECT
    unnest(@ids::text[]),
    @project_id::text,
    unnest(@user_ids::text[]),
    sqlc.narg('event_id')::text,
    unnest(@points::int[]),
    'MANUAL',
    unnest(@reasons::text[]),
    sqlc.narg('awarded_by')::text,
    now()
RETURNING *;

