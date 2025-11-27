-- +goose Up
-- +goose StatementBegin

-- Translation tables for i18n support
-- Each entity with translatable fields gets a shadow table
-- Composite primary key (entity_id, language_code) for efficient lookups
-- English content remains in main tables; translations for other languages here

CREATE TABLE project_translations (
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (project_id, language_code)
);

CREATE TABLE event_translations (
    event_id CHAR(28) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (event_id, language_code)
);

CREATE TABLE team_translations (
    team_id CHAR(28) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (team_id, language_code)
);

CREATE TABLE super_team_translations (
    super_team_id CHAR(28) NOT NULL REFERENCES super_teams(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (super_team_id, language_code)
);

CREATE TABLE streak_translations (
    streak_id CHAR(28) NOT NULL REFERENCES streaks(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (streak_id, language_code)
);

CREATE TABLE challenge_translations (
    challenge_id CHAR(28) NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    button_text VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (challenge_id, language_code)
);

CREATE TABLE achievement_translations (
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (achievement_id, language_code)
);

CREATE TABLE reading_achievement_article_translations (
    article_id CHAR(28) NOT NULL REFERENCES reading_achievement_articles(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    title VARCHAR(500),
    author VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (article_id, language_code)
);

CREATE TABLE listening_achievement_track_translations (
    track_id CHAR(28) NOT NULL REFERENCES listening_achievement_tracks(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(500),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (track_id, language_code)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS listening_achievement_track_translations;
DROP TABLE IF EXISTS reading_achievement_article_translations;
DROP TABLE IF EXISTS achievement_translations;
DROP TABLE IF EXISTS challenge_translations;
DROP TABLE IF EXISTS streak_translations;
DROP TABLE IF EXISTS super_team_translations;
DROP TABLE IF EXISTS team_translations;
DROP TABLE IF EXISTS event_translations;
DROP TABLE IF EXISTS project_translations;

-- +goose StatementEnd
