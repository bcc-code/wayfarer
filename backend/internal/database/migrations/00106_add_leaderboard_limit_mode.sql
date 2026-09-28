-- +goose Up
ALTER TABLE leaderboard_configs
    ADD COLUMN limit_mode TEXT NOT NULL DEFAULT 'MANUAL' CHECK (limit_mode IN ('MANUAL', 'CHURCH_SIZE'));

-- +goose Down
ALTER TABLE leaderboard_configs DROP COLUMN limit_mode;
