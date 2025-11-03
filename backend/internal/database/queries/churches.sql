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
