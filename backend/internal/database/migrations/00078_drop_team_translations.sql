-- +goose Up
DROP TABLE IF EXISTS team_translations;

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
