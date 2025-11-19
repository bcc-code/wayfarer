-- +goose Up
-- +goose StatementBegin

-- Drop team and superteam achievement tables
-- Achievements are now only awarded to individual users
-- When awarding to a team/superteam, each member receives the achievement individually

DROP TABLE IF EXISTS team_achievements;
DROP TABLE IF EXISTS super_team_achievements;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate team_achievements table
CREATE TABLE IF NOT EXISTS team_achievements (
    team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    achievement_id TEXT NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (team_id, achievement_id)
);

CREATE INDEX idx_team_achievements_team ON team_achievements(team_id);
CREATE INDEX idx_team_achievements_achievement ON team_achievements(achievement_id);

-- Recreate super_team_achievements table
CREATE TABLE IF NOT EXISTS super_team_achievements (
    super_team_id TEXT NOT NULL REFERENCES super_teams(id) ON DELETE CASCADE,
    achievement_id TEXT NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (super_team_id, achievement_id)
);

CREATE INDEX idx_super_team_achievements_super_team ON super_team_achievements(super_team_id);
CREATE INDEX idx_super_team_achievements_achievement ON super_team_achievements(achievement_id);

-- +goose StatementEnd
