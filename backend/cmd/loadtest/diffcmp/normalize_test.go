package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decode(t *testing.T, s string) any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader([]byte(s)))
	dec.UseNumber()
	var v any
	require.NoError(t, dec.Decode(&v))
	return v
}

func TestSeededID(t *testing.T) {
	// Must reproduce gendata.sql's mkid(prefix, n) exactly.
	assert.Equal(t, "PR00000000000000000000000001", seededID("PR", 1))
	assert.Equal(t, "TM0000000000000000000000000A", seededID("TM", 10))
	assert.Equal(t, "US0000000000000000000000012C", seededID("US", 300))
	assert.Len(t, seededID("US", 13162), 28)
}

func TestUlidTimestamp(t *testing.T) {
	// Seeded mkid IDs decode to ~1970 and must pass through untouched.
	ts, ok := ulidTimestamp("00000000000000000000000001")
	require.True(t, ok)
	assert.True(t, ts.Before(ulidCutoff))

	// A real ULID minted now decodes to a recent timestamp.
	// 01K… ULIDs (2025+) start with a high timestamp component.
	ts, ok = ulidTimestamp("01K00000000000000000000000")
	require.True(t, ok)
	assert.False(t, ts.Before(ulidCutoff))
}

func TestNormalizeRuntimeULIDs(t *testing.T) {
	// Two references to the same runtime ULID must map to the same
	// placeholder; seeded and fixture IDs must survive unchanged.
	in := decode(t, `{
		"submission": {"id": "QS01K3XW5E8ZJ4M9N2P6Q7R8S9XY", "again": "QS01K3XW5E8ZJ4M9N2P6Q7R8S9XY"},
		"seeded": "US0000000000000000000000012C",
		"fixture": "QZ01LOADTESTFREETEXT00000000"
	}`)
	out := Normalize(in, &NormOpts{}).(map[string]any)
	sub := out["submission"].(map[string]any)
	assert.Equal(t, sub["id"], sub["again"])
	assert.Contains(t, sub["id"], "ULID:QS:")
	assert.Equal(t, "US0000000000000000000000012C", out["seeded"])
	assert.Equal(t, "QZ01LOADTESTFREETEXT00000000", out["fixture"])
}

func TestNormalizePerSideConsistency(t *testing.T) {
	// Different runtime ULIDs on the two sides normalize to the same
	// placeholder sequence, so structurally identical responses diff clean.
	a := decode(t, `{"id": "QS01K3XW5E8ZJ4M9N2P6Q7R8S9XY", "score": 3}`)
	b := decode(t, `{"id": "QS01K3XW5FAAAAAAAAAAAAAAAAAA", "score": 3}`)
	na := Normalize(a, &NormOpts{})
	nb := Normalize(b, &NormOpts{})
	assert.Empty(t, Diff(na, nb))
}

func TestNormalizeTimestamps(t *testing.T) {
	runStart := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	in := decode(t, `{
		"recent": "2026-08-25T12:03:04Z",
		"old": "2026-03-01T00:00:00Z",
		"notATime": "hello world"
	}`)
	out := Normalize(in, &NormOpts{RunStart: runStart}).(map[string]any)
	assert.Equal(t, "TS:recent", out["recent"])
	assert.Equal(t, "2026-03-01T00:00:00Z", out["old"])
	assert.Equal(t, "hello world", out["notATime"])
}

func TestNormalizeRedact(t *testing.T) {
	in := decode(t, `{"firebaseToken": {"token": "eyJhbGc...", "expiresIn": 3600}}`)
	out := Normalize(in, &NormOpts{RedactPaths: []string{"firebaseToken.token"}}).(map[string]any)
	tok := out["firebaseToken"].(map[string]any)
	assert.Equal(t, "REDACTED:nonempty", tok["token"])
	assert.Equal(t, json.Number("3600"), tok["expiresIn"])
}

func TestNormalizeSortPaths(t *testing.T) {
	a := decode(t, `{"items": [{"n": "b"}, {"n": "a"}]}`)
	b := decode(t, `{"items": [{"n": "a"}, {"n": "b"}]}`)
	optsA := &NormOpts{SortPaths: []string{"items"}}
	optsB := &NormOpts{SortPaths: []string{"items"}}
	assert.Empty(t, Diff(Normalize(a, optsA), Normalize(b, optsB)))
	// Without the opt-in, order is correctness and must diff.
	assert.NotEmpty(t, Diff(Normalize(a, &NormOpts{}), Normalize(b, &NormOpts{})))
}

