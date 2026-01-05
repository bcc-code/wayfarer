-- name: UpsertExternalContent :one
INSERT INTO external_content (
    id, plan_id, task_id, content_id, content_type, published_at, synced_at, source, url
) VALUES (
    @id, @planid::text, @taskid::text, @contentid::text, @contenttype::text, @publishedat::timestamptz, @syncedat::timestamptz, @source::text, @url::text
)
ON CONFLICT (plan_id, task_id) DO UPDATE SET
    content_id = EXCLUDED.content_id,
    content_type = EXCLUDED.content_type,
    published_at = EXCLUDED.published_at,
    synced_at = EXCLUDED.synced_at,
    source = EXCLUDED.source,
    url = EXCLUDED.url,
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

-- ==================== Admin Search Queries ====================

-- name: SearchExternalContentAdmin :many
-- Admin search with filters
SELECT * FROM external_content
WHERE
    (@planid::text = '' OR plan_id = @planid::text)
    AND (@taskid::text = '' OR task_id = @taskid::text)
    AND (@contentid::text = '' OR content_id = @contentid::text)
    AND (@contenttype::text = '' OR content_type = @contenttype::text)
    AND (@source::text = '' OR source = @source::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz)
ORDER BY
    CASE WHEN @sortby::text = 'published_at_desc' THEN published_at END DESC NULLS LAST,
    CASE WHEN @sortby::text = 'published_at_asc' THEN published_at END ASC NULLS LAST,
    CASE WHEN @sortby::text = 'created_at_desc' OR @sortby::text = '' THEN created_at END DESC,
    CASE WHEN @sortby::text = 'created_at_asc' THEN created_at END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: CountExternalContentAdmin :one
-- Count for pagination
SELECT COUNT(*) FROM external_content
WHERE
    (@planid::text = '' OR plan_id = @planid::text)
    AND (@taskid::text = '' OR task_id = @taskid::text)
    AND (@contentid::text = '' OR content_id = @contentid::text)
    AND (@contenttype::text = '' OR content_type = @contenttype::text)
    AND (@source::text = '' OR source = @source::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz);

-- name: GetExternalContentByID :one
SELECT * FROM external_content WHERE id = @id;

-- name: GetExternalContentByIDs :many
SELECT * FROM external_content WHERE id = ANY(@ids::text[]);

-- name: GetExternalContentTranslationsByContentIDs :many
SELECT * FROM external_content_translations
WHERE external_content_id = ANY(@externalcontentids::text[]);
