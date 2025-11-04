-- name: GetUserByID :one
SELECT id, members_id, gender, church_id, birthdate, email, name, avatar_url
FROM users
WHERE id = @id;

-- name: GetUsersByIDs :many
SELECT id, members_id, gender, church_id, birthdate, email, name, avatar_url
FROM users
WHERE id = ANY(@ids::text[]);

-- name: GetUserByMembersID :one
SELECT id, members_id, gender, church_id, birthdate, email, name, avatar_url
FROM users
WHERE members_id = @members_id;

-- name: CreateUser :one
INSERT INTO users (id, members_id, email, name, gender, birthdate, church_id, avatar_url)
VALUES (@id, @members_id, @email, @name, @gender, @birthdate, @church_id, @avatar_url)
RETURNING id, members_id, gender, church_id, birthdate, email, name, avatar_url;
