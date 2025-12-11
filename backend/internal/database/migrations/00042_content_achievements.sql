-- +goose Up
-- +goose StatementBegin

-- Migration: Unify Reading and Listening achievements into Content achievements
-- This is a breaking change - removes READING and LISTENING types, replaces with CONTENT

-- Step 1: Add CONTENT to achievement_type enum (temporarily keep READING/LISTENING)
ALTER TABLE achievements
    DROP CONSTRAINT achievements_achievement_type_check;

ALTER TABLE achievements
    ADD CONSTRAINT achievements_achievement_type_check
    CHECK (achievement_type IN ('SIMPLE', 'READING', 'LISTENING', 'CONTENT', 'STREAK', 'QUIZ'));

-- Step 2: Create content_achievements junction table
CREATE TABLE content_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE
);

-- Step 3: Create content_achievement_items table (replaces reading_achievement_articles and listening_achievement_tracks)
CREATE TABLE content_achievement_items (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CI[0-9A-Z]{26}$'),
    achievement_id CHAR(28) NOT NULL REFERENCES content_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) NOT NULL REFERENCES external_content(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    UNIQUE (achievement_id, external_content_id)
);

CREATE INDEX idx_content_items_achievement ON content_achievement_items(achievement_id);
CREATE INDEX idx_content_items_external_content ON content_achievement_items(external_content_id);

-- Step 4: Create unified user progress table (replaces user_reading_progress and user_listening_progress)
CREATE TABLE user_content_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES content_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) NOT NULL REFERENCES external_content(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, external_content_id)
);

CREATE INDEX idx_user_content_progress_user ON user_content_progress(user_id);
CREATE INDEX idx_user_content_progress_achievement ON user_content_progress(achievement_id);
CREATE INDEX idx_user_content_progress_content ON user_content_progress(external_content_id);

-- Step 5: Migrate existing data from reading achievements
INSERT INTO content_achievements (achievement_id)
SELECT achievement_id FROM reading_achievements;

INSERT INTO content_achievement_items (id, achievement_id, external_content_id, sort_order)
SELECT
    'CI' || substring(id from 3),
    achievement_id,
    external_content_id,
    ROW_NUMBER() OVER (PARTITION BY achievement_id ORDER BY id) - 1
FROM reading_achievement_articles
WHERE external_content_id IS NOT NULL;

-- Note: user_reading_progress uses article_id which is a VARCHAR, not a reference to reading_achievement_articles.id
-- We need to migrate by matching article_id to the reading_achievement_articles table
INSERT INTO user_content_progress (user_id, achievement_id, external_content_id, completed_at)
SELECT DISTINCT urp.user_id, urp.achievement_id, raa.external_content_id, urp.read_at
FROM user_reading_progress urp
JOIN reading_achievement_articles raa ON raa.achievement_id = urp.achievement_id
WHERE raa.external_content_id IS NOT NULL;

-- Step 6: Migrate existing data from listening achievements
INSERT INTO content_achievements (achievement_id)
SELECT achievement_id FROM listening_achievements
ON CONFLICT (achievement_id) DO NOTHING;

INSERT INTO content_achievement_items (id, achievement_id, external_content_id, sort_order)
SELECT
    'CI' || substring(id from 3),
    achievement_id,
    external_content_id,
    ROW_NUMBER() OVER (PARTITION BY achievement_id ORDER BY id) - 1
FROM listening_achievement_tracks
WHERE external_content_id IS NOT NULL;

INSERT INTO user_content_progress (user_id, achievement_id, external_content_id, completed_at)
SELECT DISTINCT ulp.user_id, ulp.achievement_id, lat.external_content_id, ulp.listened_at
FROM user_listening_progress ulp
JOIN listening_achievement_tracks lat ON lat.achievement_id = ulp.achievement_id
WHERE lat.external_content_id IS NOT NULL
ON CONFLICT (user_id, achievement_id, external_content_id) DO NOTHING;

-- Step 7: Update achievement types from READING/LISTENING to CONTENT
UPDATE achievements SET achievement_type = 'CONTENT' WHERE achievement_type IN ('READING', 'LISTENING');

-- Step 8: Drop old translation tables first (they reference the content tables)
DROP TABLE IF EXISTS reading_achievement_article_translations;
DROP TABLE IF EXISTS listening_achievement_track_translations;

