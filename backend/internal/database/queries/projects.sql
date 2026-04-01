-- name: GetProjectsByUserIDs :many
SELECT
    p.id,
    p.name,
    p.description,
    p.rules,
    p.info_message,
    p.info_message_start,
    p.info_message_end,
    p.start_date,
    p.end_date,
    p.logo_url,
    p.banner_url,
    p.color_light_accent,
    p.color_light_accent_contrast,
    p.color_light_on_accent,
    p.color_light_background_default,
    p.color_light_background_raised,
    p.color_light_background_indent,
    p.color_light_text_default,
    p.color_light_text_muted,
    p.color_light_text_hint,
    p.color_light_shadow_default,
    p.color_light_shadow_blank,
    p.color_light_border_default,
    p.color_dark_accent,
    p.color_dark_accent_contrast,
    p.color_dark_on_accent,
    p.color_dark_background_default,
    p.color_dark_background_raised,
    p.color_dark_background_indent,
    p.color_dark_text_default,
    p.color_dark_text_muted,
    p.color_dark_text_hint,
    p.color_dark_shadow_default,
    p.color_dark_shadow_blank,
    p.color_dark_border_default,
    p.rounding,
    p.archived,
    up.user_id
FROM projects p
JOIN user_projects up ON p.id = up.project_id
WHERE up.user_id = ANY(@user_ids::text[])
ORDER BY up.user_id, p.start_date DESC;

-- name: GetProjectByID :one
SELECT id, name, description, rules, info_message, info_message_start, info_message_end, start_date, end_date, logo_url, banner_url,
    color_light_accent, color_light_accent_contrast, color_light_on_accent,
    color_light_background_default, color_light_background_raised, color_light_background_indent,
    color_light_text_default, color_light_text_muted, color_light_text_hint,
    color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
    color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
    color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
    color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
    color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
    rounding, archived
FROM projects
WHERE id = @id;

-- name: GetProjectsByIDs :many
SELECT id, name, description, rules, info_message, info_message_start, info_message_end, start_date, end_date, logo_url, banner_url,
    color_light_accent, color_light_accent_contrast, color_light_on_accent,
    color_light_background_default, color_light_background_raised, color_light_background_indent,
    color_light_text_default, color_light_text_muted, color_light_text_hint,
    color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
    color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
    color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
    color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
    color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
    rounding, archived
FROM projects
WHERE id = ANY(@ids::text[]);

-- name: GetAllProjects :many
SELECT id, name, description, rules, info_message, info_message_start, info_message_end, start_date, end_date, logo_url, banner_url,
    color_light_accent, color_light_accent_contrast, color_light_on_accent,
    color_light_background_default, color_light_background_raised, color_light_background_indent,
    color_light_text_default, color_light_text_muted, color_light_text_hint,
    color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
    color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
    color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
    color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
    color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
    rounding, archived
FROM projects
ORDER BY start_date DESC;

-- name: GetProjectsFilteredCursor :many
SELECT id, name, description, rules, info_message, info_message_start, info_message_end, start_date, end_date, logo_url, banner_url,
    color_light_accent, color_light_accent_contrast, color_light_on_accent,
    color_light_background_default, color_light_background_raised, color_light_background_indent,
    color_light_text_default, color_light_text_muted, color_light_text_hint,
    color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
    color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
    color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
    color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
    color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
    rounding, archived
FROM projects
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (sqlc.narg('archived')::boolean IS NULL OR archived = sqlc.narg('archived')::boolean)
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
    AND (sqlc.narg('archived')::boolean IS NULL OR archived = sqlc.narg('archived')::boolean)
    AND (@startdateafter::timestamptz IS NULL OR start_date >= @startdateafter::timestamptz)
    AND (@startdatebefore::timestamptz IS NULL OR start_date <= @startdatebefore::timestamptz)
    AND (@enddateafter::timestamptz IS NULL OR end_date >= @enddateafter::timestamptz)
    AND (@enddatebefore::timestamptz IS NULL OR end_date <= @enddatebefore::timestamptz);

