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
    person_uuid UUID,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('MALE', 'FEMALE', 'UNKNOWN')),
    birthdate DATE,
    church_id CHAR(28) NOT NULL REFERENCES churches(id) ON DELETE RESTRICT,
    church_locked_until TIMESTAMPTZ,
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
    banner_url VARCHAR(500),
    -- Light mode colors
    color_light_accent VARCHAR(50) NOT NULL,
    color_light_accent_contrast VARCHAR(50) NOT NULL,
    color_light_on_accent VARCHAR(50) NOT NULL,
    color_light_background_default VARCHAR(50) NOT NULL,
    color_light_background_raised VARCHAR(50) NOT NULL,
    color_light_background_indent VARCHAR(50) NOT NULL,
    color_light_text_default VARCHAR(50) NOT NULL,
    color_light_text_muted VARCHAR(50) NOT NULL,
    color_light_text_hint VARCHAR(50) NOT NULL,
    color_light_shadow_default VARCHAR(50) NOT NULL,
    color_light_shadow_blank VARCHAR(50) NOT NULL,
    color_light_border_default VARCHAR(50) NOT NULL,
    -- Dark mode colors
    color_dark_accent VARCHAR(50) NOT NULL,
    color_dark_accent_contrast VARCHAR(50) NOT NULL,
    color_dark_on_accent VARCHAR(50) NOT NULL,
    color_dark_background_default VARCHAR(50) NOT NULL,
    color_dark_background_raised VARCHAR(50) NOT NULL,
    color_dark_background_indent VARCHAR(50) NOT NULL,
    color_dark_text_default VARCHAR(50) NOT NULL,
    color_dark_text_muted VARCHAR(50) NOT NULL,
    color_dark_text_hint VARCHAR(50) NOT NULL,
    color_dark_shadow_default VARCHAR(50) NOT NULL,
    color_dark_shadow_blank VARCHAR(50) NOT NULL,
    color_dark_border_default VARCHAR(50) NOT NULL,
    rounding INT NOT NULL DEFAULT 0,
    archived BOOLEAN DEFAULT false,
    rules TEXT,
    info_message TEXT,
    info_message_start TIMESTAMPTZ,
    info_message_end TIMESTAMPTZ,
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
    image_url VARCHAR(500),
    color VARCHAR(7),
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
    leaderboard_excluded BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_teams_project (project_id),
    INDEX idx_teams_super_team (super_team_id),
    INDEX idx_teams_join_code (join_code),
    INDEX idx_teams_leaderboard_excluded (leaderboard_excluded) WHERE leaderboard_excluded = true
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
    challenge_type VARCHAR(50) NOT NULL DEFAULT 'SIMPLE' CHECK (challenge_type IN ('SIMPLE', 'QUIZ', 'EXTERNAL', 'PLUGIN')),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    image_url VARCHAR(500),
    url VARCHAR(500),
    button_text VARCHAR(100) NOT NULL,
    published_at TIMESTAMPTZ,
    visible_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    allow_self_completion BOOLEAN DEFAULT true NOT NULL,
    requires_team_membership BOOLEAN DEFAULT false NOT NULL,
    requires_super_team_membership BOOLEAN DEFAULT false NOT NULL,
    plugin_challenge_id VARCHAR(100),
    plugin_data JSONB,
    notification_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    -- Type constraints: EXTERNAL requires url, QUIZ must not have url, PLUGIN requires plugin_challenge_id
    CONSTRAINT challenge_type_url_constraint CHECK (
        (challenge_type = 'EXTERNAL' AND url IS NOT NULL AND url != '') OR
        (challenge_type = 'QUIZ' AND (url IS NULL OR url = '')) OR
        (challenge_type = 'SIMPLE') OR
        (challenge_type = 'PLUGIN' AND plugin_challenge_id IS NOT NULL AND plugin_challenge_id != '')
    ),
    INDEX idx_challenges_project (project_id),
    INDEX idx_challenges_event (event_id),
    INDEX idx_challenges_published (published_at),
    INDEX idx_challenges_visible (visible_at),
    INDEX idx_challenges_type (challenge_type),
    UNIQUE INDEX idx_challenges_plugin_challenge_id (plugin_challenge_id) WHERE plugin_challenge_id IS NOT NULL
);

CREATE TABLE achievements (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^AC[0-9A-Z]{26}$'),
    achievement_type VARCHAR(50) NOT NULL CHECK (achievement_type IN ('SIMPLE', 'CONTENT', 'STREAK', 'QUIZ')),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL,
    challenge_id CHAR(28) REFERENCES challenges(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description_pending TEXT NOT NULL,
    description_completed TEXT NOT NULL,
    notification_text TEXT NOT NULL DEFAULT '',
    image_pending VARCHAR(500) NOT NULL,
    image_completed VARCHAR(500) NOT NULL,
    points INT NOT NULL DEFAULT 0,
    hidden BOOLEAN DEFAULT false,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_achievements_project (project_id),
    INDEX idx_achievements_event (event_id),
    INDEX idx_achievements_challenge (challenge_id),
    INDEX idx_achievements_type (achievement_type)
);

-- Type-specific achievement data
CREATE TABLE content_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE
);

