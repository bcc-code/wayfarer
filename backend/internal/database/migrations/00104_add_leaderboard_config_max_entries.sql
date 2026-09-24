-- +goose Up
ALTER TABLE leaderboard_configs
    ADD COLUMN max_entries INT CHECK (max_entries > 0);

-- +goose Down
ALTER TABLE leaderboard_configs DROP COLUMN max_entries;
