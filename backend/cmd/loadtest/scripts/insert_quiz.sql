-- Insert Quiz Script
-- This script inserts a quiz with 10 predefined questions, each with 4 choices.
-- The first answer for each question is always correct.
--
-- Prerequisites:
--   - An existing project (use :project_id)
--   - An existing challenge in that project (use :challenge_id)
--
-- Usage:
--   psql -v project_id="'PR01K9W1YDE2QSV2WH4NQJAPMQ1A'" -v challenge_id="'CL01K9VZ88BVVQXDY3WMJA21R3A4'" -f insert_quiz.sql

-- ID format: 2 char prefix + 26 chars [0-9A-Z] = 28 chars total
-- Real example: PR01K9W1YDE2QSV2WH4NQJAPMQ1A
--               PR (2) + 01K9W1YDE2QSV2WH4NQJAPMQ1A (26) = 28

-- First delete any existing test quiz submissions and data
DELETE FROM quiz_submissions WHERE quiz_id = 'QZ01LOADTESTQUIZ000000000000';
DELETE FROM quizzes WHERE id = 'QZ01LOADTESTQUIZ000000000000';

-- Insert the quiz
-- QZ (2) + 01LOADTESTQUIZ000000000000 (26) = 28
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
    completion_points,
    published_at
) VALUES (
    'QZ01LOADTESTQUIZ000000000000',
    :project_id,
    :challenge_id,
    'Load Test Quiz',
    'A quiz designed for load testing with 10 questions and 4 choices each.',
    1800,
    false,
    true,
    true,
    100,
    NOW()
);

-- Question IDs: QQ (2) + 01LOADTESTQ001000000000000 (26) = 28
-- Answer IDs:   QA (2) + 01LOADTEST001A0000000000001 (26) = 28

-- Question 1
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ001000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'What is the capital of France?', 1, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST001A000000000001', 'QQ01LOADTESTQ001000000000000', 'Paris', true, 1),
    ('QA01LOADTEST001A000000000002', 'QQ01LOADTESTQ001000000000000', 'London', false, 2),
    ('QA01LOADTEST001A000000000003', 'QQ01LOADTESTQ001000000000000', 'Berlin', false, 3),
    ('QA01LOADTEST001A000000000004', 'QQ01LOADTESTQ001000000000000', 'Madrid', false, 4);

-- Question 2
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ002000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'Which planet is known as the Red Planet?', 2, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST002A000000000001', 'QQ01LOADTESTQ002000000000000', 'Mars', true, 1),
    ('QA01LOADTEST002A000000000002', 'QQ01LOADTESTQ002000000000000', 'Venus', false, 2),
    ('QA01LOADTEST002A000000000003', 'QQ01LOADTESTQ002000000000000', 'Jupiter', false, 3),
    ('QA01LOADTEST002A000000000004', 'QQ01LOADTESTQ002000000000000', 'Saturn', false, 4);

-- Question 3
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ003000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'What is the largest ocean on Earth?', 3, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST003A000000000001', 'QQ01LOADTESTQ003000000000000', 'Pacific Ocean', true, 1),
    ('QA01LOADTEST003A000000000002', 'QQ01LOADTESTQ003000000000000', 'Atlantic Ocean', false, 2),
    ('QA01LOADTEST003A000000000003', 'QQ01LOADTESTQ003000000000000', 'Indian Ocean', false, 3),
    ('QA01LOADTEST003A000000000004', 'QQ01LOADTESTQ003000000000000', 'Arctic Ocean', false, 4);

-- Question 4
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ004000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'Who wrote "Romeo and Juliet"?', 4, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST004A000000000001', 'QQ01LOADTESTQ004000000000000', 'William Shakespeare', true, 1),
    ('QA01LOADTEST004A000000000002', 'QQ01LOADTESTQ004000000000000', 'Charles Dickens', false, 2),
    ('QA01LOADTEST004A000000000003', 'QQ01LOADTESTQ004000000000000', 'Jane Austen', false, 3),
    ('QA01LOADTEST004A000000000004', 'QQ01LOADTESTQ004000000000000', 'Mark Twain', false, 4);

