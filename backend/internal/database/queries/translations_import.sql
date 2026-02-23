-- Import queries - upsert translated content

-- name: UpsertProjectTranslation :exec
INSERT INTO project_translations (project_id, language_code, name, description, rules, updated_at)
VALUES (@project_id::text, @language_code::text, @name::text, @description::text, @rules::text, now())
ON CONFLICT (project_id, language_code)
DO UPDATE SET
    name = COALESCE(NULLIF(@name::text, ''), project_translations.name),
    description = COALESCE(NULLIF(@description::text, ''), project_translations.description),
    rules = COALESCE(NULLIF(@rules::text, ''), project_translations.rules),
    updated_at = now();

-- name: UpsertEventTranslation :exec
INSERT INTO event_translations (event_id, language_code, name, description, updated_at)
VALUES (@event_id::text, @language_code::text, @name::text, @description::text, now())
ON CONFLICT (event_id, language_code)
DO UPDATE SET
    name = COALESCE(NULLIF(@name::text, ''), event_translations.name),
    description = COALESCE(NULLIF(@description::text, ''), event_translations.description),
    updated_at = now();

-- name: UpsertStreakTranslation :exec
INSERT INTO streak_translations (streak_id, language_code, name, description, updated_at)
VALUES (@streak_id::text, @language_code::text, @name::text, @description::text, now())
ON CONFLICT (streak_id, language_code)
DO UPDATE SET
    name = COALESCE(NULLIF(@name::text, ''), streak_translations.name),
    description = COALESCE(NULLIF(@description::text, ''), streak_translations.description),
    updated_at = now();

-- name: UpsertChallengeTranslation :exec
INSERT INTO challenge_translations (challenge_id, language_code, name, description, button_text, notification_text, updated_at)
VALUES (@challenge_id::text, @language_code::text, @name::text, @description::text, @button_text::text, @notification_text::text, now())
ON CONFLICT (challenge_id, language_code)
DO UPDATE SET
    name = COALESCE(NULLIF(@name::text, ''), challenge_translations.name),
    description = COALESCE(NULLIF(@description::text, ''), challenge_translations.description),
    button_text = COALESCE(NULLIF(@button_text::text, ''), challenge_translations.button_text),
    notification_text = COALESCE(NULLIF(@notification_text::text, ''), challenge_translations.notification_text),
    updated_at = now();

-- name: UpsertAchievementTranslation :exec
INSERT INTO achievement_translations (
    achievement_id, language_code, name,
    description_pending, description_completed, notification_text, updated_at
)
VALUES (
    @achievement_id::text, @language_code::text, @name::text,
    @description_pending::text, @description_completed::text, @notification_text::text, now()
)
ON CONFLICT (achievement_id, language_code)
DO UPDATE SET
    name = COALESCE(NULLIF(@name::text, ''), achievement_translations.name),
    description_pending = COALESCE(NULLIF(@description_pending::text, ''), achievement_translations.description_pending),
    description_completed = COALESCE(NULLIF(@description_completed::text, ''), achievement_translations.description_completed),
    notification_text = COALESCE(NULLIF(@notification_text::text, ''), achievement_translations.notification_text),
    updated_at = now();

-- name: UpsertQuizTranslation :exec
INSERT INTO quiz_translations (quiz_id, language_code, name, description, updated_at)
VALUES (@quiz_id::text, @language_code::text, @name::text, @description::text, now())
ON CONFLICT (quiz_id, language_code)
DO UPDATE SET
    name = COALESCE(NULLIF(@name::text, ''), quiz_translations.name),
    description = COALESCE(NULLIF(@description::text, ''), quiz_translations.description),
    updated_at = now();

-- name: UpsertQuizQuestionTranslation :exec
INSERT INTO quiz_question_translations (question_id, language_code, question_text, updated_at)
VALUES (@question_id::text, @language_code::text, @question_text::text, now())
ON CONFLICT (question_id, language_code)
DO UPDATE SET
    question_text = COALESCE(NULLIF(@question_text::text, ''), quiz_question_translations.question_text),
    updated_at = now();

-- name: UpsertQuizAnswerTranslation :exec
INSERT INTO quiz_answer_translations (answer_id, language_code, answer_text, updated_at)
VALUES (@answer_id::text, @language_code::text, @answer_text::text, now())
ON CONFLICT (answer_id, language_code)
DO UPDATE SET
    answer_text = COALESCE(NULLIF(@answer_text::text, ''), quiz_answer_translations.answer_text),
    updated_at = now();

-- name: UpsertConsentTranslationFromPhrase :exec
INSERT INTO consent_translations (consent_id, language_code, title, short_text, body, updated_at)
VALUES (@consent_id::text, @language_code::text, @title::text, @short_text::text, @body::text, now())
ON CONFLICT (consent_id, language_code)
DO UPDATE SET
    title = COALESCE(NULLIF(@title::text, ''), consent_translations.title),
    short_text = COALESCE(NULLIF(@short_text::text, ''), consent_translations.short_text),
    body = COALESCE(NULLIF(@body::text, ''), consent_translations.body),
    updated_at = now();
