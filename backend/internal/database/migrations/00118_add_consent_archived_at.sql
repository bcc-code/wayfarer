-- +goose Up
-- +goose StatementBegin

-- Consents are retired, never deleted: user_consent_history has to outlive them.
ALTER TABLE consents
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_consents_archived ON consents(archived_at) WHERE archived_at IS NOT NULL;

COMMENT ON COLUMN consents.archived_at IS 'When the consent was retired; NULL means active. A key is live only while its latest published version is not archived';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_consents_archived;
ALTER TABLE consents DROP COLUMN IF EXISTS archived_at;

-- +goose StatementEnd
