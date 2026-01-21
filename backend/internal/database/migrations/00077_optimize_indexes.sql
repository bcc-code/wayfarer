-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- PostgreSQL Query Optimization - Index Changes
-- ============================================================================
-- This migration adds new indexes to improve query performance and drops
-- unused indexes that waste space and slow down writes.
--
-- Key improvements:
-- 1. Better consent history lookups (100% seq scans -> index usage)
-- 2. Covering index for teams table (reduce table lookups)
-- 3. Composite indexes for challenges and achievements
-- 4. Removal of ~43MB of unused indexes
-- ============================================================================

-- ----------------------------------------------------------------------------
-- NEW INDEXES
-- ----------------------------------------------------------------------------

-- 1. Better consent history lookups
-- The existing idx_user_consent_history_user_key_time doesn't include action,
-- causing 100% sequential scans on this table
CREATE INDEX IF NOT EXISTS idx_user_consent_history_user_key_action_time
ON user_consent_history (user_id, consent_key, action, occurred_at DESC);

-- 2. Covering index for teams (reduces table lookups in leaderboard queries)
-- The teams table has 1.8 billion rows read via sequential scan
CREATE INDEX IF NOT EXISTS idx_teams_project_covering
ON teams (project_id)
INCLUDE (id, name, super_team_id, leaderboard_excluded);

-- 3. Challenges composite for filtered queries
CREATE INDEX IF NOT EXISTS idx_challenges_project_published_composite
ON challenges (project_id, published_at DESC NULLS LAST)
WHERE published_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_challenges_event_published_composite
ON challenges (event_id, published_at DESC NULLS LAST)
WHERE event_id IS NOT NULL AND published_at IS NOT NULL;

-- 4. Achievements for filtered listing
CREATE INDEX IF NOT EXISTS idx_achievements_project_hidden_sort
ON achievements (project_id, hidden, sort_order, created_at DESC);

-- ----------------------------------------------------------------------------
-- DROP UNUSED INDEXES
-- ----------------------------------------------------------------------------

-- idx_user_projects_project_user: 37 MB, 1 use - duplicate of PK
DROP INDEX IF EXISTS idx_user_projects_project_user;

-- NOTE: idx_user_projects_project (3.2 MB, 59 uses) is NOT dropped.
-- Benchmarking showed dropping it causes 5x regression on UserProjectsJoin queries.

-- idx_users_display_name: 912 KB, never used
DROP INDEX IF EXISTS idx_users_display_name;

-- idx_users_first_name: 432 KB, never used
DROP INDEX IF EXISTS idx_users_first_name;

-- idx_users_last_name: 296 KB, never used
DROP INDEX IF EXISTS idx_users_last_name;

-- idx_users_birthdate: 600 KB, 5 uses total
DROP INDEX IF EXISTS idx_users_birthdate;

-- idx_achievements_challenge: 16 KB, never used
DROP INDEX IF EXISTS idx_achievements_challenge;

-- idx_user_achievements_time: 40 KB, never used
DROP INDEX IF EXISTS idx_user_achievements_time;

-- idx_super_teams_project: 16 KB, never used
DROP INDEX IF EXISTS idx_super_teams_project;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate dropped indexes
CREATE INDEX IF NOT EXISTS idx_user_projects_project_user ON user_projects (project_id, user_id);
CREATE INDEX IF NOT EXISTS idx_users_display_name ON users (display_name);
CREATE INDEX IF NOT EXISTS idx_users_first_name ON users (first_name);
CREATE INDEX IF NOT EXISTS idx_users_last_name ON users (last_name);
CREATE INDEX IF NOT EXISTS idx_users_birthdate ON users (birthdate);
CREATE INDEX IF NOT EXISTS idx_achievements_challenge ON achievements (challenge_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_time ON user_achievements (achieved_at);
CREATE INDEX IF NOT EXISTS idx_super_teams_project ON super_teams (project_id);

-- Drop new indexes
DROP INDEX IF EXISTS idx_user_consent_history_user_key_action_time;
DROP INDEX IF EXISTS idx_teams_project_covering;
DROP INDEX IF EXISTS idx_challenges_project_published_composite;
DROP INDEX IF EXISTS idx_challenges_event_published_composite;
DROP INDEX IF EXISTS idx_achievements_project_hidden_sort;

-- +goose StatementEnd
