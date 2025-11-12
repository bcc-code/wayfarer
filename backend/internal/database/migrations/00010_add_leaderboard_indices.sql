-- +goose Up
-- +goose StatementBegin

-- Index to support project-first lookups in leaderboards
-- This is the most critical index for leaderboard performance
-- Allows filtering users by project_id first, reducing scan size from all users to project users
CREATE INDEX IF NOT EXISTS idx_user_projects_project_user
ON user_projects(project_id, user_id);

-- Index to support efficient achievement lookups with points
-- INCLUDE clause creates a covering index so points can be retrieved without table access
CREATE INDEX IF NOT EXISTS idx_achievements_project_points
ON achievements(project_id) INCLUDE (id, points);

-- Composite index for score adjustments filtered by project and entity type
-- INCLUDE points to avoid table lookups
CREATE INDEX IF NOT EXISTS idx_score_adjustments_project_entity_type
ON score_adjustments(project_id, entity_type, entity_id) INCLUDE (points);

-- Index to support age-based filtering (optional but helpful)
-- Partial index excludes NULL birthdates
CREATE INDEX IF NOT EXISTS idx_users_birthdate
ON users(birthdate) WHERE birthdate IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indices in reverse order
DROP INDEX IF EXISTS idx_users_birthdate;
DROP INDEX IF EXISTS idx_score_adjustments_project_entity_type;
DROP INDEX IF EXISTS idx_achievements_project_points;
DROP INDEX IF EXISTS idx_user_projects_project_user;

-- +goose StatementEnd
