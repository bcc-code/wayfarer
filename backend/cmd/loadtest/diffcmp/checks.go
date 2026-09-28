package main

import (
	"encoding/json"
	"fmt"
)

// runCheck evaluates a named per-side assertion against one side's response.
// It returns an error describing the violation, or nil when the check passes.
// Checks run on BOTH sides independently: on the pinned probes the branch (B)
// is expected to fail some of these — the failure text becomes the evidence.
func runCheck(name string, resp *gqlResponse, callerUserID string, vars map[string]any) error {
	switch name {
	case "":
		return nil
	case "statusOnly":
		// Comparison-mode marker, not an assertion.
		return nil
	case "meIsCaller":
		v, err := capturePath(resp.Data, "myCurrentProject.leaderboard.me.id")
		if err != nil {
			return fmt.Errorf("meIsCaller: %w", err)
		}
		if v != callerUserID {
			return fmt.Errorf("meIsCaller: leaderboard.me.id = %s, caller = %s (foreign user data served)", v, callerUserID)
		}
		return nil
	case "edgeCountIsFirst":
		want, err := intVar(vars, "first")
		if err != nil {
			return fmt.Errorf("edgeCountIsFirst: %w", err)
		}
		edges, err := arrayPath(resp.Data, "myCurrentProject.leaderboard.edges")
		if err != nil {
			return fmt.Errorf("edgeCountIsFirst: %w", err)
		}
		if len(edges) != want {
			return fmt.Errorf("edgeCountIsFirst: got %d edges, requested first=%d", len(edges), want)
		}
		return nil
	default:
		return fmt.Errorf("unknown check %q", name)
	}
}

func intVar(vars map[string]any, key string) (int, error) {
	v, ok := vars[key]
	if !ok {
		return 0, fmt.Errorf("variable %q not set", key)
	}
	switch t := v.(type) {
	case int:
		return t, nil
	case json.Number:
		i, err := t.Int64()
		return int(i), err
	default:
		return 0, fmt.Errorf("variable %q is %T, want int", key, v)
	}
}

func arrayPath(data any, path string) ([]any, error) {
	cur := data
	var err error
	cur, err = anyPath(cur, path)
	if err != nil {
		return nil, err
	}
	arr, ok := cur.([]any)
	if !ok {
		return nil, fmt.Errorf("path %q: value is %T, want array", path, cur)
	}
	return arr, nil
}

func anyPath(data any, path string) (any, error) {
	cur := data
	for _, seg := range splitPath(path) {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("path %q: cannot descend into %T at %q", path, cur, seg)
		}
		cur, ok = m[seg]
		if !ok {
			return nil, fmt.Errorf("path %q: key %q not found", path, seg)
		}
	}
	return cur, nil
}
