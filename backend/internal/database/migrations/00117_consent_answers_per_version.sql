-- +goose Up
-- +goose StatementBegin

-- Answers are per consent version, not per key: a user has not consented to a
-- version they have not seen, so a new version needs a new answer, even from
-- users who accepted (or rejected) the previous one.
--
-- Same rule as 00116 otherwise; it must match buildConsentStatus
-- (loaders/user.go) and blockingConsentIds (frontend/app/utils/consentGate.ts):
--   * consents: latest published version per key, global or this project's.
--     is_required is per version, so it is tested after picking the latest.
--   * remote consents never block: they cannot be accepted in the app.
--   * blocked = the user's latest action on that version is missing or
--     REJECTED. consent_key is matched too, for the (user_id, consent_key,
--     occurred_at) index.
--   * under-13s never block, because local consents are not shown to them.
--     Age is the year difference, as in calculateAge.
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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Back to 00116: answers per key
CREATE OR REPLACE FUNCTION consent_blocked_user_ids(p_project_id CHAR(28))
RETURNS TABLE (user_id CHAR(28))
LANGUAGE sql STABLE AS $$
    WITH latest_consents AS (
        SELECT DISTINCT ON (c.key) c.key, c.is_required, c.is_remote
        FROM consents c
        WHERE c.published_at IS NOT NULL
          AND c.published_at <= now()
          AND (c.project_id IS NULL OR c.project_id = p_project_id)
        ORDER BY c.key, c.version DESC
    ),
    required_keys AS (
        SELECT key FROM latest_consents
        WHERE is_required AND NOT is_remote
    )
    SELECT up.user_id
    FROM user_projects up
    INNER JOIN users u ON u.id = up.user_id
    WHERE up.project_id = p_project_id
      AND EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate) >= 13
      AND EXISTS (
          SELECT 1
          FROM required_keys rk
          WHERE (
              SELECT uch.action
              FROM user_consent_history uch
              WHERE uch.user_id = up.user_id
                AND uch.consent_key = rk.key
              ORDER BY uch.occurred_at DESC
              LIMIT 1
          ) IS DISTINCT FROM 'ACCEPTED'
      )
$$;

COMMENT ON FUNCTION consent_blocked_user_ids(CHAR(28)) IS 'Users of the project who have not accepted a required consent; hidden from other users';

-- +goose StatementEnd
