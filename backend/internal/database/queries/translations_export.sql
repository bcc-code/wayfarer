-- Export queries for translations - fetch base language content

-- name: GetProjectsForTranslation :many
SELECT id, name, description, rules
FROM projects
WHERE archived = false;

-- name: GetEventsForTranslation :many
SELECT id, name, description
FROM events;

-- name: GetStreaksForTranslation :many
SELECT id, name, description
FROM streaks;

-- name: GetChallengesForTranslation :many
SELECT id, name, description, button_text
FROM challenges
WHERE published_at IS NOT NULL;

-- name: GetAchievementsForTranslation :many
SELECT id, name, description_pending, description_completed, notification_text
FROM achievements
WHERE hidden = false;

-- name: GetQuizzesForTranslation :many
SELECT id, name, description
FROM quizzes;

-- name: GetQuizQuestionsForTranslation :many
SELECT id, question_text
FROM quiz_questions
WHERE quiz_id = @quiz_id::text
ORDER BY question_order;

-- name: GetQuizAnswersForTranslation :many
SELECT id, answer_text
FROM quiz_predefined_answers
WHERE question_id = @question_id::text
ORDER BY answer_order;

-- name: GetConsentsForTranslation :many
SELECT id, title, short_text, body
FROM consents;
