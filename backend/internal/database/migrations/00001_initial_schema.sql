-- +goose Up
-- +goose StatementBegin

-- ==================== Core Tables ====================
-- All IDs use ULID format with 2-character table prefix
-- Format: XX + 26-character ULID = 28 characters total
-- Example: CH01ARZ3NDEKTSV4RRFFQ69G5FAV (Church)
--          US01ARZ3NDEKTSV4RRFFQ69G5FAV (User)
-- ULIDs must be generated in application code with appropriate prefix

CREATE TABLE churches (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CH[0-9A-Z]{26}$'),
    name VARCHAR(255) NOT NULL,
    country VARCHAR(100) NOT NULL,
    category VARCHAR(10) NOT NULL CHECK (category IN ('S', 'L', 'XL')),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE users (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^US[0-9A-Z]{26}$'),
    members_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('MALE', 'FEMALE')),
    age INT NOT NULL CHECK (age > 0),
    church_id CHAR(28) NOT NULL REFERENCES churches(id) ON DELETE RESTRICT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_users_church ON users(church_id);
CREATE INDEX idx_users_gender ON users(gender);
CREATE INDEX idx_users_age ON users(age);

CREATE TABLE projects (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^PR[0-9A-Z]{26}$'),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    logo_url VARCHAR(500) NOT NULL,
    color_primary VARCHAR(50) NOT NULL,
    color_secondary VARCHAR(50) NOT NULL,
    color_tertiary VARCHAR(50) NOT NULL,
    archived BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CHECK (end_date > start_date)
);

CREATE TABLE events (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^EV[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CHECK (end_date > start_date)
);

CREATE INDEX idx_events_project ON events(project_id);

CREATE TABLE super_teams (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^ST[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_super_teams_project ON super_teams(project_id);

CREATE TABLE teams (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^TM[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    join_code VARCHAR(50) UNIQUE NOT NULL,
    super_team_id CHAR(28) REFERENCES super_teams(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_teams_project ON teams(project_id);
CREATE INDEX idx_teams_super_team ON teams(super_team_id);
CREATE INDEX idx_teams_join_code ON teams(join_code);

CREATE TABLE streaks (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SK[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_streaks_project ON streaks(project_id);

CREATE TABLE streak_relevant_days (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SD[0-9A-Z]{26}$'),
    streak_id CHAR(28) NOT NULL REFERENCES streaks(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    CHECK (end_date >= start_date)
);

CREATE INDEX idx_streak_relevant_days_streak ON streak_relevant_days(streak_id);

CREATE TABLE challenges (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CL[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    url VARCHAR(500) NOT NULL,
    button_text VARCHAR(100) NOT NULL,
    published_at TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_challenges_project ON challenges(project_id);
CREATE INDEX idx_challenges_event ON challenges(event_id);
CREATE INDEX idx_challenges_published ON challenges(published_at);

CREATE TABLE achievements (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^AC[0-9A-Z]{26}$'),
    achievement_type VARCHAR(50) NOT NULL CHECK (achievement_type IN ('SIMPLE', 'READING', 'LISTENING', 'STREAK')),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL,
    challenge_id CHAR(28) REFERENCES challenges(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    points INT NOT NULL DEFAULT 0,
    hidden BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_achievements_project ON achievements(project_id);
CREATE INDEX idx_achievements_event ON achievements(event_id);
CREATE INDEX idx_achievements_challenge ON achievements(challenge_id);
CREATE INDEX idx_achievements_type ON achievements(achievement_type);

-- Type-specific achievement data
CREATE TABLE reading_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE
);

CREATE TABLE reading_achievement_articles (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^RA[0-9A-Z]{26}$'),
    achievement_id CHAR(28) NOT NULL REFERENCES reading_achievements(achievement_id) ON DELETE CASCADE,
    article_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    author VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    UNIQUE (achievement_id, article_id)
);

CREATE INDEX idx_reading_articles_achievement ON reading_achievement_articles(achievement_id);

CREATE TABLE listening_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE
);

CREATE TABLE listening_achievement_tracks (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^LT[0-9A-Z]{26}$'),
    achievement_id CHAR(28) NOT NULL REFERENCES listening_achievements(achievement_id) ON DELETE CASCADE,
    track_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    image_url VARCHAR(500),
    UNIQUE (achievement_id, track_id)
);

CREATE INDEX idx_listening_tracks_achievement ON listening_achievement_tracks(achievement_id);

CREATE TABLE streak_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE,
    streak_id CHAR(28) NOT NULL REFERENCES streaks(id) ON DELETE CASCADE,
    needed_streak INT NOT NULL CHECK (needed_streak > 0)
);

-- ==================== Junction Tables ====================

CREATE TABLE user_projects (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, project_id)
);

CREATE INDEX idx_user_projects_user ON user_projects(user_id);
CREATE INDEX idx_user_projects_project ON user_projects(project_id);

CREATE TABLE user_events (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id CHAR(28) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, event_id)
);

CREATE INDEX idx_user_events_user ON user_events(user_id);
CREATE INDEX idx_user_events_event ON user_events(event_id);

CREATE TABLE team_members (
    team_id CHAR(28) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_team ON team_members(team_id);
CREATE INDEX idx_team_members_user ON team_members(user_id);

-- ==================== User Progress Tracking ====================

CREATE TABLE user_achievements (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id)
);

CREATE INDEX idx_user_achievements_user ON user_achievements(user_id);
CREATE INDEX idx_user_achievements_achievement ON user_achievements(achievement_id);
CREATE INDEX idx_user_achievements_time ON user_achievements(achieved_at);

CREATE TABLE team_achievements (
    team_id CHAR(28) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (team_id, achievement_id)
);

CREATE INDEX idx_team_achievements_team ON team_achievements(team_id);
CREATE INDEX idx_team_achievements_achievement ON team_achievements(achievement_id);

CREATE TABLE super_team_achievements (
    super_team_id CHAR(28) NOT NULL REFERENCES super_teams(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (super_team_id, achievement_id)
);

CREATE INDEX idx_super_team_achievements_super_team ON super_team_achievements(super_team_id);
CREATE INDEX idx_super_team_achievements_achievement ON super_team_achievements(achievement_id);

CREATE TABLE user_challenge_completions (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id CHAR(28) NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, challenge_id)
);

CREATE INDEX idx_user_challenges_user ON user_challenge_completions(user_id);
CREATE INDEX idx_user_challenges_challenge ON user_challenge_completions(challenge_id);
CREATE INDEX idx_user_challenges_time ON user_challenge_completions(completed_at);

CREATE TABLE user_reading_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES reading_achievements(achievement_id) ON DELETE CASCADE,
    article_id VARCHAR(255) NOT NULL,
    read_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, article_id)
);

CREATE INDEX idx_user_reading_user ON user_reading_progress(user_id);
CREATE INDEX idx_user_reading_achievement ON user_reading_progress(achievement_id);

CREATE TABLE user_listening_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES listening_achievements(achievement_id) ON DELETE CASCADE,
    track_id VARCHAR(255) NOT NULL,
    listened_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, track_id)
);

CREATE INDEX idx_user_listening_user ON user_listening_progress(user_id);
CREATE INDEX idx_user_listening_achievement ON user_listening_progress(achievement_id);

CREATE TABLE user_streak_activity (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    streak_id CHAR(28) NOT NULL REFERENCES streaks(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, streak_id, activity_date)
);

CREATE INDEX idx_streak_activity_user ON user_streak_activity(user_id);
CREATE INDEX idx_streak_activity_streak ON user_streak_activity(streak_id);
CREATE INDEX idx_streak_activity_date ON user_streak_activity(activity_date);

-- ==================== Audit/Activity Log ====================

CREATE TABLE score_adjustments (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SA[0-9A-Z]{26}$'),
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('USER', 'TEAM', 'SUPER_TEAM')),
    entity_id CHAR(28) NOT NULL,
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    points INT NOT NULL,
    reason TEXT,
    adjusted_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_score_adjustments_entity ON score_adjustments(entity_type, entity_id);
CREATE INDEX idx_score_adjustments_project ON score_adjustments(project_id);
CREATE INDEX idx_score_adjustments_time ON score_adjustments(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS score_adjustments;
DROP TABLE IF EXISTS user_streak_activity;
DROP TABLE IF EXISTS user_listening_progress;
DROP TABLE IF EXISTS user_reading_progress;
DROP TABLE IF EXISTS user_challenge_completions;
DROP TABLE IF EXISTS super_team_achievements;
DROP TABLE IF EXISTS team_achievements;
DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS user_events;
DROP TABLE IF EXISTS user_projects;
DROP TABLE IF EXISTS streak_achievements;
DROP TABLE IF EXISTS listening_achievement_tracks;
DROP TABLE IF EXISTS listening_achievements;
DROP TABLE IF EXISTS reading_achievement_articles;
DROP TABLE IF EXISTS reading_achievements;
DROP TABLE IF EXISTS achievements;
DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS streak_relevant_days;
DROP TABLE IF EXISTS streaks;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS super_teams;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS churches;

-- +goose StatementEnd
