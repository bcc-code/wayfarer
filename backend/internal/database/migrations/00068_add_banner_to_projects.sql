-- +goose Up
ALTER TABLE projects ADD COLUMN banner_url VARCHAR(500);

-- +goose Down
ALTER TABLE projects DROP COLUMN IF EXISTS banner_url;
