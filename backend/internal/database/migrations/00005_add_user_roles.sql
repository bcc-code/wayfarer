-- +goose Up
-- +goose StatementBegin

-- Create user_roles table for permission management
CREATE TABLE user_roles (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^UR[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('SUPERADMIN', 'ADMIN', 'CHURCH_ADMIN', 'PROJECT_ADMIN', 'TEAM_LEAD', 'USER', 'M2M')),

    -- Scope columns (only one should be non-null for scoped roles)
    church_id CHAR(28) REFERENCES churches(id) ON DELETE CASCADE,
    project_id CHAR(28) REFERENCES projects(id) ON DELETE CASCADE,
    team_id CHAR(28) REFERENCES teams(id) ON DELETE CASCADE,

    assigned_by CHAR(28) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assigned_at TIMESTAMPTZ DEFAULT now(),

    -- Constraints to enforce proper scoping
    CHECK (
        -- Global roles must have no scope
        (role IN ('SUPERADMIN', 'ADMIN', 'USER', 'M2M') AND church_id IS NULL AND project_id IS NULL AND team_id IS NULL)
        OR
        -- Church admin must have exactly one church_id
        (role = 'CHURCH_ADMIN' AND church_id IS NOT NULL AND project_id IS NULL AND team_id IS NULL)
        OR
        -- Project admin must have exactly one project_id
        (role = 'PROJECT_ADMIN' AND church_id IS NULL AND project_id IS NOT NULL AND team_id IS NULL)
        OR
        -- Team lead must have exactly one team_id
        (role = 'TEAM_LEAD' AND church_id IS NULL AND project_id IS NULL AND team_id IS NOT NULL)
    ),

    -- Prevent duplicate role assignments
    UNIQUE (user_id, role, church_id, project_id, team_id)
);

-- Create indexes for efficient querying
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role);
CREATE INDEX idx_user_roles_church ON user_roles(church_id);
CREATE INDEX idx_user_roles_project ON user_roles(project_id);
CREATE INDEX idx_user_roles_team ON user_roles(team_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_user_roles_team;
DROP INDEX IF EXISTS idx_user_roles_project;
DROP INDEX IF EXISTS idx_user_roles_church;
DROP INDEX IF EXISTS idx_user_roles_role;
DROP INDEX IF EXISTS idx_user_roles_user;

-- Drop user_roles table
DROP TABLE IF EXISTS user_roles;

-- +goose StatementEnd
