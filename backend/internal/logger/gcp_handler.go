package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// gcpHandler is a custom slog handler for Google Cloud Logging.
// It outputs JSON with fields that GCP expects: "severity", "message", "time".
type gcpHandler struct {
	level  slog.Level
	attrs  []slog.Attr
	groups []string
	mu     sync.Mutex
}

func newGCPHandler(level slog.Level) *gcpHandler {
	return &gcpHandler{level: level}
}

// gcpSeverity maps slog levels to Google Cloud Logging severity strings.
func gcpSeverity(level slog.Level) string {
	switch {
	case level < slog.LevelInfo:
		return "DEBUG"
	case level < slog.LevelWarn:
		return "INFO"
	case level < slog.LevelError:
		return "WARNING"
	default:
		return "ERROR"
	}
}

func (h *gcpHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *gcpHandler) Handle(_ context.Context, r slog.Record) error {
	// Build the log entry as a map
	entry := make(map[string]any)
	entry["severity"] = gcpSeverity(r.Level)
	entry["message"] = r.Message
	entry["time"] = r.Time.UTC().Format("2006-01-02T15:04:05.000000Z")

	// Add pre-existing attributes (from WithAttrs)
	for _, attr := range h.attrs {
		addAttrToMap(entry, h.groups, attr)
	}

	// Add record attributes
	r.Attrs(func(a slog.Attr) bool {
		addAttrToMap(entry, h.groups, a)
		return true
	})

	// Marshal and write
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = fmt.Fprintln(os.Stdout, string(data))
	return err
}

func (h *gcpHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	newAttrs = append(newAttrs, attrs...)
	return &gcpHandler{
		level:  h.level,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *gcpHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups), len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups = append(newGroups, name)
	return &gcpHandler{
		level:  h.level,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// addAttrToMap adds a slog.Attr to a map, handling groups by nesting.
func addAttrToMap(m map[string]any, groups []string, attr slog.Attr) {
	// Resolve the attribute value
	attr.Value = attr.Value.Resolve()

	// Skip empty attributes
	if attr.Equal(slog.Attr{}) {
		return
	}

	// Navigate to the correct nested map based on groups
	target := m
	for _, g := range groups {
		if existing, ok := target[g]; ok {
			if nested, ok := existing.(map[string]any); ok {
				target = nested
			} else {
				// Group key already exists with non-map value, overwrite
				nested := make(map[string]any)
				target[g] = nested
				target = nested
			}
		} else {
			nested := make(map[string]any)
			target[g] = nested
			target = nested
		}
	}

	// Handle the attribute based on its kind
	switch attr.Value.Kind() {
	case slog.KindGroup:
		groupAttrs := attr.Value.Group()
		if len(groupAttrs) == 0 {
			return
		}
		if attr.Key == "" {
			// Inline group - add attrs directly to target
			for _, ga := range groupAttrs {
				addAttrToMap(target, nil, ga)
			}
		} else {
			// Named group - create nested map
			nested := make(map[string]any)
			for _, ga := range groupAttrs {
				addAttrToMap(nested, nil, ga)
			}
			target[attr.Key] = nested
		}
	default:
		target[attr.Key] = attr.Value.Any()
	}
}
