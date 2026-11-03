# MAP_PIN quiz question ("Find the position in the temple")

A photo taken inside a 3D model of the temple is shown on a physical screen.
In the app, participants see a top-down map of the temple and drop a single
pin where they think the photo was taken. The pin is scored against the
correct answer, up to the question's maximum points.

Coordinates are covered in `quiz-map-pin-coordinates.md`.

## Model

`MAP_PIN` is a quiz question type, next to `PREDEFINED`, `FREE_TEXT`,
`NUMBER`, `JSON` and `ORDERING`. It reuses the quiz machinery: sessions
(OPEN while the group is at the screen; access can be granted per team),
`quiz_responses`, finalize and the QUIZ `score_journal` entry.

There is **one** question type with **two scoring modes**, because the
participant's interaction is identical in both — one pin on one map. Only the
correct-answer geometry and the scoring differ:

| Mode | Correct answer | Scoring |
|---|---|---|
| `POINT_RADIUS` | One point + `fullPointsRadius` + `zeroPointsRadius` | Linear partial credit by distance |
| `POLYGON` | An area drawn as an ordered list of vertices | Binary: inside (or on the boundary) = max, outside = 0 |

- **The map** is the question's existing `image_url`. The photo is shown only
  on the physical screen, so there is no second image column.
- **Maximum points** is the existing `quiz_questions.points` (required, > 0).
- **The answer** is stored in `quiz_responses.json_response` as
  `{"x": …, "y": …}`, in both modes. A pin outside
  `[0, mapWidth] × [0, mapHeight]` is rejected. On `MAP_PIN` only
  `mapPosition` is accepted; other answer fields (including `jsonResponse`)
  are rejected, so the stored JSON is always a validated point.
- **Individual answers.** Each participant pins and scores on their own. A
  team's score moves because team leaderboards average member points.
- **No betting.** `MAP_PIN` always has points, and points and betting are
  exclusive. The DB enforces `betting_enabled = false`.

## Storage

Migration 00119 adds nullable columns to `quiz_questions`:

| Column | Used by |
|---|---|
| `map_scoring_mode` (`POINT_RADIUS` / `POLYGON`) | both |
| `map_width`, `map_height` | both |
| `map_target_x`, `map_target_y` | `POINT_RADIUS` |
| `map_full_points_radius`, `map_zero_points_radius` | `POINT_RADIUS` |
| `map_polygon` JSONB `[{"x","y"}, …]` | `POLYGON` |

One CHECK (`quiz_questions_map_pin_check`) enforces that each row has exactly
the configuration it needs: every map column NULL for other question types;
mode and a positive size for `MAP_PIN`; target in bounds and
`0 <= full < zero` with no polygon for `POINT_RADIUS`; a 3–100 element array
with no target or radii for `POLYGON`. Geometric validity (simple polygon,
non-zero area) is enforced in Go — a CHECK cannot express it.

GraphQL takes the whole config as one `mapPin: MapPinConfigInput` on create
and update. On update it replaces all map columns at once (omitted = leave
alone), so switching modes clears the other mode's columns. Fields that belong
to the other mode are rejected, not ignored.

## Scoring

The pure functions live in `internal/services/quizgrading/mappin.go`.

**POINT_RADIUS**

```
d = hypot(x − tx, y − ty)
d <= full → points
d >= zero → 0
otherwise → round(points × (zero − d) / (zero − full))
```

`math.Round` because `points_earned` is `INT`. `full = 0` gives plain linear
falloff.

Rounding blurs both radii slightly. A pin just outside `fullPointsRadius`
can still round up to the maximum (`points = 100`, `full = 40`,
`zero = 200`: `d = 40.5` gives `round(99.69) = 100`), and a pin just inside
`zeroPointsRadius` can round down to 0. The radii mark where the exact
formula reaches max and 0. The awarded integer only matches them to within
half a point.

**POLYGON**

Max points when the pin is inside the polygon **or on its boundary** (an edge
or a vertex), otherwise 0. No partial points.

The test checks the boundary first — a pin within `Epsilon` of any edge is
inside — then uses the crossing-number rule with half-open edges
(`(yi > y) != (yj > y)`), which handles a ray passing exactly through a vertex
and works for concave polygons.

**Precision.** One constant, `Epsilon = 1e-6` px, is used both for the
boundary test and for polygon validation. It absorbs floating-point noise from
the frontend's zoom transform and is far below one pixel.

