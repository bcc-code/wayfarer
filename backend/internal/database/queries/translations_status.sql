-- name: GetChallengeTranslationStatus :many
SELECT language_code,
  (name IS NOT NULL AND name != '') AS has_name,
  (description IS NOT NULL AND description != '') AS has_description,
  (button_text IS NOT NULL AND button_text != '') AS has_button_text,
  (notification_text IS NOT NULL AND notification_text != '') AS has_notification_text
FROM challenge_translations
WHERE challenge_id = @challenge_id::text;

-- name: GetProjectTranslationStatus :many
SELECT language_code,
  (name IS NOT NULL AND name != '') AS has_name,
  (description IS NOT NULL AND description != '') AS has_description,
  (rules IS NOT NULL AND rules != '') AS has_rules
FROM project_translations
WHERE project_id = @project_id::text;

-- name: GetEventTranslationStatus :many
SELECT language_code,
  (name IS NOT NULL AND name != '') AS has_name,
  (description IS NOT NULL AND description != '') AS has_description
FROM event_translations
WHERE event_id = @event_id::text;

-- name: GetAchievementTranslationStatus :many
SELECT language_code,
  (name IS NOT NULL AND name != '') AS has_name,
  (description_pending IS NOT NULL AND description_pending != '') AS has_description_pending,
  (description_completed IS NOT NULL AND description_completed != '') AS has_description_completed,
  (notification_text IS NOT NULL AND notification_text != '') AS has_notification_text
FROM achievement_translations
WHERE achievement_id = @achievement_id::text;

-- name: GetConsentTranslationStatus :many
SELECT language_code,
  (title IS NOT NULL AND title != '') AS has_title,
  (short_text IS NOT NULL AND short_text != '') AS has_short_text,
  (body IS NOT NULL AND body != '') AS has_body
FROM consent_translations
WHERE consent_id = @consent_id::text;

-- name: GetQuizTranslationStatus :many
SELECT language_code,
  (name IS NOT NULL AND name != '') AS has_name,
  (description IS NOT NULL AND description != '') AS has_description
FROM quiz_translations
WHERE quiz_id = @quiz_id::text;

-- name: GetQuizQuestionTranslationStatus :many
SELECT language_code,
  (question_text IS NOT NULL AND question_text != '') AS has_question_text
FROM quiz_question_translations
WHERE question_id = @question_id::text;

-- name: GetQuizAnswerTranslationStatus :many
SELECT language_code,
  (answer_text IS NOT NULL AND answer_text != '') AS has_answer_text
FROM quiz_answer_translations
WHERE answer_id = @answer_id::text;
