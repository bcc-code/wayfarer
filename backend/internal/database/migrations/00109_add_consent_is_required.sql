-- +goose Up
-- +goose StatementBegin

-- Per version, not per key: the latest published version of a key decides, so
-- an optional consent is made required by publishing a new version. Contrast
-- project_id, which consentProjectForNewVersion locks across versions.
ALTER TABLE consents
    ADD COLUMN is_required BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN consents.is_required IS 'Consent must be accepted to take part in the project; cannot be rejected';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE consents DROP COLUMN is_required;

-- +goose StatementEnd