**Where points are written.** `submitQuizAnswer`, `updateQuizAnswer` and the
M2M `createQuizSubmission` all go through one helper, `scoreQuizResponse`.
`MAP_PIN` is handled before the `IsGradable` short-circuit — without that, it
would silently score 0. `updateQuizAnswer` always writes the value, including
0, because `UpdateQuizResponse` keeps the old value on NULL (`COALESCE`).
`recordBetResult(s)` rejects `MAP_PIN` responses, so nothing else can
overwrite their points.

## Polygon validation

Applied on create and update, in this order:

1. One closing vertex equal to the first is stripped (the polygon is stored
   open).
2. 3 to 100 vertices.
3. All coordinates finite.
4. All vertices inside `[0, mapWidth] × [0, mapHeight]`.
5. No duplicate vertices, adjacent or not.
6. Shoelace area greater than `Epsilon`.
7. Not self-intersecting: non-adjacent edges may not touch at all, and
   adjacent edges may not overlap (a backtracking spike). The first and last
   edges count as adjacent.

Clockwise and counter-clockwise are both accepted. Concave polygons are fine,
and so are collinear vertices in the middle of an edge.

## `is_correct` is always NULL

The point-radius task is "closer is better", with no strict right or wrong,
and both modes are scored by points, not a verdict. `MAP_PIN` is not gradable
(`quizgrading.IsGradable` is false), so `is_correct` is NULL, as for
FREE_TEXT, NUMBER and JSON. That is supported everywhere:

- `quiz_responses.is_correct` is nullable and `QuizResponse.isCorrect` is
  nullable in GraphQL.
- Bet settlement skips ungraded responses and only covers
  PREDEFINED/ORDERING.
- `correctCount` only exists on the predefined/ordering results types.

Frontend: review mode must not show a NULL verdict as "wrong".

## Visibility

Hidden values are removed by the backend, not just by the UI, so a user cannot
probe for the answer by moving their pin while the session is OPEN.

| Field | Visible to |
|---|---|
| `MapPinResponse.submittedPosition` | The respondent and project managers (existing submission access checks) |
| `MapPinResponse.pointsEarned`, `distance` | Hidden from everyone while the submission's session is OPEN; visible from LOCKED (same rule as `isCorrect`) |
| `MapPinQuestion.correctPosition`, `correctArea` | Managers always; others only when their most recent accessible session is FINISHED |
| `scoringMode`, radii, map size | Everyone |

`distance` is only set in `POINT_RADIUS` mode.

**Stricter than other question types.** `canRevealQuestionAnswers` reveals
correct answers to a user with *no* accessible session, and `quiz.questions`
has no access check. For `MAP_PIN` that would let a team whose session hasn't
been created yet read the target, so the geometry stays hidden when the user
has no session. PREDEFINED/ORDERING have the same leak, as does
`pointsEarned` while OPEN on those types; both are separate issues.

**Running sessions.** LOCK each team's session as soon as the team has
played, and FINISH all sessions together at the end. Points are visible from
LOCKED, but the correct geometry only from FINISHED — finishing a session early
lets that team tell the others where the answer is.

## Changing a question after answers exist

Nothing regrades existing responses. To keep `points_earned` consistent, a
`MAP_PIN` question that already has responses rejects changes to `mapPin` and
`points`. To change it after a test run, clear the test answers first
(`deleteQuizSession` / `resetQuizSessionSubmission`).

The map image is not locked (a same-size colour fix is fine), but swapping it
for a different resolution invalidates the geometry — see the coordinates
note.

## Results

`MapPinQuestionResults` (manager-only, like all of `quizResults`):

- `question` — the `MapPinQuestion`, with its mode, radii and correct
  geometry (not duplicated on the results type)
- `responseCount`
- `fullPointsCount` — responses where `points_earned == question.points`,
  the same rule in both modes. For `POLYGON` that is exactly the pins inside
  or on the boundary. For `POINT_RADIUS` it also includes pins slightly
  beyond `fullPointsRadius` that rounded up to the maximum (see Scoring).
  It is a count of awarded points, not of distances.
- `averagePoints`
- `averageDistance` — `POINT_RADIUS` only; null for `POLYGON`, where distance
  means nothing
- `pins` — position, points and (`POINT_RADIUS` only) distance for every
  response

## Implementation steps

1. Docs (this note and the coordinates note)
2. `quizgrading`: scoring, point-in-polygon, validation + tests
3. Migration 00119 + `quiz_questions.sql`
4. GraphQL `MapPinQuestion`: create/update/read, rules, geometry gating
5. Answering: `mapPosition`, `MapPinResponse`, scoring helper, guards
6. `MapPinQuestionResults`
7. E2E tests for both modes

Frontend (zoomable map, admin target/polygon editor, `QuizChallenge.vue`
typename allowlists, review mode) is not part of these steps.
