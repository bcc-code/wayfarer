-- ==================== Core Tables ====================
-- All IDs use ULID format with 2-character table prefix
-- Format: XX + 26-character ULID = 28 characters total
-- Example: CH01ARZ3NDEKTSV4RRFFQ69G5FAV (Church)
--          US01ARZ3NDEKTSV4RRFFQ69G5FAV (User)
-- ULIDs must be generated in application code with appropriate prefix

CREATE TABLE churches (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CH[0-9A-Z]{26}$'),
    external_id INT UNIQUE,
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
    birthdate DATE,
    church_id CHAR(28) NOT NULL REFERENCES churches(id) ON DELETE RESTRICT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_users_church (church_id),
    INDEX idx_users_gender (gender),
    INDEX idx_users_birthdate (birthdate)
);

-- ==================== Authorization Tables ====================

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

    -- Indexes
    INDEX idx_user_roles_user (user_id),
    INDEX idx_user_roles_role (role),
    INDEX idx_user_roles_church (church_id),
    INDEX idx_user_roles_project (project_id),
    INDEX idx_user_roles_team (team_id),

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

CREATE TABLE projects (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^PR[0-9A-Z]{26}$'),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    logo_url VARCHAR(500),
    color_primary VARCHAR(50) NOT NULL,
    color_secondary VARCHAR(50) NOT NULL,
    color_tertiary VARCHAR(50) NOT NULL,
    rounding INT NOT NULL DEFAULT 0,
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
    INDEX idx_events_project (project_id),
    CHECK (end_date > start_date)
);

CREATE TABLE super_teams (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^ST[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_super_teams_project (project_id)
);

CREATE TABLE teams (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^TM[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    join_code VARCHAR(50) UNIQUE NOT NULL,
    super_team_id CHAR(28) REFERENCES super_teams(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_teams_project (project_id),
    INDEX idx_teams_super_team (super_team_id),
    INDEX idx_teams_join_code (join_code)
);

CREATE TABLE streaks (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SK[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_streaks_project (project_id)
);

CREATE TABLE streak_relevant_days (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SD[0-9A-Z]{26}$'),
    streak_id CHAR(28) NOT NULL REFERENCES streaks(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    INDEX idx_streak_relevant_days_streak (streak_id),
    CHECK (end_date >= start_date)
);

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
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_challenges_project (project_id),
    INDEX idx_challenges_event (event_id),
    INDEX idx_challenges_published (published_at)
);

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
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_achievements_project (project_id),
    INDEX idx_achievements_event (event_id),
    INDEX idx_achievements_challenge (challenge_id),
    INDEX idx_achievements_type (achievement_type)
);

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
    INDEX idx_reading_articles_achievement (achievement_id),
    UNIQUE (achievement_id, article_id)
);

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
    INDEX idx_listening_tracks_achievement (achievement_id),
    UNIQUE (achievement_id, track_id)
);

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
    PRIMARY KEY (user_id, project_id),
    INDEX idx_user_projects_user (user_id),
    INDEX idx_user_projects_project (project_id)
);

CREATE TABLE user_events (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id CHAR(28) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, event_id),
    INDEX idx_user_events_user (user_id),
    INDEX idx_user_events_event (event_id)
);

CREATE TABLE team_members (
    team_id CHAR(28) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (team_id, user_id),
    INDEX idx_team_members_team (team_id),
    INDEX idx_team_members_user (user_id)
);

-- ==================== User Progress Tracking ====================

CREATE TABLE user_achievements (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id),
    INDEX idx_user_achievements_user (user_id),
    INDEX idx_user_achievements_achievement (achievement_id),
    INDEX idx_user_achievements_time (achieved_at)
);

CREATE TABLE team_achievements (
    team_id CHAR(28) NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (team_id, achievement_id),
    INDEX idx_team_achievements_team (team_id),
    INDEX idx_team_achievements_achievement (achievement_id)
);

CREATE TABLE super_team_achievements (
    super_team_id CHAR(28) NOT NULL REFERENCES super_teams(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    achieved_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (super_team_id, achievement_id),
    INDEX idx_super_team_achievements_super_team (super_team_id),
    INDEX idx_super_team_achievements_achievement (achievement_id)
);

CREATE TABLE user_challenge_completions (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id CHAR(28) NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, challenge_id),
    INDEX idx_user_challenges_user (user_id),
    INDEX idx_user_challenges_challenge (challenge_id),
    INDEX idx_user_challenges_time (completed_at)
);

CREATE TABLE user_reading_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES reading_achievements(achievement_id) ON DELETE CASCADE,
    article_id VARCHAR(255) NOT NULL,
    read_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, article_id),
    INDEX idx_user_reading_user (user_id),
    INDEX idx_user_reading_achievement (achievement_id)
);

CREATE TABLE user_listening_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES listening_achievements(achievement_id) ON DELETE CASCADE,
    track_id VARCHAR(255) NOT NULL,
    listened_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, track_id),
    INDEX idx_user_listening_user (user_id),
    INDEX idx_user_listening_achievement (achievement_id)
);

CREATE TABLE user_streak_activity (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    streak_id CHAR(28) NOT NULL REFERENCES streaks(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, streak_id, activity_date),
    INDEX idx_streak_activity_user (user_id),
    INDEX idx_streak_activity_streak (streak_id),
    INDEX idx_streak_activity_date (activity_date)
);

-- ==================== Audit/Activity Log ====================

CREATE TABLE score_adjustments (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SA[0-9A-Z]{26}$'),
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('USER', 'TEAM', 'SUPER_TEAM')),
    entity_id CHAR(28) NOT NULL,
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    points INT NOT NULL,
    reason TEXT,
    adjusted_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_score_adjustments_entity (entity_type, entity_id),
    INDEX idx_score_adjustments_project (project_id),
    INDEX idx_score_adjustments_time (created_at)
);
