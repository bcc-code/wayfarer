-- name: GetQuizQuestionsByQuizID :many
SELECT id, quiz_id, question_type, question_text, image_url, question_order, allow_multiple_selection, answer_label_style, min_value, max_value, step_value, timeout_seconds, points, betting_enabled, betting_min_percentage, betting_max_percentage, betting_min_absolute, betting_max_absolute, betting_multiplier_correct, betting_multiplier_wrong, map_scoring_mode, map_width, map_height, map_target_x, map_target_y, map_full_points_radius, map_zero_points_radius, map_polygon, created_at, updated_at
FROM quiz_questions
WHERE quiz_id = @quizid::char(28)
ORDER BY question_order ASC;

-- name: GetQuizQuestionsByQuizIDs :many
SELECT id, quiz_id, question_type, question_text, image_url, question_order, allow_multiple_selection, answer_label_style, min_value, max_value, step_value, timeout_seconds, points, betting_enabled, betting_min_percentage, betting_max_percentage, betting_min_absolute, betting_max_absolute, betting_multiplier_correct, betting_multiplier_wrong, map_scoring_mode, map_width, map_height, map_target_x, map_target_y, map_full_points_radius, map_zero_points_radius, map_polygon, created_at, updated_at
FROM quiz_questions
WHERE quiz_id = ANY(@quiz_ids::char(28)[])
ORDER BY quiz_id, question_order ASC;

-- name: GetQuizQuestionByID :one
SELECT id, quiz_id, question_type, question_text, image_url, question_order, allow_multiple_selection, answer_label_style, min_value, max_value, step_value, timeout_seconds, points, betting_enabled, betting_min_percentage, betting_max_percentage, betting_min_absolute, betting_max_absolute, betting_multiplier_correct, betting_multiplier_wrong, map_scoring_mode, map_width, map_height, map_target_x, map_target_y, map_full_points_radius, map_zero_points_radius, map_polygon, created_at, updated_at
FROM quiz_questions
WHERE id = @id::char(28);

-- name: GetQuizQuestionsByIDs :many
SELECT id, quiz_id, question_type, question_text, image_url, question_order, allow_multiple_selection, answer_label_style, min_value, max_value, step_value, timeout_seconds, points, betting_enabled, betting_min_percentage, betting_max_percentage, betting_min_absolute, betting_max_absolute, betting_multiplier_correct, betting_multiplier_wrong, map_scoring_mode, map_width, map_height, map_target_x, map_target_y, map_full_points_radius, map_zero_points_radius, map_polygon, created_at, updated_at
FROM quiz_questions
WHERE id = ANY(@ids::char(28)[]);

