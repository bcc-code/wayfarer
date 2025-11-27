-- Translation queries for i18n support
-- Each query fetches translations for multiple entity IDs with a specific language code

-- name: GetProjectTranslationsByIDs :many
SELECT project_id, language_code, name, description
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

-- name: GetArticleTranslationsByIDs :many
SELECT article_id, language_code, title, author
FROM reading_achievement_article_translations
WHERE article_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: GetTrackTranslationsByIDs :many
SELECT track_id, language_code, name, description
FROM listening_achievement_track_translations
WHERE track_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;
