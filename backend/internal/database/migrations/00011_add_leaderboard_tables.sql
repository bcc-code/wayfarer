-- +goose Up
-- +goose StatementBegin

-- Migration: Add materialized leaderboard tables
-- These tables store pre-computed scores for fast leaderboard queries
-- Maintained by database triggers on achievement and score_adjustment changes

-- ==================== Project Leaderboards ====================

CREATE TABLE leaderboard_project_persons (
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX idx_leaderboard_project_persons_score
    ON leaderboard_project_persons(project_id, score DESC, user_id);

CREATE TABLE leaderboard_project_teams (
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    team_id CHAR(28) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, team_id)
);

CREATE INDEX idx_leaderboard_project_teams_score
    ON leaderboard_project_teams(project_id, score DESC, team_id);

CREATE TABLE leaderboard_project_superteams (
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    super_team_id CHAR(28) NOT NULL REFERENCES super_teams(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, super_team_id)
);

CREATE INDEX idx_leaderboard_project_superteams_score
    ON leaderboard_project_superteams(project_id, score DESC, super_team_id);

CREATE TABLE leaderboard_project_churches (
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    church_id CHAR(28) NOT NULL REFERENCES churches(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, church_id)
);

CREATE INDEX idx_leaderboard_project_churches_score
    ON leaderboard_project_churches(project_id, score DESC, church_id);

-- ==================== Event Leaderboards ====================

CREATE TABLE leaderboard_event_persons (
    event_id CHAR(28) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, user_id)
);

CREATE INDEX idx_leaderboard_event_persons_score
    ON leaderboard_event_persons(event_id, score DESC, user_id);

CREATE TABLE leaderboard_event_teams (
    event_id CHAR(28) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    team_id CHAR(28) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, team_id)
);

CREATE INDEX idx_leaderboard_event_teams_score
    ON leaderboard_event_teams(event_id, score DESC, team_id);

CREATE TABLE leaderboard_event_superteams (
    event_id CHAR(28) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    super_team_id CHAR(28) NOT NULL REFERENCES super_teams(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, super_team_id)
);

CREATE INDEX idx_leaderboard_event_superteams_score
    ON leaderboard_event_superteams(event_id, score DESC, super_team_id);

CREATE TABLE leaderboard_event_churches (
    event_id CHAR(28) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    church_id CHAR(28) NOT NULL REFERENCES churches(id) ON DELETE CASCADE,
    score BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, church_id)
);

CREATE INDEX idx_leaderboard_event_churches_score
    ON leaderboard_event_churches(event_id, score DESC, church_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop tables in reverse order (will cascade to indexes)
DROP TABLE IF EXISTS leaderboard_event_churches;
DROP TABLE IF EXISTS leaderboard_event_superteams;
DROP TABLE IF EXISTS leaderboard_event_teams;
DROP TABLE IF EXISTS leaderboard_event_persons;
DROP TABLE IF EXISTS leaderboard_project_churches;
DROP TABLE IF EXISTS leaderboard_project_superteams;
DROP TABLE IF EXISTS leaderboard_project_teams;
DROP TABLE IF EXISTS leaderboard_project_persons;

-- +goose StatementEnd