-- name: CreateQuizQuestion :one
INSERT INTO quiz_questions (
    id,
    quiz_id,
    question_type,
    question_text,
    image_url,
    question_order,
    allow_multiple_selection,
    answer_label_style,
    min_value,
    max_value,
    step_value,
    timeout_seconds,
    points,
    betting_enabled,
    betting_min_percentage,
    betting_max_percentage,
    betting_min_absolute,
    betting_max_absolute,
    betting_multiplier_correct,
    betting_multiplier_wrong,
    map_scoring_mode,
    map_width,
    map_height,
    map_target_x,
    map_target_y,
    map_full_points_radius,
    map_zero_points_radius,
    map_polygon
)
VALUES (
    @id::text,
    @quizid::text,
    @questiontype::text,
    @questiontext::text,
    sqlc.narg('imageurl')::text,
    @questionorder::int,
    sqlc.narg('allowmultipleselection')::bool,
    COALESCE(sqlc.narg('answerlabelstyle')::text, 'NONE'),
    sqlc.narg('minvalue')::decimal,
    sqlc.narg('maxvalue')::decimal,
    sqlc.narg('stepvalue')::decimal,
    sqlc.narg('timeoutseconds')::int,
    sqlc.narg('points')::int,
    COALESCE(sqlc.narg('bettingenabled')::bool, false),
    sqlc.narg('bettingminpercentage')::decimal,
    sqlc.narg('bettingmaxpercentage')::decimal,
    sqlc.narg('bettingminabsolute')::int,
    sqlc.narg('bettingmaxabsolute')::int,
    sqlc.narg('bettingmultipliercorrect')::decimal,
    sqlc.narg('bettingmultiplierwrong')::decimal,
    -- Nullable params, not @x::type: a non-null param would write 0 instead
    -- of NULL and break quiz_questions_map_pin_check for every other type.
    sqlc.narg('mapscoringmode')::text,
    sqlc.narg('mapwidth')::int,
    sqlc.narg('mapheight')::int,
    sqlc.narg('maptargetx')::float8,
    sqlc.narg('maptargety')::float8,
    sqlc.narg('mapfullpointsradius')::float8,
    sqlc.narg('mapzeropointsradius')::float8,
    sqlc.narg('mappolygon')::jsonb
)
RETURNING id, quiz_id, question_type, question_text, image_url, question_order, allow_multiple_selection, answer_label_style, min_value, max_value, step_value, timeout_seconds, points, betting_enabled, betting_min_percentage, betting_max_percentage, betting_min_absolute, betting_max_absolute, betting_multiplier_correct, betting_multiplier_wrong, map_scoring_mode, map_width, map_height, map_target_x, map_target_y, map_full_points_radius, map_zero_points_radius, map_polygon, created_at, updated_at;

-- name: UpdateQuizQuestion :one
UPDATE quiz_questions
SET
    question_text = COALESCE(sqlc.narg('questiontext')::text, question_text),
    image_url = CASE WHEN @clearimage::bool THEN NULL
                     ELSE COALESCE(sqlc.narg('imageurl')::text, image_url) END,
    question_order = COALESCE(sqlc.narg('questionorder')::int, question_order),
    allow_multiple_selection = COALESCE(sqlc.narg('allowmultipleselection')::bool, allow_multiple_selection),
    answer_label_style = COALESCE(sqlc.narg('answerlabelstyle')::text, answer_label_style),
    min_value = COALESCE(sqlc.narg('minvalue')::decimal, min_value),
    max_value = COALESCE(sqlc.narg('maxvalue')::decimal, max_value),
    step_value = COALESCE(sqlc.narg('stepvalue')::decimal, step_value),
    timeout_seconds = COALESCE(sqlc.narg('timeoutseconds')::int, timeout_seconds),
    points = COALESCE(sqlc.narg('points')::int, points),
    betting_enabled = COALESCE(sqlc.narg('bettingenabled')::bool, betting_enabled),
    betting_min_percentage = CASE WHEN @clearbettingminpercentage::bool THEN NULL
                     ELSE COALESCE(sqlc.narg('bettingminpercentage')::decimal, betting_min_percentage) END,
    betting_max_percentage = CASE WHEN @clearbettingmaxpercentage::bool THEN NULL
                     ELSE COALESCE(sqlc.narg('bettingmaxpercentage')::decimal, betting_max_percentage) END,
    betting_min_absolute = CASE WHEN @clearbettingminabsolute::bool THEN NULL
                     ELSE COALESCE(sqlc.narg('bettingminabsolute')::int, betting_min_absolute) END,
    betting_max_absolute = CASE WHEN @clearbettingmaxabsolute::bool THEN NULL
                     ELSE COALESCE(sqlc.narg('bettingmaxabsolute')::int, betting_max_absolute) END,
    betting_multiplier_correct = CASE WHEN @clearbettingmultipliercorrect::bool THEN NULL
                     ELSE COALESCE(sqlc.narg('bettingmultipliercorrect')::decimal, betting_multiplier_correct) END,
    betting_multiplier_wrong = CASE WHEN @clearbettingmultiplierwrong::bool THEN NULL
                     ELSE COALESCE(sqlc.narg('bettingmultiplierwrong')::decimal, betting_multiplier_wrong) END,
    -- The map config is replaced as a whole when @setmappin is true, NULLs
    -- included, so switching scoring mode clears the other mode's columns.
    -- COALESCE would keep them and fail quiz_questions_map_pin_check.
    map_scoring_mode = CASE WHEN @setmappin::bool THEN sqlc.narg('mapscoringmode')::text ELSE map_scoring_mode END,
    map_width = CASE WHEN @setmappin::bool THEN sqlc.narg('mapwidth')::int ELSE map_width END,
    map_height = CASE WHEN @setmappin::bool THEN sqlc.narg('mapheight')::int ELSE map_height END,
    map_target_x = CASE WHEN @setmappin::bool THEN sqlc.narg('maptargetx')::float8 ELSE map_target_x END,
    map_target_y = CASE WHEN @setmappin::bool THEN sqlc.narg('maptargety')::float8 ELSE map_target_y END,
    map_full_points_radius = CASE WHEN @setmappin::bool THEN sqlc.narg('mapfullpointsradius')::float8 ELSE map_full_points_radius END,
    map_zero_points_radius = CASE WHEN @setmappin::bool THEN sqlc.narg('mapzeropointsradius')::float8 ELSE map_zero_points_radius END,
    map_polygon = CASE WHEN @setmappin::bool THEN sqlc.narg('mappolygon')::jsonb ELSE map_polygon END,
    updated_at = now()
