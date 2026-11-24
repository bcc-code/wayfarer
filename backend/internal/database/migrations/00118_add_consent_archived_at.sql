-- +goose Up
-- +goose StatementBegin

-- Consents are retired, never deleted: user_consent_history has to outlive them.
ALTER TABLE consents
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_consents_archived ON consents(archived_at) WHERE archived_at IS NOT NULL;

COMMENT ON COLUMN consents.archived_at IS 'When the consent was retired; NULL means active. A key is live only while its latest published version is not archived';

-- Archiving a required consent stops it blocking. Same rule as 00117 otherwise,
-- and it must still match buildConsentStatus and blockingConsentIds, which get
-- it from GetLatestPublishedConsentsForProject. archived_at is tested after the
-- DISTINCT ON, as there: filtering inside would let archiving v2 make v1 the
-- blocking version again.
CREATE OR REPLACE FUNCTION consent_blocked_user_ids(p_project_id CHAR(28))
RETURNS TABLE (user_id CHAR(28))
LANGUAGE sql STABLE AS $$
    WITH latest_consents AS (
        SELECT DISTINCT ON (c.key) c.id, c.key, c.is_required, c.is_remote, c.archived_at
        FROM consents c
        WHERE c.published_at IS NOT NULL
          AND c.published_at <= now()
          AND (c.project_id IS NULL OR c.project_id = p_project_id)
        ORDER BY c.key, c.version DESC
    ),
    required_versions AS (
        SELECT id, key FROM latest_consents
        WHERE is_required AND NOT is_remote
          AND archived_at IS NULL
    )
    SELECT up.user_id
    FROM user_projects up
    INNER JOIN users u ON u.id = up.user_id
    WHERE up.project_id = p_project_id
      AND EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate) >= 13
      AND EXISTS (
          SELECT 1
          FROM required_versions rv
          WHERE (
              SELECT uch.action
              FROM user_consent_history uch
              WHERE uch.user_id = up.user_id
                AND uch.consent_key = rv.key
                AND uch.consent_id = rv.id
              ORDER BY uch.occurred_at DESC
              LIMIT 1
          ) IS DISTINCT FROM 'ACCEPTED'
      )
$$;

COMMENT ON FUNCTION consent_blocked_user_ids(CHAR(28)) IS 'Users of the project who have not accepted the latest version of a required, unarchived consent; hidden from other users';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Back to 00117, before the column it reads goes away
CREATE OR REPLACE FUNCTION consent_blocked_user_ids(p_project_id CHAR(28))
RETURNS TABLE (user_id CHAR(28))
LANGUAGE sql STABLE AS $$
    WITH latest_consents AS (
        SELECT DISTINCT ON (c.key) c.id, c.key, c.is_required, c.is_remote
        FROM consents c
        WHERE c.published_at IS NOT NULL
          AND c.published_at <= now()
          AND (c.project_id IS NULL OR c.project_id = p_project_id)
        ORDER BY c.key, c.version DESC
    ),
    required_versions AS (
        SELECT id, key FROM latest_consents
        WHERE is_required AND NOT is_remote
    )
    SELECT up.user_id
    FROM user_projects up
    INNER JOIN users u ON u.id = up.user_id
    WHERE up.project_id = p_project_id
      AND EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate) >= 13
      AND EXISTS (
          SELECT 1
          FROM required_versions rv
          WHERE (
              SELECT uch.action
              FROM user_consent_history uch
              WHERE uch.user_id = up.user_id
                AND uch.consent_key = rv.key
                AND uch.consent_id = rv.id
              ORDER BY uch.occurred_at DESC
              LIMIT 1
          ) IS DISTINCT FROM 'ACCEPTED'
      )
$$;

COMMENT ON FUNCTION consent_blocked_user_ids(CHAR(28)) IS 'Users of the project who have not accepted the latest version of a required consent; hidden from other users';

DROP INDEX IF EXISTS idx_consents_archived;
ALTER TABLE consents DROP COLUMN IF EXISTS archived_at;

-- +goose StatementEnd
