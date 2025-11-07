-- name: GetChallengesByIDs :many
SELECT id, project_id, event_id, name, description, image_url, url, button_text, published_at, end_time, created_at, updated_at
FROM challenges
WHERE id = ANY(@ids::text[]);

-- name: GetChallengesByProjectIDs :many
SELECT id, project_id, event_id, name, description, image_url, url, button_text, published_at, end_time, created_at, updated_at
FROM challenges
WHERE project_id = ANY(@project_ids::text[])
    AND published_at IS NOT NULL
    AND published_at <= NOW()
ORDER BY project_id, published_at DESC;

-- name: GetChallengesFilteredCursor :many
SELECT id, project_id, event_id, name, description, image_url, url, button_text, published_at, end_time, created_at, updated_at
FROM challenges
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@eventid::text = '' OR event_id = @eventid::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountChallengesFiltered :one
SELECT COUNT(DISTINCT id)
FROM challenges
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@eventid::text = '' OR event_id = @eventid::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz);
