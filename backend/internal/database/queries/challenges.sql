-- name: GetChallengesByIDs :many
SELECT id, project_id, event_id, name, description, image_url, url, button_text, published_at, end_time, created_at, updated_at
FROM challenges
WHERE id = ANY(@ids::text[]);
