-- Translation queries for i18n support
-- Each query fetches translations for multiple entity IDs with a specific language code

-- name: GetProjectTranslationsByIDs :many
SELECT project_id, language_code, name, description, rules
FROM project_translations
WHERE project_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetEventTranslationsByIDs :many
SELECT event_id, language_code, name, description
FROM event_translations
WHERE event_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetStreakTranslationsByIDs :many
SELECT streak_id, language_code, name, description
FROM streak_translations
WHERE streak_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetChallengeTranslationsByIDs :many
SELECT challenge_id, language_code, name, description, button_text, notification_text
FROM challenge_translations
WHERE challenge_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetAchievementTranslationsByIDs :many
SELECT achievement_id, language_code, name, description_pending, description_completed, notification_text
FROM achievement_translations
WHERE achievement_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetQuizTranslationsByIDs :many
SELECT quiz_id, language_code, name, description
FROM quiz_translations
WHERE quiz_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetQuizQuestionTranslationsByIDs :many
SELECT question_id, language_code, question_text
FROM quiz_question_translations
WHERE question_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetQuizAnswerTranslationsByIDs :many
SELECT answer_id, language_code, answer_text
FROM quiz_answer_translations
WHERE answer_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- Delete queries for translation invalidation when base content changes

-- name: DeleteProjectTranslations :exec
DELETE FROM project_translations WHERE project_id = @project_id::text;

-- name: DeleteEventTranslations :exec
DELETE FROM event_translations WHERE event_id = @event_id::text;

-- name: DeleteStreakTranslations :exec
DELETE FROM streak_translations WHERE streak_id = @streak_id::text;

-- name: DeleteChallengeTranslations :exec
DELETE FROM challenge_translations WHERE challenge_id = @challenge_id::text;

-- name: DeleteAchievementTranslations :exec
DELETE FROM achievement_translations WHERE achievement_id = @achievement_id::text;
