-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN info_message TEXT;
ALTER TABLE projects ADD COLUMN info_message_start TIMESTAMPTZ;
ALTER TABLE projects ADD COLUMN info_message_end TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN IF EXISTS info_message_end;
ALTER TABLE projects DROP COLUMN IF EXISTS info_message_start;
ALTER TABLE projects DROP COLUMN IF EXISTS info_message;
-- +goose StatementEnd
