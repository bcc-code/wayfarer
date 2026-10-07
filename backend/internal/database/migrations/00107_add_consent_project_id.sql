-- +goose Up
-- +goose StatementBegin

-- Consents become project-scoped. NULL means the consent applies to every
-- project, which is how every existing consent behaves today, so all existing
-- rows keep working unchanged.
--
-- ON DELETE RESTRICT, unlike the CASCADE used by events/teams/super_teams:
-- consents are referenced by user_consent_history with ON DELETE RESTRICT, so a
-- cascade from projects could only ever fail at runtime once a single user has
-- accepted the consent. Projects are archived, not deleted.
ALTER TABLE consents
    ADD COLUMN project_id CHAR(28) REFERENCES projects(id) ON DELETE RESTRICT;

CREATE INDEX idx_consents_project ON consents(project_id) WHERE project_id IS NOT NULL;

COMMENT ON COLUMN consents.project_id IS 'Project this consent applies to; NULL means it applies to every project';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_consents_project;
ALTER TABLE consents DROP COLUMN project_id;

-- +goose StatementEnd