-- Step 9: Drop old progress tables
DROP TABLE IF EXISTS user_reading_progress;
DROP TABLE IF EXISTS user_listening_progress;

-- Step 10: Drop old content tables
DROP TABLE IF EXISTS reading_achievement_articles;
DROP TABLE IF EXISTS listening_achievement_tracks;

-- Step 11: Drop old junction tables
DROP TABLE IF EXISTS reading_achievements;
DROP TABLE IF EXISTS listening_achievements;

-- Step 12: Remove READING and LISTENING from enum
ALTER TABLE achievements
    DROP CONSTRAINT achievements_achievement_type_check;

ALTER TABLE achievements
    ADD CONSTRAINT achievements_achievement_type_check
    CHECK (achievement_type IN ('SIMPLE', 'CONTENT', 'STREAK', 'QUIZ'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverse migration: Restore READING and LISTENING achievement types

-- Step 1: Re-add READING and LISTENING to enum
ALTER TABLE achievements
    DROP CONSTRAINT achievements_achievement_type_check;

ALTER TABLE achievements
    ADD CONSTRAINT achievements_achievement_type_check
    CHECK (achievement_type IN ('SIMPLE', 'READING', 'LISTENING', 'CONTENT', 'STREAK', 'QUIZ'));

-- Step 2: Recreate reading_achievements table
CREATE TABLE reading_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE
);

-- Step 3: Recreate reading_achievement_articles table
CREATE TABLE reading_achievement_articles (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^RA[0-9A-Z]{26}$'),
    achievement_id CHAR(28) NOT NULL REFERENCES reading_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) REFERENCES external_content(id) ON DELETE CASCADE,
    UNIQUE (achievement_id, external_content_id)
);

CREATE INDEX idx_reading_articles_achievement ON reading_achievement_articles(achievement_id);
CREATE INDEX idx_reading_articles_external_content ON reading_achievement_articles(external_content_id);

-- Step 4: Recreate listening_achievements table
CREATE TABLE listening_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE
);

-- Step 5: Recreate listening_achievement_tracks table
CREATE TABLE listening_achievement_tracks (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^LT[0-9A-Z]{26}$'),
    achievement_id CHAR(28) NOT NULL REFERENCES listening_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) REFERENCES external_content(id) ON DELETE CASCADE,
    UNIQUE (achievement_id, external_content_id)
);

CREATE INDEX idx_listening_tracks_achievement ON listening_achievement_tracks(achievement_id);
CREATE INDEX idx_listening_tracks_external_content ON listening_achievement_tracks(external_content_id);

-- Step 6: Recreate user_reading_progress table
CREATE TABLE user_reading_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES reading_achievements(achievement_id) ON DELETE CASCADE,
    article_id VARCHAR(255) NOT NULL,
    read_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, article_id)
);

CREATE INDEX idx_user_reading_user ON user_reading_progress(user_id);
CREATE INDEX idx_user_reading_achievement ON user_reading_progress(achievement_id);

-- Step 7: Recreate user_listening_progress table
CREATE TABLE user_listening_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES listening_achievements(achievement_id) ON DELETE CASCADE,
    track_id VARCHAR(255) NOT NULL,
    listened_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, track_id)
);

CREATE INDEX idx_user_listening_user ON user_listening_progress(user_id);
CREATE INDEX idx_user_listening_achievement ON user_listening_progress(achievement_id);

-- Note: Data migration back is lossy - we cannot distinguish between reading and listening achievements
-- Admins would need to manually re-categorize CONTENT achievements back to READING or LISTENING

-- Step 8: Drop new tables
DROP TABLE IF EXISTS user_content_progress;
DROP TABLE IF EXISTS content_achievement_items;
DROP TABLE IF EXISTS content_achievements;

-- Step 9: Remove CONTENT from enum
ALTER TABLE achievements
    DROP CONSTRAINT achievements_achievement_type_check;

ALTER TABLE achievements
    ADD CONSTRAINT achievements_achievement_type_check
    CHECK (achievement_type IN ('SIMPLE', 'READING', 'LISTENING', 'STREAK', 'QUIZ'));

-- +goose StatementEnd
