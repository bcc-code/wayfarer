package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// NormOpts controls response normalization before diffing.
type NormOpts struct {
	// RunStart anchors the "recent" window: RFC3339 strings at or after
	// RunStart-15m are considered runtime-generated and collapse to a
	// placeholder. Older timestamps must match exactly (both DBs are clones,
	// so seeded values are byte-identical).
	RunStart time.Time
	// RedactPaths are dot-path suffixes (arrays as *) whose string values are
	// replaced with REDACTED:<empty|nonempty> — used for values that are
	// legitimately different per side (e.g. signed Firebase tokens).
	RedactPaths []string
	// SortPaths are dot-path suffixes of arrays with documented
	// nondeterministic order; the array is sorted by its canonical JSON
	// encoding before diffing. Off by default: element order is correctness.
	SortPaths []string
	// ulids maps runtime ULIDs to stable placeholders, one map per side, so
	// cross-references inside one response stay equal after scrubbing.
	ulids map[string]string
}

// crockford is the ULID base32 alphabet (no I, L, O, U).
var ulidRe = regexp.MustCompile(`^[A-Z]{2}[0-9ABCDEFGHJKMNPQRSTVWXYZ]{26}$`)

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ulidCutoff separates seeded IDs (mkid hex-padding decodes to ~1970) from
// runtime-generated ULIDs (timestamp = now). Anything at or after this is
// runtime-generated.
var ulidCutoff = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

// ulidTimestamp decodes the 48-bit millisecond timestamp from the first 10
// characters of a 26-char ULID suffix. Returns false if any character is
// outside the Crockford alphabet.
func ulidTimestamp(suffix string) (time.Time, bool) {
	var ms uint64
	for i := range 10 {
		idx := strings.IndexByte(crockford, suffix[i])
		if idx < 0 {
			return time.Time{}, false
		}
		ms = ms<<5 | uint64(idx)
	}
	return time.UnixMilli(int64(ms)).UTC(), true
}

// Normalize rewrites a decoded JSON value (json.Number for numbers) into a
// canonical form: runtime ULIDs become stable per-side placeholders, recent
// timestamps collapse, redacted paths lose their values, and opted-in array
// paths are sorted. The input is not modified.
func Normalize(v any, opts *NormOpts) any {
	if opts.ulids == nil {
		opts.ulids = map[string]string{}
	}
	return normalize(v, "", opts)
}

func normalize(v any, path string, opts *NormOpts) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		// Deterministic traversal so per-side ULID placeholder numbering is
		// reproducible regardless of Go map iteration order.
		sort.Strings(keys)
		for _, k := range keys {
			out[k] = normalize(t[k], joinPath(path, k), opts)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = normalize(e, joinPath(path, "*"), opts)
		}
		if pathMatches(path, opts.SortPaths) {
			sort.Slice(out, func(i, j int) bool {
				return canonicalJSON(out[i]) < canonicalJSON(out[j])
			})
		}
		return out
	case string:
		return normalizeString(t, path, opts)
	default:
		return v
	}
}

func normalizeString(s, path string, opts *NormOpts) string {
	if pathMatches(path, opts.RedactPaths) {
		if s == "" {
			return "REDACTED:empty"
		}
		return "REDACTED:nonempty"
	}
	if ulidRe.MatchString(s) {
		if ts, ok := ulidTimestamp(s[2:]); ok && !ts.Before(ulidCutoff) {
			ph, seen := opts.ulids[s]
			if !seen {
				ph = fmt.Sprintf("ULID:%s:%d", s[:2], len(opts.ulids)+1)
				opts.ulids[s] = ph
			}
			return ph
		}
		return s
	}
	if ts, err := time.Parse(time.RFC3339Nano, s); err == nil {
		if !opts.RunStart.IsZero() && !ts.Before(opts.RunStart.Add(-15*time.Minute)) {
			return "TS:recent"
		}
	}
	return s
}

func joinPath(base, seg string) string {
	if base == "" {
		return seg
	}
	return base + "." + seg
}

// pathMatches reports whether path equals or ends with any of the suffixes.
func pathMatches(path string, suffixes []string) bool {
	for _, s := range suffixes {
		if path == s || strings.HasSuffix(path, "."+s) {
			return true
		}
	}
	return false
}

// canonicalJSON renders a value with sorted keys for stable comparison.
func canonicalJSON(v any) string {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			kb, _ := json.Marshal(k)
			b.Write(kb)
			b.WriteByte(':')
			b.WriteString(canonicalJSON(t[k]))
		}
		b.WriteByte('}')
		return b.String()
	case []any:
		var b strings.Builder
		b.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(canonicalJSON(e))
		}
		b.WriteByte(']')
		return b.String()
	case json.Number:
		return t.String()
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
