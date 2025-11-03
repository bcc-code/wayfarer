-- name: GetUserByID :one
SELECT id, members_id, gender, church_id, age, email, name, avatar_url
FROM users
WHERE id = @id;

-- name: GetUsersByIDs :many
SELECT id, members_id, gender, church_id, age, email, name, avatar_url
FROM users
WHERE id = ANY(@ids::text[]);
