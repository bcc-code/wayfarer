-- +goose Up
ALTER TABLE super_teams ADD COLUMN image_url VARCHAR(500);
ALTER TABLE super_teams ADD COLUMN color VARCHAR(7);

-- +goose Down
ALTER TABLE super_teams DROP COLUMN IF EXISTS image_url;
ALTER TABLE super_teams DROP COLUMN IF EXISTS color;
