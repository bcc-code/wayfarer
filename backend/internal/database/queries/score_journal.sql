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
    (@project_id::text = '' OR project_id = @project_id::text)
    AND (@user_id::text = '' OR user_id = @user_id::text)
    AND (@event_id::text = '' OR event_id = @event_id::text)
    AND (@challenge_id::text = '' OR challenge_id = @challenge_id::text)
    AND (@source_type::text = '' OR source_type = @source_type::text)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN created_at END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN created_at END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountScoreJournalFiltered :one
SELECT COUNT(*)
FROM score_journal
WHERE
    (@project_id::text = '' OR project_id = @project_id::text)
    AND (@user_id::text = '' OR user_id = @user_id::text)
    AND (@event_id::text = '' OR event_id = @event_id::text)
    AND (@challenge_id::text = '' OR challenge_id = @challenge_id::text)
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
WHERE id = ANY(@ids::text[])
ORDER BY created_at DESC;

-- name: GetUserScore :one
SELECT COALESCE(SUM(points), 0)::bigint AS total_score
FROM score_journal
WHERE user_id = @user_id::text
    AND project_id = @project_id::text
    AND (@event_id::text = '' OR event_id = @event_id::text);

-- name: DeleteScoreJournalEntry :exec
DELETE FROM score_journal
WHERE id = @id::text;

-- name: DeleteScoreJournalByAchievement :exec
DELETE FROM score_journal
WHERE user_id = @user_id::text
    AND source_type = 'ACHIEVEMENT'
    AND source_id = @achievement_id::text;

-- name: CheckScoreJournalEntryExists :one
-- Check if a score journal entry already exists for a specific source (e.g., achievement)
SELECT EXISTS(
    SELECT 1 FROM score_journal
    WHERE user_id = @user_id::text
      AND source_type = @source_type::text
      AND source_id = @source_id::text
) AS exists;
