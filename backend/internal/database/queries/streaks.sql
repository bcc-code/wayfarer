-- name: GetStreaksByIDs :many
SELECT id, project_id, name, description, created_at, updated_at
FROM streaks
WHERE id = ANY(@ids::text[]);

-- name: GetStreaksFilteredCursor :many
SELECT id, project_id, name, description, created_at, updated_at
FROM streaks
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountStreaksFiltered :one
SELECT COUNT(DISTINCT id)
FROM streaks
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text);
