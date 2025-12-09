-- +goose Up
-- +goose StatementBegin

-- Migration: Add quiz system tables
-- This migration adds support for quizzes with multiple question types,
-- configurable timing, answer correctness tracking, and achievement integration.

-- ==================== Core Quiz Tables ====================

CREATE TABLE quizzes (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QZ[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL,

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

    CHECK (timeout_seconds IS NULL OR question_timeout_seconds IS NULL)
);

CREATE INDEX idx_quizzes_project ON quizzes(project_id);
CREATE INDEX idx_quizzes_event ON quizzes(event_id);
CREATE INDEX idx_quizzes_published ON quizzes(published_at);

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

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE (quiz_id, question_order)
);

CREATE INDEX idx_quiz_questions_quiz ON quiz_questions(quiz_id);

CREATE TABLE quiz_predefined_answers (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QA[0-9A-Z]{26}$'),
    question_id CHAR(28) NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,

    answer_text TEXT NOT NULL,
    is_correct BOOLEAN DEFAULT false NOT NULL,
    answer_order INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE (question_id, answer_order)
);

CREATE INDEX idx_quiz_answers_question ON quiz_predefined_answers(question_id);

-- ==================== User Submission Tables ====================

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

    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_quiz_submissions_quiz ON quiz_submissions(quiz_id);
CREATE INDEX idx_quiz_submissions_user ON quiz_submissions(user_id);
CREATE INDEX idx_quiz_submissions_completed ON quiz_submissions(completed_at);
CREATE INDEX idx_quiz_submissions_expires ON quiz_submissions(expires_at);

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

    -- Timing
    answered_at TIMESTAMPTZ DEFAULT now(),
    time_spent_seconds INT,

    UNIQUE (submission_id, question_id)
);

CREATE INDEX idx_quiz_responses_submission ON quiz_responses(submission_id);
CREATE INDEX idx_quiz_responses_question ON quiz_responses(question_id);

-- ==================== Achievement Integration ====================

CREATE TABLE quiz_achievements (
    achievement_id CHAR(28) PRIMARY KEY REFERENCES achievements(id) ON DELETE CASCADE,
    quiz_id CHAR(28) NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,

    -- Requirements for earning the achievement
    min_score_percentage INT,     -- e.g., 80 means need 80% correct to earn
    require_completion BOOLEAN DEFAULT true NOT NULL,

    CHECK (min_score_percentage >= 0 AND min_score_percentage <= 100)
);

-- ==================== Translation Tables ====================

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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS quiz_answer_translations;
DROP TABLE IF EXISTS quiz_question_translations;
DROP TABLE IF EXISTS quiz_translations;
DROP TABLE IF EXISTS quiz_achievements;
DROP TABLE IF EXISTS quiz_responses;
DROP TABLE IF EXISTS quiz_submissions;
DROP TABLE IF EXISTS quiz_predefined_answers;
DROP TABLE IF EXISTS quiz_questions;
DROP TABLE IF EXISTS quizzes;

-- +goose StatementEnd
