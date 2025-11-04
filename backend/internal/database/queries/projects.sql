-- name: GetProjectsByUserIDs :many
SELECT
    p.id,
    p.name,
    p.description,
    p.start_date,
    p.end_date,
    p.logo_url,
    p.color_primary,
    p.color_secondary,
    p.color_tertiary,
    p.rounding,
    up.user_id
FROM projects p
JOIN user_projects up ON p.id = up.project_id
WHERE up.user_id = ANY(@user_ids::text[])
ORDER BY up.user_id, p.start_date DESC;

-- name: GetProjectByID :one
SELECT id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding
FROM projects
WHERE id = @id;

-- name: GetProjectsByIDs :many
SELECT id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding
FROM projects
WHERE id = ANY(@ids::text[]);

-- name: GetAllProjects :many
SELECT id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding
FROM projects
ORDER BY start_date DESC;
