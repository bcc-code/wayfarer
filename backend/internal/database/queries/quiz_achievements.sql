-- name: GetQuizAchievementByAchievementID :one
SELECT achievement_id, quiz_id, min_score_percentage, require_completion
FROM quiz_achievements
WHERE achievement_id = @achievementid::text;

-- name: GetQuizAchievementsByQuizID :many
SELECT achievement_id, quiz_id, min_score_percentage, require_completion
FROM quiz_achievements
WHERE quiz_id = @quizid::text;

-- name: GetQuizAchievementsByQuizIDs :many
SELECT achievement_id, quiz_id, min_score_percentage, require_completion
FROM quiz_achievements
WHERE quiz_id = ANY(@quiz_ids::text[]);

-- name: CreateQuizAchievement :one
INSERT INTO quiz_achievements (
    achievement_id,
    quiz_id,
    min_score_percentage,
    require_completion
)
VALUES (
    @achievementid::text,
    @quizid::text,
    sqlc.narg('minscorepercentage')::int,
    @requirecompletion::bool
)
RETURNING achievement_id, quiz_id, min_score_percentage, require_completion;

-- name: DeleteQuizAchievement :exec
DELETE FROM quiz_achievements
WHERE achievement_id = @achievementid::text;

-- name: UpsertQuizAchievement :one
INSERT INTO quiz_achievements (
    achievement_id,
    quiz_id,
    min_score_percentage,
    require_completion
)
VALUES (
    @achievementid::text,
    @quizid::text,
    sqlc.narg('minscorepercentage')::int,
    @requirecompletion::bool
)
ON CONFLICT (achievement_id) DO UPDATE SET
    quiz_id = EXCLUDED.quiz_id,
    min_score_percentage = EXCLUDED.min_score_percentage,
    require_completion = EXCLUDED.require_completion
RETURNING achievement_id, quiz_id, min_score_percentage, require_completion;
