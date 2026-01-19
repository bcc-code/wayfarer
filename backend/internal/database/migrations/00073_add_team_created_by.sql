-- +goose Up
ALTER TABLE teams ADD COLUMN created_by_user_id CHAR(28) REFERENCES users(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE teams DROP COLUMN created_by_user_id;