-- Question 5
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ005000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'What is the chemical symbol for gold?', 5, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST005A000000000001', 'QQ01LOADTESTQ005000000000000', 'Au', true, 1),
    ('QA01LOADTEST005A000000000002', 'QQ01LOADTESTQ005000000000000', 'Ag', false, 2),
    ('QA01LOADTEST005A000000000003', 'QQ01LOADTESTQ005000000000000', 'Fe', false, 3),
    ('QA01LOADTEST005A000000000004', 'QQ01LOADTESTQ005000000000000', 'Cu', false, 4);

-- Question 6
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ006000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'How many continents are there on Earth?', 6, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST006A000000000001', 'QQ01LOADTESTQ006000000000000', '7', true, 1),
    ('QA01LOADTEST006A000000000002', 'QQ01LOADTESTQ006000000000000', '5', false, 2),
    ('QA01LOADTEST006A000000000003', 'QQ01LOADTESTQ006000000000000', '6', false, 3),
    ('QA01LOADTEST006A000000000004', 'QQ01LOADTESTQ006000000000000', '8', false, 4);

-- Question 7
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ007000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'What is the speed of light in a vacuum?', 7, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST007A000000000001', 'QQ01LOADTESTQ007000000000000', '299,792 km/s', true, 1),
    ('QA01LOADTEST007A000000000002', 'QQ01LOADTESTQ007000000000000', '150,000 km/s', false, 2),
    ('QA01LOADTEST007A000000000003', 'QQ01LOADTESTQ007000000000000', '500,000 km/s', false, 3),
    ('QA01LOADTEST007A000000000004', 'QQ01LOADTESTQ007000000000000', '1,000,000 km/s', false, 4);

-- Question 8
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ008000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'Which element has the atomic number 1?', 8, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST008A000000000001', 'QQ01LOADTESTQ008000000000000', 'Hydrogen', true, 1),
    ('QA01LOADTEST008A000000000002', 'QQ01LOADTESTQ008000000000000', 'Helium', false, 2),
    ('QA01LOADTEST008A000000000003', 'QQ01LOADTESTQ008000000000000', 'Oxygen', false, 3),
    ('QA01LOADTEST008A000000000004', 'QQ01LOADTESTQ008000000000000', 'Carbon', false, 4);

-- Question 9
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ009000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'In what year did World War II end?', 9, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST009A000000000001', 'QQ01LOADTESTQ009000000000000', '1945', true, 1),
    ('QA01LOADTEST009A000000000002', 'QQ01LOADTESTQ009000000000000', '1944', false, 2),
    ('QA01LOADTEST009A000000000003', 'QQ01LOADTESTQ009000000000000', '1946', false, 3),
    ('QA01LOADTEST009A000000000004', 'QQ01LOADTESTQ009000000000000', '1943', false, 4);

-- Question 10
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order, allow_multiple_selection)
VALUES ('QQ01LOADTESTQ010000000000000', 'QZ01LOADTESTQUIZ000000000000', 'PREDEFINED', 'What is the largest mammal in the world?', 10, false);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order) VALUES
    ('QA01LOADTEST010A000000000001', 'QQ01LOADTESTQ010000000000000', 'Blue Whale', true, 1),
    ('QA01LOADTEST010A000000000002', 'QQ01LOADTESTQ010000000000000', 'African Elephant', false, 2),
    ('QA01LOADTEST010A000000000003', 'QQ01LOADTESTQ010000000000000', 'Giraffe', false, 3),
    ('QA01LOADTEST010A000000000004', 'QQ01LOADTESTQ010000000000000', 'Hippopotamus', false, 4);

-- Output confirmation
SELECT
    'Quiz inserted: ' || q.name AS status,
    q.id AS quiz_id,
    COUNT(qq.id) AS question_count,
    q.completion_points AS points
FROM quizzes q
LEFT JOIN quiz_questions qq ON qq.quiz_id = q.id
WHERE q.id = 'QZ01LOADTESTQUIZ000000000000'
GROUP BY q.id, q.name, q.completion_points;
