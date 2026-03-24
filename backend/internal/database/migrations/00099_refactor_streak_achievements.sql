-- +goose Up
-- +goose StatementBegin

-- Migration: Refactor streak achievements from consecutive-day tracking to external content with deadlines
-- Drops: streaks, streak_relevant_days, user_streak_activity
-- Modifies: streak_achievements becomes a marker table (like content_achievements)
-- Creates: streak_achievement_items, user_streak_progress

-- Step 1: Drop user_streak_activity (depends on streaks)
DROP TABLE IF EXISTS user_streak_activity;

-- Step 2: Drop streak_relevant_days (depends on streaks)
DROP TABLE IF EXISTS streak_relevant_days;

-- Step 3: Drop streak_translations (depends on streaks)
DROP TABLE IF EXISTS streak_translations;

-- Step 4: Modify streak_achievements - drop FK to streaks and data columns
ALTER TABLE streak_achievements
    DROP COLUMN streak_id,
    DROP COLUMN needed_streak;

-- Step 4: Drop streaks table (no more dependents)
DROP TABLE IF EXISTS streaks;

-- Step 5: Create streak_achievement_items (mirrors content_achievement_items)
CREATE TABLE streak_achievement_items (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SI[0-9A-Z]{26}$'),
    achievement_id CHAR(28) NOT NULL REFERENCES streak_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) NOT NULL REFERENCES external_content(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    UNIQUE (achievement_id, external_content_id)
);

CREATE INDEX idx_streak_items_achievement ON streak_achievement_items(achievement_id);
CREATE INDEX idx_streak_items_external_content ON streak_achievement_items(external_content_id);

-- Step 6: Create user_streak_progress (mirrors user_content_progress)
CREATE TABLE user_streak_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES streak_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) NOT NULL REFERENCES external_content(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, external_content_id)
);

CREATE INDEX idx_user_streak_progress_user ON user_streak_progress(user_id);
CREATE INDEX idx_user_streak_progress_achievement ON user_streak_progress(achievement_id);
CREATE INDEX idx_user_streak_progress_content ON user_streak_progress(external_content_id);
CREATE INDEX idx_user_streak_progress_user_achievement ON user_streak_progress(user_id, achievement_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverse: Drop new tables
DROP TABLE IF EXISTS user_streak_progress;
DROP TABLE IF EXISTS streak_achievement_items;

-- Reverse: Recreate streaks table
CREATE TABLE streaks (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SK[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_streaks_project ON streaks(project_id);

-- Reverse: Recreate streak_relevant_days
CREATE TABLE streak_relevant_days (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SD[0-9A-Z]{26}$'),
    streak_id CHAR(28) NOT NULL REFERENCES streaks(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    CHECK (end_date >= start_date)
);

CREATE INDEX idx_streak_relevant_days_streak ON streak_relevant_days(streak_id);

-- Reverse: Re-add columns to streak_achievements
ALTER TABLE streak_achievements
    ADD COLUMN streak_id CHAR(28) REFERENCES streaks(id) ON DELETE CASCADE,
    ADD COLUMN needed_streak INT CHECK (needed_streak > 0);

-- Reverse: Recreate user_streak_activity
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

-- +goose StatementEnd
