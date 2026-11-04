-- +goose Up
-- +goose StatementBegin

-- MAP_PIN: the participant drops one pin on the question's image (a map).
-- Coordinates are natural pixels of that image, origin top-left, y down.
-- See notes/quiz-map-pin-question.md and notes/quiz-map-pin-coordinates.md.
ALTER TABLE quiz_questions
DROP CONSTRAINT quiz_questions_question_type_check;

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_question_type_check
CHECK (question_type IN ('PREDEFINED', 'FREE_TEXT', 'NUMBER', 'JSON', 'ORDERING', 'MAP_PIN'));

ALTER TABLE quiz_questions
    ADD COLUMN map_scoring_mode VARCHAR(16),
    ADD COLUMN map_width INT,
    ADD COLUMN map_height INT,
    -- POINT_RADIUS: full points within the full radius of the target, falling
    -- linearly to 0 at the zero radius.
    ADD COLUMN map_target_x DOUBLE PRECISION,
    ADD COLUMN map_target_y DOUBLE PRECISION,
    ADD COLUMN map_full_points_radius DOUBLE PRECISION,
    ADD COLUMN map_zero_points_radius DOUBLE PRECISION,
    -- POLYGON: full points inside or on the boundary, 0 outside.
    -- [{"x": .., "y": ..}, ...] in drawing order, without a closing vertex.
    ADD COLUMN map_polygon JSONB;

-- Each row has exactly the map configuration its type and mode need. Geometric
-- validity (a simple polygon with non-zero area) is checked in Go
-- (quizgrading.ValidateMapPinConfig); a CHECK cannot express it.
--
-- A CHECK passes on NULL, so the CASE is wrapped in COALESCE(.., false): a
-- missing column makes a comparison NULL, which must fail, not pass.
-- jsonb_array_length raises on a non-array, and AND does not guarantee
-- evaluation order, so the array test guards it through a nested CASE.
ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_map_pin_check
CHECK (COALESCE(
    CASE
        WHEN question_type <> 'MAP_PIN' THEN
            map_scoring_mode IS NULL
            AND map_width IS NULL
            AND map_height IS NULL
            AND map_target_x IS NULL
            AND map_target_y IS NULL
            AND map_full_points_radius IS NULL
            AND map_zero_points_radius IS NULL
            AND map_polygon IS NULL
        WHEN map_scoring_mode = 'POINT_RADIUS' THEN
            map_width > 0
            AND map_height > 0
            AND points > 0
            AND NOT betting_enabled
            AND map_target_x BETWEEN 0 AND map_width
            AND map_target_y BETWEEN 0 AND map_height
            AND map_full_points_radius >= 0
            AND map_zero_points_radius > map_full_points_radius
            AND map_polygon IS NULL
        WHEN map_scoring_mode = 'POLYGON' THEN
            map_width > 0
            AND map_height > 0
            AND points > 0
            AND NOT betting_enabled
            AND CASE
                WHEN jsonb_typeof(map_polygon) = 'array'
                    THEN jsonb_array_length(map_polygon) BETWEEN 3 AND 100
                ELSE false
            END
            AND map_target_x IS NULL
            AND map_target_y IS NULL
            AND map_full_points_radius IS NULL
            AND map_zero_points_radius IS NULL
        ELSE false
    END,
    false
));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Fails while MAP_PIN questions exist: the restored type check rejects them.
-- Delete or convert them first; this migration does not drop them silently.
ALTER TABLE quiz_questions
DROP CONSTRAINT quiz_questions_map_pin_check;

ALTER TABLE quiz_questions
    DROP COLUMN map_polygon,
    DROP COLUMN map_zero_points_radius,
    DROP COLUMN map_full_points_radius,
    DROP COLUMN map_target_y,
    DROP COLUMN map_target_x,
    DROP COLUMN map_height,
    DROP COLUMN map_width,
    DROP COLUMN map_scoring_mode;

ALTER TABLE quiz_questions
DROP CONSTRAINT quiz_questions_question_type_check;

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_question_type_check
CHECK (question_type IN ('PREDEFINED', 'FREE_TEXT', 'NUMBER', 'JSON', 'ORDERING'));

-- +goose StatementEnd