func TestDiffReportsPaths(t *testing.T) {
	a := decode(t, `{"x": {"y": [1, 2, 3]}, "z": "same"}`)
	b := decode(t, `{"x": {"y": [1, 9, 3]}, "z": "same"}`)
	d := Diff(a, b)
	require.Len(t, d, 1)
	assert.Contains(t, d[0], "$.x.y[1]")
}

func TestDiffLengthMismatch(t *testing.T) {
	a := decode(t, `{"edges": [1, 2, 3, 4, 5]}`)
	b := decode(t, `{"edges": [1, 2, 3]}`)
	d := Diff(a, b)
	require.NotEmpty(t, d)
	assert.Contains(t, d[0], "array length 5 vs 3")
}

func TestDiffMissingKeys(t *testing.T) {
	a := decode(t, `{"present": 1, "onlyA": 2}`)
	b := decode(t, `{"present": 1, "onlyB": 3}`)
	d := Diff(a, b)
	assert.Len(t, d, 2)
}

func TestApplyVars(t *testing.T) {
	captured := map[string]string{"subId": "QS01ABC", "item1": "QI01X"}
	in := map[string]any{
		"submissionId": "{{subId}}",
		"input": map[string]any{
			"submittedOrder": []any{"{{item1}}", "literal"},
			"count":          5,
		},
	}
	out, err := applyVars(in, captured)
	require.NoError(t, err)
	m := out.(map[string]any)
	assert.Equal(t, "QS01ABC", m["submissionId"])
	inner := m["input"].(map[string]any)
	assert.Equal(t, []any{"QI01X", "literal"}, inner["submittedOrder"])
	assert.Equal(t, 5, inner["count"])

	_, err = applyVars(map[string]any{"x": "{{missing}}"}, captured)
	assert.Error(t, err)
}

func TestCapturePath(t *testing.T) {
	data := decode(t, `{"startQuizSession": {"id": "QS01A", "orderedQuestions": [{"id": "QQ01"}, {"id": "QQ02"}]}}`)
	v, err := capturePath(data, "startQuizSession.id")
	require.NoError(t, err)
	assert.Equal(t, "QS01A", v)

	v, err = capturePath(data, "startQuizSession.orderedQuestions.1.id")
	require.NoError(t, err)
	assert.Equal(t, "QQ02", v)

	_, err = capturePath(data, "startQuizSession.missing")
	assert.Error(t, err)
}

func TestScenariosBuildAndResolve(t *testing.T) {
	// Every referenced query must exist; capture templates must be
	// resolvable given the captures declared by earlier steps.
	for _, sc := range buildScenarios() {
		declared := map[string]string{}
		for _, step := range sc.Steps {
			if step.Query != "" {
				_, ok := queries[step.Query]
				assert.True(t, ok, "%s/%s references unknown query %q", sc.Name, step.Name, step.Query)
			} else {
				assert.NotEmpty(t, step.Path, "%s/%s has neither query nor path", sc.Name, step.Name)
			}
			// Simulate captures so templates in later steps resolve.
			if step.Vars != nil {
				_, err := applyVars(map[string]any(step.Vars), declared)
				assert.NoError(t, err, "%s/%s has unresolved vars", sc.Name, step.Name)
			}
			if step.Body != nil {
				_, err := applyVars(map[string]any(step.Body), declared)
				assert.NoError(t, err, "%s/%s has unresolved body vars", sc.Name, step.Name)
			}
			for name := range step.Capture {
				declared[name] = "CAPTURED"
			}
		}
	}
}

func TestRunCheckMeIsCaller(t *testing.T) {
	resp := &gqlResponse{Data: decode(t, `{"myCurrentProject": {"leaderboard": {"me": {"id": "US01"}}}}`)}
	assert.NoError(t, runCheck("meIsCaller", resp, "US01", nil))
	err := runCheck("meIsCaller", resp, "US02", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "foreign user data")
}

func TestRunCheckEdgeCount(t *testing.T) {
	resp := &gqlResponse{Data: decode(t, `{"myCurrentProject": {"leaderboard": {"edges": [{}, {}, {}]}}}`)}
	assert.NoError(t, runCheck("edgeCountIsFirst", resp, "", map[string]any{"first": 3}))
	assert.Error(t, runCheck("edgeCountIsFirst", resp, "", map[string]any{"first": 50}))
}
