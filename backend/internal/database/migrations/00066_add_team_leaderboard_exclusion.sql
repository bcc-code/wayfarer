-- +goose Up
ALTER TABLE teams
ADD COLUMN leaderboard_excluded BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_teams_leaderboard_excluded ON teams(leaderboard_excluded) WHERE leaderboard_excluded = true;

-- +goose Down
DROP INDEX IF EXISTS idx_teams_leaderboard_excluded;
ALTER TABLE teams DROP COLUMN IF EXISTS leaderboard_excluded;
