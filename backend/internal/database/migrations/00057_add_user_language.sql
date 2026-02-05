-- +goose Up
ALTER TABLE users ADD COLUMN language VARCHAR(10) NOT NULL DEFAULT 'no';

-- +goose Down
ALTER TABLE users DROP COLUMN language;