CREATE TABLE content_achievement_items (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CI[0-9A-Z]{26}$'),
    achievement_id CHAR(28) NOT NULL REFERENCES content_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) NOT NULL REFERENCES external_content(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    INDEX idx_content_items_achievement (achievement_id),
    INDEX idx_content_items_external_content (external_content_id),
    UNIQUE (achievement_id, external_content_id)
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
    celebrated_at TIMESTAMPTZ,
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

CREATE TABLE user_content_progress (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id CHAR(28) NOT NULL REFERENCES content_achievements(achievement_id) ON DELETE CASCADE,
    external_content_id CHAR(28) NOT NULL REFERENCES external_content(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id, external_content_id),
    INDEX idx_user_content_progress_user (user_id),
    INDEX idx_user_content_progress_achievement (achievement_id),
    INDEX idx_user_content_progress_content (external_content_id),
    INDEX idx_user_content_progress_user_achievement (user_id, achievement_id)
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

-- ==================== Translation Tables ====================
-- Shadow tables for i18n support
-- English content remains in main tables; translations for other languages here
-- Composite primary key (entity_id, language_code) for efficient lookups

CREATE TABLE project_translations (
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    rules TEXT,
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
    notification_text TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (challenge_id, language_code)
);

CREATE TABLE achievement_translations (
    achievement_id CHAR(28) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description_pending TEXT,
    description_completed TEXT,
    notification_text TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (achievement_id, language_code)
);


-- ==================== Consent Tables ====================
-- Global user consent management (system-wide, not project-scoped)
-- Consents are versioned to support re-acceptance on updates

CREATE TABLE consents (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CN[0-9A-Z]{26}$'),
    key VARCHAR(100) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    title VARCHAR(255) NOT NULL,
    short_text TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    url VARCHAR(500),
    published_at TIMESTAMPTZ,
    managed_by VARCHAR(100),
    is_remote BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (key, version),
    INDEX idx_consents_key (key),
    INDEX idx_consents_published (published_at) WHERE published_at IS NOT NULL,
    INDEX idx_consents_is_remote (is_remote) WHERE is_remote = true
);

CREATE TABLE consent_translations (
    consent_id CHAR(28) NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    title VARCHAR(255),
    short_text TEXT,
    body TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (consent_id, language_code)
);

CREATE TABLE user_consent_history (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^UH[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_id CHAR(28) NOT NULL REFERENCES consents(id) ON DELETE RESTRICT,
    consent_key VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('ACCEPTED', 'REJECTED')),
    occurred_at TIMESTAMPTZ NOT NULL,
    source VARCHAR(100),
    external_consent_id VARCHAR(255),
    external_timestamp TIMESTAMPTZ,
    INDEX idx_user_consent_history_user (user_id),
    INDEX idx_user_consent_history_consent (consent_id),
    INDEX idx_user_consent_history_key (consent_key),
    INDEX idx_user_consent_history_user_key (user_id, consent_key),
    INDEX idx_user_consent_history_occurred (occurred_at),
    UNIQUE INDEX idx_user_consent_history_remote_idempotent (user_id, consent_key, external_consent_id) WHERE external_consent_id IS NOT NULL
);

-- ==================== Quiz System Tables ====================

CREATE TABLE quizzes (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QZ[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    challenge_id CHAR(28) NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,

    -- Basic info
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    image_url VARCHAR(500),

    -- Timing configuration (one of these must be set, not both)
    timeout_seconds INT,              -- Quiz-level timeout (e.g., 1800 for 30 min)
    question_timeout_seconds INT,     -- Per-question timeout (e.g., 30 sec)

    -- Configuration
    randomize_questions BOOLEAN DEFAULT false NOT NULL,
    reveal_correct_answers BOOLEAN DEFAULT true NOT NULL,
    allow_retakes BOOLEAN DEFAULT false NOT NULL,

    -- Points awarded on completion (independent of correctness)
    completion_points INT DEFAULT 0 NOT NULL,

    -- Publishing
    published_at TIMESTAMPTZ,
    end_time TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    INDEX idx_quizzes_project (project_id),
    INDEX idx_quizzes_challenge (challenge_id),
    INDEX idx_quizzes_published (published_at),
    CHECK (timeout_seconds IS NULL OR question_timeout_seconds IS NULL)
);

CREATE TABLE quiz_questions (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QQ[0-9A-Z]{26}$'),
    quiz_id CHAR(28) NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,

    -- Question details
    question_type VARCHAR(50) NOT NULL CHECK (question_type IN ('PREDEFINED', 'FREE_TEXT', 'NUMBER', 'JSON')),
    question_text TEXT NOT NULL,
    question_order INT NOT NULL,

    -- For predefined questions
    allow_multiple_selection BOOLEAN DEFAULT false,

    -- For number questions
    min_value DECIMAL,
    max_value DECIMAL,
    step_value DECIMAL,

    -- Points for this question (null = use quiz default)
    points INT,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    INDEX idx_quiz_questions_quiz (quiz_id),
    UNIQUE (quiz_id, question_order)
);

CREATE TABLE quiz_predefined_answers (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QA[0-9A-Z]{26}$'),
    question_id CHAR(28) NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,

    answer_text TEXT NOT NULL,
    is_correct BOOLEAN DEFAULT false NOT NULL,
    answer_order INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now(),

    INDEX idx_quiz_answers_question (question_id),
    UNIQUE (question_id, answer_order)
);

CREATE TABLE quiz_submissions (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QS[0-9A-Z]{26}$'),
    quiz_id CHAR(28) NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Session tracking
    started_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,

    -- Question order for this specific submission (JSON array of question IDs)
    -- Supports randomization per user
    question_order JSONB NOT NULL,

    -- Scoring (calculated from correct answers)
    score INT,           -- Number of correct answers
    max_score INT,       -- Total number of gradable questions

    -- Points awarded (copied from quiz.completion_points when completed)
    points_awarded INT,

    created_at TIMESTAMPTZ DEFAULT now(),

    INDEX idx_quiz_submissions_quiz (quiz_id),
    INDEX idx_quiz_submissions_user (user_id),
    INDEX idx_quiz_submissions_completed (completed_at),
    INDEX idx_quiz_submissions_expires (expires_at)
);

CREATE TABLE quiz_responses (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QR[0-9A-Z]{26}$'),
    submission_id CHAR(28) NOT NULL REFERENCES quiz_submissions(id) ON DELETE CASCADE,
    question_id CHAR(28) NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,

    -- Response data (polymorphic based on question_type)
    -- For PREDEFINED: array of selected answer IDs
    selected_answer_ids JSONB,

    -- For FREE_TEXT: text response
    text_response TEXT,

    -- For NUMBER: numeric value
    number_response DECIMAL,

    -- For JSON: structured data
    json_response JSONB,

    -- Correctness (null for FREE_TEXT and JSON types)
    is_correct BOOLEAN,

    -- Points earned for this response (null if not correct or not gradable)
    points_earned INT,

    -- Timing
    answered_at TIMESTAMPTZ DEFAULT now(),
    time_spent_seconds INT,

    INDEX idx_quiz_responses_submission (submission_id),
    INDEX idx_quiz_responses_question (question_id),
    UNIQUE (submission_id, question_id)
);

CREATE TABLE quiz_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE,
    quiz_id CHAR(28) NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,

    -- Requirements for earning the achievement
    min_score_percentage INT,     -- e.g., 80 means need 80% correct to earn
    require_completion BOOLEAN DEFAULT true NOT NULL,

    CHECK (min_score_percentage >= 0 AND min_score_percentage <= 100)
);

CREATE TABLE quiz_translations (
    quiz_id CHAR(28) NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (quiz_id, language_code)
);

CREATE TABLE quiz_question_translations (
    question_id CHAR(28) NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    question_text TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (question_id, language_code)
);

CREATE TABLE quiz_answer_translations (
    answer_id CHAR(28) NOT NULL REFERENCES quiz_predefined_answers(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    answer_text TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (answer_id, language_code)
);

-- Push Notification Tables

CREATE TABLE push_subscriptions (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^PS[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh_key TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    INDEX idx_push_subscriptions_user_id (user_id)
);

CREATE TABLE push_notification_preferences (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, notification_type)
);

CREATE TABLE push_notification_log (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^PN[0-9A-Z]{26}$'),
    notification_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    url TEXT,
    data JSONB,
    target_criteria JSONB NOT NULL,
    sent_by CHAR(28) REFERENCES users(id) ON DELETE SET NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    total_recipients INT NOT NULL DEFAULT 0,
    successful_deliveries INT NOT NULL DEFAULT 0,
    failed_deliveries INT NOT NULL DEFAULT 0,
    INDEX idx_push_notification_log_sent_at (sent_at)
);

-- ==================== User Feedback ====================

CREATE TABLE user_feedback (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^FB[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    can_contact_me BOOLEAN NOT NULL DEFAULT FALSE,
    user_agent TEXT,
    platform TEXT,
    screen_width INT,
    screen_height INT,
    app_version TEXT,
    locale TEXT,
    project_id CHAR(28) REFERENCES projects(id) ON DELETE SET NULL,
    timezone TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    INDEX idx_user_feedback_user_id (user_id),
    INDEX idx_user_feedback_created_at (created_at DESC)
);

-- ==================== Webhooks ====================

CREATE TABLE webhooks (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^WH[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(2000) NOT NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('external_content_event', 'points_awarded')),
    include_user_data BOOLEAN DEFAULT true NOT NULL,
    include_event_data BOOLEAN DEFAULT true NOT NULL,
    active BOOLEAN DEFAULT true NOT NULL,
    secret VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_webhooks_project (project_id),
    INDEX idx_webhooks_event_active (project_id, event_type) WHERE active = true
);

CREATE TABLE webhook_logs (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^WL[0-9A-Z]{26}$'),
    webhook_id CHAR(28) NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    request_payload JSONB NOT NULL,
    response_status_code INT,
    response_body TEXT,
    duration_ms INT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    INDEX idx_webhook_logs_webhook (webhook_id),
    INDEX idx_webhook_logs_created (created_at DESC)
);