-- name: CreateProject :one
INSERT INTO projects (
    id,
    name,
    description,
    rules,
    info_message,
    info_message_start,
    info_message_end,
    start_date,
    end_date,
    logo_url,
    banner_url,
    color_light_accent,
    color_light_accent_contrast,
    color_light_on_accent,
    color_light_background_default,
    color_light_background_raised,
    color_light_background_indent,
    color_light_text_default,
    color_light_text_muted,
    color_light_text_hint,
    color_light_shadow_default,
    color_light_shadow_blank,
    color_light_border_default,
    color_dark_accent,
    color_dark_accent_contrast,
    color_dark_on_accent,
    color_dark_background_default,
    color_dark_background_raised,
    color_dark_background_indent,
    color_dark_text_default,
    color_dark_text_muted,
    color_dark_text_hint,
    color_dark_shadow_default,
    color_dark_shadow_blank,
    color_dark_border_default,
    rounding
)
VALUES (
    @id::text,
    @name::text,
    @description::text,
    sqlc.narg('rules')::text,
    sqlc.narg('infomessage')::text,
    sqlc.narg('infomessagestart')::timestamptz,
    sqlc.narg('infomessageend')::timestamptz,
    @startdate::timestamptz,
    @enddate::timestamptz,
    sqlc.narg('logourl')::text,
    sqlc.narg('bannerurl')::text,
    @colorlightaccent::text,
    @colorlightaccentcontrast::text,
    @colorlightonaccent::text,
    @colorlightbackgrounddefault::text,
    @colorlightbackgroundraised::text,
    @colorlightbackgroundindent::text,
    @colorlighttextdefault::text,
    @colorlighttextmuted::text,
    @colorlighttexthint::text,
    @colorlightshadowdefault::text,
    @colorlightshadowblank::text,
    @colorlightborderdefault::text,
    @colordarkaccent::text,
    @colordarkaccentcontrast::text,
    @colordarkonaccent::text,
    @colordarkbackgrounddefault::text,
    @colordarkbackgroundraised::text,
    @colordarkbackgroundindent::text,
    @colordarktextdefault::text,
    @colordarktextmuted::text,
    @colordarktexthint::text,
    @colordarkshadowdefault::text,
    @colordarkshadowblank::text,
    @colordarkborderdefault::text,
    @rounding::int
)
RETURNING id, name, description, rules, info_message, info_message_start, info_message_end, start_date, end_date, logo_url, banner_url,
    color_light_accent, color_light_accent_contrast, color_light_on_accent,
    color_light_background_default, color_light_background_raised, color_light_background_indent,
    color_light_text_default, color_light_text_muted, color_light_text_hint,
    color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
    color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
    color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
    color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
    color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
    rounding, archived;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description),
    rules = COALESCE(sqlc.narg('rules')::text, rules),
    info_message = COALESCE(sqlc.narg('infomessage')::text, info_message),
    info_message_start = COALESCE(sqlc.narg('infomessagestart')::timestamptz, info_message_start),
    info_message_end = COALESCE(sqlc.narg('infomessageend')::timestamptz, info_message_end),
    start_date = COALESCE(sqlc.narg('startdate')::timestamptz, start_date),
    end_date = COALESCE(sqlc.narg('enddate')::timestamptz, end_date),
    logo_url = CASE
        WHEN sqlc.narg('logourl')::text = '' THEN NULL
        ELSE COALESCE(sqlc.narg('logourl')::text, logo_url)
    END,
    banner_url = CASE
        WHEN sqlc.narg('bannerurl')::text = '' THEN NULL
        ELSE COALESCE(sqlc.narg('bannerurl')::text, banner_url)
    END,
    color_light_accent = COALESCE(sqlc.narg('colorlightaccent')::text, color_light_accent),
    color_light_accent_contrast = COALESCE(sqlc.narg('colorlightaccentcontrast')::text, color_light_accent_contrast),
    color_light_on_accent = COALESCE(sqlc.narg('colorlightonaccent')::text, color_light_on_accent),
    color_light_background_default = COALESCE(sqlc.narg('colorlightbackgrounddefault')::text, color_light_background_default),
    color_light_background_raised = COALESCE(sqlc.narg('colorlightbackgroundraised')::text, color_light_background_raised),
    color_light_background_indent = COALESCE(sqlc.narg('colorlightbackgroundindent')::text, color_light_background_indent),
    color_light_text_default = COALESCE(sqlc.narg('colorlighttextdefault')::text, color_light_text_default),
    color_light_text_muted = COALESCE(sqlc.narg('colorlighttextmuted')::text, color_light_text_muted),
    color_light_text_hint = COALESCE(sqlc.narg('colorlighttexthint')::text, color_light_text_hint),
    color_light_shadow_default = COALESCE(sqlc.narg('colorlightshadowdefault')::text, color_light_shadow_default),
    color_light_shadow_blank = COALESCE(sqlc.narg('colorlightshadowblank')::text, color_light_shadow_blank),
    color_light_border_default = COALESCE(sqlc.narg('colorlightborderdefault')::text, color_light_border_default),
    color_dark_accent = COALESCE(sqlc.narg('colordarkaccent')::text, color_dark_accent),
    color_dark_accent_contrast = COALESCE(sqlc.narg('colordarkaccentcontrast')::text, color_dark_accent_contrast),
    color_dark_on_accent = COALESCE(sqlc.narg('colordarkonaccent')::text, color_dark_on_accent),
    color_dark_background_default = COALESCE(sqlc.narg('colordarkbackgrounddefault')::text, color_dark_background_default),
    color_dark_background_raised = COALESCE(sqlc.narg('colordarkbackgroundraised')::text, color_dark_background_raised),
    color_dark_background_indent = COALESCE(sqlc.narg('colordarkbackgroundindent')::text, color_dark_background_indent),
    color_dark_text_default = COALESCE(sqlc.narg('colordarktextdefault')::text, color_dark_text_default),
    color_dark_text_muted = COALESCE(sqlc.narg('colordarktextmuted')::text, color_dark_text_muted),
    color_dark_text_hint = COALESCE(sqlc.narg('colordarktexthint')::text, color_dark_text_hint),
    color_dark_shadow_default = COALESCE(sqlc.narg('colordarkshadowdefault')::text, color_dark_shadow_default),
    color_dark_shadow_blank = COALESCE(sqlc.narg('colordarkshadowblank')::text, color_dark_shadow_blank),
    color_dark_border_default = COALESCE(sqlc.narg('colordarkborderdefault')::text, color_dark_border_default),
    rounding = COALESCE(sqlc.narg('rounding')::int, rounding),
    updated_at = now()
WHERE id = @id::text
RETURNING id, name, description, rules, info_message, info_message_start, info_message_end, start_date, end_date, logo_url, banner_url,
    color_light_accent, color_light_accent_contrast, color_light_on_accent,
    color_light_background_default, color_light_background_raised, color_light_background_indent,
    color_light_text_default, color_light_text_muted, color_light_text_hint,
    color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
    color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
    color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
    color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
    color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
    rounding, archived;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = @id::text;

-- name: ArchiveProject :one
UPDATE projects
SET
    archived = true,
    updated_at = now()
WHERE id = @id::text
RETURNING id, name, description, rules, info_message, info_message_start, info_message_end, start_date, end_date, logo_url, banner_url,
    color_light_accent, color_light_accent_contrast, color_light_on_accent,
    color_light_background_default, color_light_background_raised, color_light_background_indent,
    color_light_text_default, color_light_text_muted, color_light_text_hint,
    color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
    color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
    color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
    color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
    color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
    rounding, archived;

-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM projects WHERE id = @projectid::text);
