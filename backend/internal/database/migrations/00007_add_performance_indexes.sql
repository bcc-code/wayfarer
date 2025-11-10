-- +goose Up
-- +goose StatementBegin

-- Add performance indexes based on CockroachDB recommendations
-- These indexes optimize common query patterns identified in production

-- Projects indexes for filtering and sorting
CREATE INDEX IF NOT EXISTS idx_projects_end_date ON projects (end_date)
    STORING (name, description, start_date, logo_url, color_primary, color_secondary, color_tertiary, archived, rounding);

CREATE INDEX IF NOT EXISTS idx_projects_start_date_desc ON projects (start_date DESC)
    STORING (name, description, end_date, logo_url, color_primary, color_secondary, color_tertiary, rounding);

-- Churches index for country filtering
CREATE INDEX IF NOT EXISTS idx_churches_country ON churches (country)
    STORING (name, category, external_id);

-- Achievements index for project queries
CREATE INDEX IF NOT EXISTS idx_achievements_created_at_desc ON achievements (created_at DESC)
    STORING (achievement_type, project_id, event_id, challenge_id, name, description, image_url, points, hidden, updated_at);

-- Events index for project and date-based queries
CREATE INDEX IF NOT EXISTS idx_events_start_date_desc ON events (start_date DESC)
    STORING (project_id, name, description, end_date, created_at, updated_at);

-- User streak activity index for streak lookups
CREATE INDEX IF NOT EXISTS idx_user_streak_activity_streak_date ON user_streak_activity (streak_id, activity_date DESC)
    STORING (created_at);

-- Streak relevant days index
CREATE INDEX IF NOT EXISTS idx_streak_relevant_days_streak_start ON streak_relevant_days (streak_id, start_date)
    STORING (end_date);

-- Challenges indexes for project and event filtering
CREATE INDEX IF NOT EXISTS idx_challenges_project_published ON challenges (project_id, published_at)
    STORING (event_id, name, description, image_url, url, button_text, end_time, created_at, updated_at);

CREATE INDEX IF NOT EXISTS idx_challenges_event_published ON challenges (event_id, published_at DESC)
    STORING (project_id, name, description, image_url, url, button_text, end_time, created_at, updated_at);

-- Teams index for project queries
CREATE INDEX IF NOT EXISTS idx_teams_created_at_desc ON teams (created_at DESC)
    STORING (project_id, name, description, join_code, super_team_id, updated_at);

-- Streaks index for project queries
CREATE INDEX IF NOT EXISTS idx_streaks_created_at_desc ON streaks (created_at DESC)
    STORING (project_id, name, description, updated_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove all performance indexes
DROP INDEX IF EXISTS idx_projects_end_date;
DROP INDEX IF EXISTS idx_projects_start_date_desc;
DROP INDEX IF EXISTS idx_churches_country;
DROP INDEX IF EXISTS idx_achievements_created_at_desc;
DROP INDEX IF EXISTS idx_events_start_date_desc;
DROP INDEX IF EXISTS idx_user_streak_activity_streak_date;
DROP INDEX IF EXISTS idx_streak_relevant_days_streak_start;
DROP INDEX IF EXISTS idx_challenges_project_published;
DROP INDEX IF EXISTS idx_challenges_event_published;
DROP INDEX IF EXISTS idx_teams_created_at_desc;
DROP INDEX IF EXISTS idx_streaks_created_at_desc;

-- +goose StatementEnd
