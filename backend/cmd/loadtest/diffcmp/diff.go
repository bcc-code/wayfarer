package main

import (
	"encoding/json"
	"fmt"
)

// Diff structurally compares two normalized JSON values and returns one
// human-readable line per mismatched path. Empty result means equal.
func Diff(a, b any) []string {
	var out []string
	diffValue(a, b, "$", &out)
	return out
}

func diffValue(a, b any, path string, out *[]string) {
	if len(*out) >= 200 { // cap runaway diffs
		return
	}
	switch ta := a.(type) {
	case map[string]any:
		tb, ok := b.(map[string]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: type mismatch: object vs %s", path, typeName(b)))
			return
		}
		for k, va := range ta {
			vb, present := tb[k]
			if !present {
				*out = append(*out, fmt.Sprintf("%s.%s: missing on B", path, k))
				continue
			}
			diffValue(va, vb, path+"."+k, out)
		}
		for k := range tb {
			if _, present := ta[k]; !present {
				*out = append(*out, fmt.Sprintf("%s.%s: missing on A", path, k))
			}
		}
	case []any:
		tb, ok := b.([]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: type mismatch: array vs %s", path, typeName(b)))
			return
		}
		if len(ta) != len(tb) {
			*out = append(*out, fmt.Sprintf("%s: array length %d vs %d", path, len(ta), len(tb)))
		}
		for i := range min(len(ta), len(tb)) {
			diffValue(ta[i], tb[i], fmt.Sprintf("%s[%d]", path, i), out)
		}
	case json.Number:
		tb, ok := b.(json.Number)
		if !ok || ta.String() != tb.String() {
			*out = append(*out, fmt.Sprintf("%s: %v vs %v", path, a, b))
		}
	default:
		if a != b {
			*out = append(*out, fmt.Sprintf("%s: %v vs %v", path, render(a), render(b)))
		}
	}
}

func typeName(v any) string {
	switch v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func render(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	s := string(b)
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return s
}
