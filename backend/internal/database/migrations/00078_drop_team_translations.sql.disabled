-- +goose Up
DROP TABLE IF EXISTS team_translations;
DROP TABLE IF EXISTS super_team_translations;

-- +goose Down
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
