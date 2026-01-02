-- +goose Up
-- +goose StatementBegin
ALTER TABLE consents ADD COLUMN button_text VARCHAR(100);
ALTER TABLE consent_translations ADD COLUMN button_text VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE consents DROP COLUMN button_text;
ALTER TABLE consent_translations DROP COLUMN button_text;
-- +goose StatementEnd
