-- +goose Up
-- +goose StatementBegin
ALTER TABLE consents ADD COLUMN short_text TEXT NOT NULL DEFAULT '';
ALTER TABLE consent_translations ADD COLUMN short_text TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE consents DROP COLUMN short_text;
ALTER TABLE consent_translations DROP COLUMN short_text;
-- +goose StatementEnd
