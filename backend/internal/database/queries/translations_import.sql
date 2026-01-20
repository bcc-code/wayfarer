-- Import queries - upsert translated content

-- name: UpsertProjectTranslation :exec
INSERT INTO project_translations (project_id, language_code, name, description, rules, updated_at)
VALUES (@project_id::text, @language_code::text, @name::text, @description::text, @rules::text, now())
ON CONFLICT (project_id, language_code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    rules = EXCLUDED.rules,
    updated_at = now();

-- name: UpsertEventTranslation :exec
INSERT INTO event_translations (event_id, language_code, name, description, updated_at)
VALUES (@event_id::text, @language_code::text, @name::text, @description::text, now())
ON CONFLICT (event_id, language_code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = now();

-- name: UpsertSuperTeamTranslation :exec
INSERT INTO super_team_translations (super_team_id, language_code, name, description, updated_at)
VALUES (@super_team_id::text, @language_code::text, @name::text, @description::text, now())
ON CONFLICT (super_team_id, language_code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = now();

-- name: UpsertStreakTranslation :exec
INSERT INTO streak_translations (streak_id, language_code, name, description, updated_at)
VALUES (@streak_id::text, @language_code::text, @name::text, @description::text, now())
ON CONFLICT (streak_id, language_code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = now();

-- name: UpsertChallengeTranslation :exec
INSERT INTO challenge_translations (challenge_id, language_code, name, description, button_text, updated_at)
VALUES (@challenge_id::text, @language_code::text, @name::text, @description::text, @button_text::text, now())
ON CONFLICT (challenge_id, language_code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    button_text = EXCLUDED.button_text,
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
    name = EXCLUDED.name,
    description_pending = EXCLUDED.description_pending,
    description_completed = EXCLUDED.description_completed,
    notification_text = EXCLUDED.notification_text,
    updated_at = now();

-- name: UpsertQuizTranslation :exec
INSERT INTO quiz_translations (quiz_id, language_code, name, description, updated_at)
VALUES (@quiz_id::text, @language_code::text, @name::text, @description::text, now())
ON CONFLICT (quiz_id, language_code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = now();

-- name: UpsertQuizQuestionTranslation :exec
INSERT INTO quiz_question_translations (question_id, language_code, question_text, updated_at)
VALUES (@question_id::text, @language_code::text, @question_text::text, now())
ON CONFLICT (question_id, language_code)
DO UPDATE SET
    question_text = EXCLUDED.question_text,
    updated_at = now();

-- name: UpsertQuizAnswerTranslation :exec
INSERT INTO quiz_answer_translations (answer_id, language_code, answer_text, updated_at)
VALUES (@answer_id::text, @language_code::text, @answer_text::text, now())
ON CONFLICT (answer_id, language_code)
DO UPDATE SET
    answer_text = EXCLUDED.answer_text,
    updated_at = now();

-- name: UpsertConsentTranslationFromPhrase :exec
INSERT INTO consent_translations (consent_id, language_code, title, short_text, body, updated_at)
VALUES (@consent_id::text, @language_code::text, @title::text, @short_text::text, @body::text, now())
ON CONFLICT (consent_id, language_code)
DO UPDATE SET
    title = EXCLUDED.title,
    short_text = EXCLUDED.short_text,
    body = EXCLUDED.body,
    updated_at = now();
