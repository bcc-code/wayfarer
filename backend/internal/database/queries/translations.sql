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

-- name: GetTeamTranslationsByIDs :many
SELECT team_id, language_code, name, description
FROM team_translations
WHERE team_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetSuperTeamTranslationsByIDs :many
SELECT super_team_id, language_code, name, description
FROM super_team_translations
WHERE super_team_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetStreakTranslationsByIDs :many
SELECT streak_id, language_code, name, description
FROM streak_translations
WHERE streak_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetChallengeTranslationsByIDs :many
SELECT challenge_id, language_code, name, description, button_text
FROM challenge_translations
WHERE challenge_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetAchievementTranslationsByIDs :many
SELECT achievement_id, language_code, name, description
FROM achievement_translations
WHERE achievement_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- Delete queries for translation invalidation when base content changes

-- name: DeleteProjectTranslations :exec
DELETE FROM project_translations WHERE project_id = @project_id::text;

-- name: DeleteEventTranslations :exec
DELETE FROM event_translations WHERE event_id = @event_id::text;

-- name: DeleteTeamTranslations :exec
DELETE FROM team_translations WHERE team_id = @team_id::text;

-- name: DeleteSuperTeamTranslations :exec
DELETE FROM super_team_translations WHERE super_team_id = @super_team_id::text;

-- name: DeleteStreakTranslations :exec
DELETE FROM streak_translations WHERE streak_id = @streak_id::text;

-- name: DeleteChallengeTranslations :exec
DELETE FROM challenge_translations WHERE challenge_id = @challenge_id::text;

-- name: DeleteAchievementTranslations :exec
DELETE FROM achievement_translations WHERE achievement_id = @achievement_id::text;
