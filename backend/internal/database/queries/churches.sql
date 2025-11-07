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