WHERE id = @id::char(28)
RETURNING id, quiz_id, question_type, question_text, image_url, question_order, allow_multiple_selection, answer_label_style, min_value, max_value, step_value, timeout_seconds, points, betting_enabled, betting_min_percentage, betting_max_percentage, betting_min_absolute, betting_max_absolute, betting_multiplier_correct, betting_multiplier_wrong, map_scoring_mode, map_width, map_height, map_target_x, map_target_y, map_full_points_radius, map_zero_points_radius, map_polygon, created_at, updated_at;

-- name: DeleteQuizQuestion :exec
DELETE FROM quiz_questions
WHERE id = @id::char(28);

-- name: UpdateQuizQuestionOrder :exec
UPDATE quiz_questions
SET
    question_order = @questionorder::int,
    updated_at = now()
WHERE id = @id::char(28);

-- name: DeleteQuizQuestionTranslations :exec
DELETE FROM quiz_question_translations
WHERE question_id = @questionid::char(28);

-- name: GetPredefinedAnswersByQuestionID :many
SELECT id, question_id, answer_text, is_correct, answer_order, created_at, description
FROM quiz_predefined_answers
WHERE question_id = @questionid::char(28)
ORDER BY answer_order ASC;

-- name: GetPredefinedAnswersByQuestionIDs :many
SELECT id, question_id, answer_text, is_correct, answer_order, created_at, description
FROM quiz_predefined_answers
WHERE question_id = ANY(@question_ids::char(28)[])
ORDER BY question_id, answer_order ASC;

-- name: GetPredefinedAnswersByIDs :many
SELECT id, question_id, answer_text, is_correct, answer_order, created_at, description
FROM quiz_predefined_answers
WHERE id = ANY(@ids::char(28)[]);

-- name: CreatePredefinedAnswer :one
INSERT INTO quiz_predefined_answers (
    id,
    question_id,
    answer_text,
    is_correct,
    answer_order,
    description
)
VALUES (
    @id::text,
    @questionid::text,
    @answertext::text,
    @iscorrect::bool,
    @answerorder::int,
    -- A blank description means "no description", stored as NULL. BTRIM
    -- strips only spaces unless given the characters to strip.
    NULLIF(BTRIM(sqlc.narg('description')::text, E' \t\r\n'), '')
)
RETURNING id, question_id, answer_text, is_correct, answer_order, created_at, description;

-- name: DeletePredefinedAnswersByQuestion :exec
DELETE FROM quiz_predefined_answers
WHERE question_id = @questionid::char(28);

-- name: DeletePredefinedAnswerTranslations :exec
DELETE FROM quiz_answer_translations
WHERE answer_id = @answerid::char(28);
