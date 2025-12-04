-- name: UpsertExternalContent :one
INSERT INTO external_content (
    id, plan_id, task_id, content_id, content_type, published_at, synced_at, source
) VALUES (
    @id, @planid::text, @taskid::text, @contentid::text, @contenttype::text, @publishedat::timestamptz, @syncedat::timestamptz, @source::text
)
ON CONFLICT (plan_id, task_id) DO UPDATE SET
    content_id = EXCLUDED.content_id,
    content_type = EXCLUDED.content_type,
    published_at = EXCLUDED.published_at,
    synced_at = EXCLUDED.synced_at,
    source = EXCLUDED.source,
    updated_at = now()
RETURNING *;

-- name: UpsertExternalContentTranslation :exec
INSERT INTO external_content_translations (external_content_id, language_code, title)
VALUES (@externalcontentid, @languagecode::text, @title::text)
ON CONFLICT (external_content_id, language_code) DO UPDATE SET
    title = EXCLUDED.title,
    updated_at = now();

-- name: GetExternalContentByTaskID :one
SELECT * FROM external_content WHERE task_id = @taskid::text;

-- name: GetExternalContentByPlanID :many
SELECT * FROM external_content
WHERE plan_id = @planid::text
ORDER BY published_at ASC NULLS LAST;

-- name: GetExternalContentByContentID :many
SELECT * FROM external_content
WHERE content_id = @contentid::text;

-- name: GetExternalContentByContentType :many
SELECT * FROM external_content
WHERE content_type = @contenttype::text
ORDER BY published_at ASC NULLS LAST
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: GetExternalContentBySource :many
SELECT * FROM external_content
WHERE source = @source::text
ORDER BY published_at ASC NULLS LAST
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: GetExternalContentTranslations :many
SELECT * FROM external_content_translations
WHERE external_content_id = @externalcontentid;

-- name: GetExternalContentTranslation :one
SELECT * FROM external_content_translations
WHERE external_content_id = @externalcontentid AND language_code = @languagecode::text;

-- name: DeleteExternalContentByPlanID :exec
DELETE FROM external_content WHERE plan_id = @planid::text;

-- name: CountExternalContentByPlanID :one
SELECT COUNT(*) FROM external_content WHERE plan_id = @planid::text;

-- name: CountExternalContentBySource :one
SELECT COUNT(*) FROM external_content WHERE source = @source::text;
