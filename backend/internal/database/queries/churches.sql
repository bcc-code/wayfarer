-- name: GetChurchesByIDs :many
SELECT id, external_id, name, country, category
FROM churches
WHERE id = ANY(@ids::text[]);

-- name: GetChurchByID :one
SELECT id, external_id, name, country, category
FROM churches
WHERE id = @id;

-- name: GetChurchByExternalID :one
SELECT id, external_id, name, country, category
FROM churches
WHERE external_id = @external_id;

-- name: GetDefaultChurch :one
SELECT id, external_id, name, country, category
FROM churches
WHERE external_id IS NULL
LIMIT 1;

-- name: GetChurchesFilteredCursor :many
SELECT id, external_id, name, country, category
FROM churches
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@country::text = '' OR country = @country::text)
    AND (@category::text = '' OR category = @category::text)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountChurchesFiltered :one
SELECT COUNT(*)
FROM churches
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@country::text = '' OR country = @country::text)
    AND (@category::text = '' OR category = @category::text);

-- name: CreateChurch :one
INSERT INTO churches (id, external_id, name, country, category)
VALUES (@id, @external_id, @name, @country, @category)
RETURNING id, external_id, name, country, category;

-- name: UpsertChurch :one
INSERT INTO churches (id, external_id, name, country, category)
VALUES (@id, @external_id, @name, @country, @category)
ON CONFLICT (external_id)
DO UPDATE SET
    name = EXCLUDED.name,
    country = EXCLUDED.country,
    updated_at = NOW()
RETURNING id, external_id, name, country, category;

-- name: UpdateChurch :one
UPDATE churches
SET
    name = COALESCE(NULLIF(@name::text, ''), name),
    country = COALESCE(NULLIF(@country::text, ''), country),
    category = COALESCE(NULLIF(@category::text, ''), category),
    updated_at = NOW()
WHERE id = @id
RETURNING id, external_id, name, country, category;
