-- +goose Up
ALTER TABLE users ADD COLUMN church_locked_until TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users DROP COLUMN church_locked_until;
