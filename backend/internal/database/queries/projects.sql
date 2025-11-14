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
    p.archived,
    up.user_id
FROM projects p
JOIN user_projects up ON p.id = up.project_id
WHERE up.user_id = ANY(@user_ids::text[])
ORDER BY up.user_id, p.start_date DESC;

-- name: GetProjectByID :one
SELECT id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding, archived
FROM projects
WHERE id = @id;

-- name: GetProjectsByIDs :many
SELECT id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding, archived
FROM projects
WHERE id = ANY(@ids::text[]);

-- name: GetAllProjects :many
SELECT id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding, archived
FROM projects
ORDER BY start_date DESC;

-- name: GetProjectsFilteredCursor :many
SELECT id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding, archived
FROM projects
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@archived::boolean IS NULL OR archived = @archived::boolean)
    AND (@startdateafter::timestamptz IS NULL OR start_date >= @startdateafter::timestamptz)
    AND (@startdatebefore::timestamptz IS NULL OR start_date <= @startdatebefore::timestamptz)
    AND (@enddateafter::timestamptz IS NULL OR end_date >= @enddateafter::timestamptz)
    AND (@enddatebefore::timestamptz IS NULL OR end_date <= @enddatebefore::timestamptz)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountProjectsFiltered :one
SELECT COUNT(id)
FROM projects
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@archived::boolean IS NULL OR archived = @archived::boolean)
    AND (@startdateafter::timestamptz IS NULL OR start_date >= @startdateafter::timestamptz)
    AND (@startdatebefore::timestamptz IS NULL OR start_date <= @startdatebefore::timestamptz)
    AND (@enddateafter::timestamptz IS NULL OR end_date >= @enddateafter::timestamptz)
    AND (@enddatebefore::timestamptz IS NULL OR end_date <= @enddatebefore::timestamptz);

-- name: CreateProject :one
INSERT INTO projects (
    id,
    name,
    description,
    start_date,
    end_date,
    logo_url,
    color_primary,
    color_secondary,
    color_tertiary,
    rounding
)
VALUES (
    @id::text,
    @name::text,
    @description::text,
    @startdate::timestamptz,
    @enddate::timestamptz,
    @logourl::text,
    @colorprimary::text,
    @colorsecondary::text,
    @colortertiary::text,
    @rounding::int
)
RETURNING id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding, archived;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE(@name::text, name),
    description = COALESCE(@description::text, description),
    start_date = COALESCE(@startdate::timestamptz, start_date),
    end_date = COALESCE(@enddate::timestamptz, end_date),
    logo_url = COALESCE(@logourl::text, logo_url),
    color_primary = COALESCE(@colorprimary::text, color_primary),
    color_secondary = COALESCE(@colorsecondary::text, color_secondary),
    color_tertiary = COALESCE(@colortertiary::text, color_tertiary),
    rounding = COALESCE(@rounding::int, rounding),
    updated_at = now()
WHERE id = @id::text
RETURNING id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding, archived;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = @id::text;

-- name: ArchiveProject :one
UPDATE projects
SET
    archived = true,
    updated_at = now()
WHERE id = @id::text
RETURNING id, name, description, start_date, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding, archived;
