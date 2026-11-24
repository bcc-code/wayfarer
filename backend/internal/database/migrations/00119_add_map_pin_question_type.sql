-- +goose Up
-- +goose StatementBegin

-- MAP_PIN: the participant drops one pin on the question's image (a map).
-- Coordinates are whole natural pixels of that image, origin top-left, y down.
-- Positions use the built-in point type (x = p[0], y = p[1]). It is float8
-- underneath; whole pixels are enforced by the CHECKs below and by Go, which
-- converts to integers as soon as a row is read.
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
    ADD COLUMN map_target POINT,
    ADD COLUMN map_full_points_radius INT,
    ADD COLUMN map_zero_points_radius INT,
    -- POLYGON: full points inside or on the boundary, 0 outside.
    -- Vertices in drawing order, without a closing vertex.
    ADD COLUMN map_polygon POINT[];

-- Each row has exactly the map configuration its type and mode need. Geometric
-- validity (a simple polygon with non-zero area) is checked in Go
-- (quizgrading.ValidateMapPinConfig); a CHECK cannot express it.
--
-- A CHECK passes on NULL, so the CASE is wrapped in COALESCE(.., false): a
-- missing column makes a comparison NULL, which must fail, not pass.
-- A CHECK can't loop over map_polygon's elements (and point has no equality
-- operator for array_position), so whole-pixel, non-NULL vertices are checked
-- in Go only; the target is checked here.
ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_map_pin_check
CHECK (COALESCE(
    CASE
        WHEN question_type <> 'MAP_PIN' THEN
            map_scoring_mode IS NULL
            AND map_width IS NULL
            AND map_height IS NULL
            AND map_target IS NULL
            AND map_full_points_radius IS NULL
            AND map_zero_points_radius IS NULL
            AND map_polygon IS NULL
        WHEN map_scoring_mode = 'POINT_RADIUS' THEN
            map_width > 0
            AND map_height > 0
            AND points > 0
            AND NOT betting_enabled
            AND map_target[0] BETWEEN 0 AND map_width
            AND map_target[1] BETWEEN 0 AND map_height
            AND map_target[0] = trunc(map_target[0])
            AND map_target[1] = trunc(map_target[1])
            AND map_full_points_radius >= 0
            AND map_zero_points_radius > map_full_points_radius
            AND map_polygon IS NULL
        WHEN map_scoring_mode = 'POLYGON' THEN
            map_width > 0
            AND map_height > 0
            AND points > 0
            AND NOT betting_enabled
            AND cardinality(map_polygon) BETWEEN 3 AND 100
            AND map_target IS NULL
            AND map_full_points_radius IS NULL
            AND map_zero_points_radius IS NULL
        ELSE false
    END,
    false
));

-- A MAP_PIN answer: the pin, in the same whole natural pixels. A typed point
-- rather than json_response, so every reader gets a point or nothing. A CHECK
-- can't see the question type; "MAP_PIN only" is enforced in Go.
ALTER TABLE quiz_responses
    ADD COLUMN map_position POINT,
    ADD CONSTRAINT quiz_responses_map_position_check CHECK (
        map_position IS NULL
        OR (map_position[0] = trunc(map_position[0])
            AND map_position[1] = trunc(map_position[1]))
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Fails while MAP_PIN questions exist: the restored type check rejects them.
-- Delete or convert them first; this migration does not drop them silently.
ALTER TABLE quiz_responses
    DROP CONSTRAINT quiz_responses_map_position_check,
    DROP COLUMN map_position;

ALTER TABLE quiz_questions
DROP CONSTRAINT quiz_questions_map_pin_check;

ALTER TABLE quiz_questions
    DROP COLUMN map_polygon,
    DROP COLUMN map_zero_points_radius,
    DROP COLUMN map_full_points_radius,
    DROP COLUMN map_target,
    DROP COLUMN map_height,
    DROP COLUMN map_width,
    DROP COLUMN map_scoring_mode;

ALTER TABLE quiz_questions
DROP CONSTRAINT quiz_questions_question_type_check;

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_question_type_check
CHECK (question_type IN ('PREDEFINED', 'FREE_TEXT', 'NUMBER', 'JSON', 'ORDERING'));

-- +goose StatementEnd
