-- Insert Free-Text Quiz Script
-- This script inserts a quiz challenge with four questions (2x FREE_TEXT,
-- 1x multi-select PREDEFINED where every answer is correct, 1x NUMBER) that
-- awards nothing: completion_points = 0, no per-question points, and no quiz
-- achievements. It also creates an OPEN quiz session and grants access to
-- every user, so all load-test users can start the quiz immediately.
--
-- Used by k6/freetext-quiz-spike.js.
--
-- Prerequisites:
--   - An existing project (use :project_id)
--   - Seeded users (make seed-large)
--
-- Usage:
--   psql -v project_id="'PR01K9W1YDE2QSV2WH4NQJAPMQ1A'" -f insert_freetext_quiz.sql

-- ID format: 2 char prefix + 26 chars [0-9A-Z] = 28 chars total

-- Clean up any previous test data (cascades take care of questions,
-- sessions and access grants; submissions must go first because
-- quiz_submissions.session_id has no ON DELETE clause)
DELETE FROM quiz_submissions WHERE quiz_id = 'QZ01LOADTESTFREETEXT00000000';
DELETE FROM quizzes WHERE id = 'QZ01LOADTESTFREETEXT00000000';
DELETE FROM challenges WHERE id = 'CL01LOADTESTFREETEXT00000000';

-- Quiz challenge (url must be NULL for QUIZ type); published and visible now
INSERT INTO challenges (
    id,
    project_id,
    challenge_type,
    name,
    description,
    button_text,
    notification_text,
    published_at,
    visible_at,
    started_at
) VALUES (
    'CL01LOADTESTFREETEXT00000000',
    :project_id,
    'QUIZ',
    'Load Test Free-Text Quiz',
    'A single free-text question quiz for load testing. Awards no points.',
    'Open quiz',
    '',
    NOW(),
    NOW(),
    NOW()
);

-- Quiz with zero completion points (publishing lives on the challenge;
-- quizzes.published_at was dropped in migration 00076)
INSERT INTO quizzes (
    id,
    project_id,
    challenge_id,
    name,
    description,
    timeout_seconds,
    randomize_questions,
    reveal_correct_answers,
    allow_retakes,
    completion_points
) VALUES (
    'QZ01LOADTESTFREETEXT00000000',
    :project_id,
    'CL01LOADTESTFREETEXT00000000',
    'Load Test Free-Text Quiz',
    'A quiz with two free-text, one multiple-choice and one number question, and no points.',
    1800,
    false,
    false,
    true,
    0
);

-- Questions (points stays NULL everywhere so nothing is ever awarded)
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection, min_value, max_value, step_value)
VALUES
    ('QQ01LOADTESTFREETEXT00000001', 'QZ01LOADTESTFREETEXT00000000', 'FREE_TEXT', 'What did you learn today?', 1, false, NULL, NULL, NULL),
    ('QQ01LOADTESTFREETEXT00000002', 'QZ01LOADTESTFREETEXT00000000', 'FREE_TEXT', 'What will you do differently tomorrow?', 2, false, NULL, NULL, NULL),
    ('QQ01LOADTESTFREETEXT00000003', 'QZ01LOADTESTFREETEXT00000000', 'PREDEFINED', 'Which of these did you practice this week? (select all)', 3, true, NULL, NULL, NULL),
    ('QQ01LOADTESTFREETEXT00000004', 'QZ01LOADTESTFREETEXT00000000', 'NUMBER', 'How many minutes did you spend reading today?', 4, false, 1, 100, 1);

-- Every answer is flagged correct, so a VU that selects all of them grades as
-- correct (grading is an exact-set match against the correct answers)
INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order)
VALUES
    ('QA01LOADTESTFREETEXT00000001', 'QQ01LOADTESTFREETEXT00000003', 'Patience', true, 1),
    ('QA01LOADTESTFREETEXT00000002', 'QQ01LOADTESTFREETEXT00000003', 'Gratitude', true, 2),
    ('QA01LOADTESTFREETEXT00000003', 'QQ01LOADTESTFREETEXT00000003', 'Kindness', true, 3),
    ('QA01LOADTESTFREETEXT00000004', 'QQ01LOADTESTFREETEXT00000003', 'Listening', true, 4);

-- An OPEN session so users can start the quiz right away
INSERT INTO quiz_sessions (id, quiz_id, name, state, created_by)
VALUES (
    'QN01LOADTESTFREETEXT00000000',
    'QZ01LOADTESTFREETEXT00000000',
    'Load Test Free-Text Session',
    'OPEN',
    (SELECT id FROM users ORDER BY id LIMIT 1)
);

-- Grant session access to every user. Access IDs are QX + 14 fixed chars
-- + 12 uppercase hex chars (hex is a subset of the [0-9A-Z] ID check).
INSERT INTO quiz_session_access (id, session_id, user_id, granted_by, source_type)
SELECT
    'QX01LOADTESTFTAX' || upper(lpad(to_hex(row_number() OVER (ORDER BY u.id))::text, 12, '0')),
    'QN01LOADTESTFREETEXT00000000',
    u.id,
    (SELECT id FROM users ORDER BY id LIMIT 1),
    'ALL'
FROM users u;

-- Output confirmation
SELECT
    'Quiz inserted: ' || q.name AS status,
    q.id AS quiz_id,
    'CL01LOADTESTFREETEXT00000000' AS challenge_id,
    'QN01LOADTESTFREETEXT00000000' AS session_id,
    (SELECT COUNT(*) FROM quiz_questions WHERE quiz_id = q.id) AS question_count,
    (SELECT COUNT(*) FROM quiz_session_access WHERE session_id = 'QN01LOADTESTFREETEXT00000000') AS users_with_access,
    q.completion_points AS points
FROM quizzes q
WHERE q.id = 'QZ01LOADTESTFREETEXT00000000';
