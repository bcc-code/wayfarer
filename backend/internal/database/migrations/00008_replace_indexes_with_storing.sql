-- +goose Up
-- +goose StatementBegin

-- Replace existing indexes with improved versions that include STORING clauses
-- This improves query performance by avoiding additional lookups

-- Drop old indexes
DROP INDEX IF EXISTS idx_listening_tracks_achievement;
DROP INDEX IF EXISTS idx_reading_articles_achievement;
DROP INDEX IF EXISTS idx_teams_project;
DROP INDEX IF EXISTS idx_super_teams_project;
DROP INDEX IF EXISTS idx_achievements_project;
DROP INDEX IF EXISTS idx_streaks_project;

-- Create replacement indexes with STORING clauses

-- listening_achievement_tracks: Store track details with achievement_id index
CREATE INDEX idx_listening_tracks_achievement ON listening_achievement_tracks (achievement_id)
    STORING (track_id, name, description, image_url);

-- reading_achievement_articles: Store article details with achievement_id index
CREATE INDEX idx_reading_articles_achievement ON reading_achievement_articles (achievement_id)
    STORING (article_id, title, author, url);

-- teams: Store team details with project_id index
CREATE INDEX idx_teams_project ON teams (project_id)
    STORING (name, description, join_code, super_team_id, created_at, updated_at);

-- super_teams: Store super team details with project_id index
CREATE INDEX idx_super_teams_project ON super_teams (project_id)
    STORING (name, description, created_at, updated_at);

-- achievements: Store achievement details with project_id index
CREATE INDEX idx_achievements_project ON achievements (project_id)
    STORING (achievement_type, event_id, challenge_id, name, description, image_url, points, hidden, created_at, updated_at);

-- streaks: Store streak details with project_id index
CREATE INDEX idx_streaks_project ON streaks (project_id)
    STORING (name, description, created_at, updated_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to simple indexes without STORING clauses
DROP INDEX IF EXISTS idx_listening_tracks_achievement;
DROP INDEX IF EXISTS idx_reading_articles_achievement;
DROP INDEX IF EXISTS idx_teams_project;
DROP INDEX IF EXISTS idx_super_teams_project;
DROP INDEX IF EXISTS idx_achievements_project;
DROP INDEX IF EXISTS idx_streaks_project;

-- Recreate original simple indexes
CREATE INDEX idx_listening_tracks_achievement ON listening_achievement_tracks(achievement_id);
CREATE INDEX idx_reading_articles_achievement ON reading_achievement_articles(achievement_id);
CREATE INDEX idx_teams_project ON teams(project_id);
CREATE INDEX idx_super_teams_project ON super_teams(project_id);
CREATE INDEX idx_achievements_project ON achievements(project_id);
CREATE INDEX idx_streaks_project ON streaks(project_id);

-- +goose StatementEnd
