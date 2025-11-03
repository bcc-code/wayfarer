-- name: GetChurchesByIDs :many
SELECT id, name, country, category
FROM churches
WHERE id = ANY(@ids::text[]);

-- name: GetChurchByID :one
SELECT id, name, country, category
FROM churches
WHERE id = @id;
