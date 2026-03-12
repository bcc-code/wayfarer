package pubsub

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

// LogEntry represents a single log entry captured during job processing
type LogEntry struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

// LogFlusher is called to persist buffered logs
type LogFlusher func(ctx context.Context, logs []byte) error

// JobLogHandler is a slog.Handler that captures logs for a bulk job and batches DB writes
type JobLogHandler struct {
	jobID   string
	logs    []LogEntry
	mu      sync.Mutex
	wrapped slog.Handler
	attrs   []slog.Attr
	groups  []string
}

// NewJobLogHandler creates a handler that captures logs and passes them to a wrapped handler
func NewJobLogHandler(jobID string, wrapped slog.Handler) *JobLogHandler {
	return &JobLogHandler{
		jobID:   jobID,
		logs:    make([]LogEntry, 0),
		wrapped: wrapped,
	}
}

// Enabled returns true if the level is enabled
func (h *JobLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.wrapped.Enabled(ctx, level)
}

// Handle captures the log record and passes it to the wrapped handler
func (h *JobLogHandler) Handle(ctx context.Context, r slog.Record) error {
	entry := LogEntry{
		Time:    r.Time,
		Level:   r.Level.String(),
		Message: r.Message,
		Attrs:   make(map[string]any),
	}

	// Add pre-set attrs
	for _, a := range h.attrs {
		entry.Attrs[a.Key] = a.Value.Any()
	}

	// Add record attrs
	r.Attrs(func(a slog.Attr) bool {
		entry.Attrs[a.Key] = a.Value.Any()
		return true
	})

	h.mu.Lock()
	h.logs = append(h.logs, entry)
	h.mu.Unlock()

	return h.wrapped.Handle(ctx, r)
}

// WithAttrs returns a new handler with the given attributes
func (h *JobLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &JobLogHandler{
		jobID:   h.jobID,
		logs:    h.logs, // Share the same log slice
		mu:      sync.Mutex{},
		wrapped: h.wrapped.WithAttrs(attrs),
		attrs:   append(h.attrs, attrs...),
		groups:  h.groups,
	}
	return newHandler
}

// WithGroup returns a new handler with the given group
func (h *JobLogHandler) WithGroup(name string) slog.Handler {
	return &JobLogHandler{
		jobID:   h.jobID,
		logs:    h.logs,
		mu:      sync.Mutex{},
		wrapped: h.wrapped.WithGroup(name),
		attrs:   h.attrs,
		groups:  append(h.groups, name),
	}
}

// GetLogs returns the captured logs as JSON bytes
func (h *JobLogHandler) GetLogs() ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.logs) == 0 {
		return nil, nil
	}

	return json.Marshal(h.logs)
}

// LogCount returns the number of captured logs
func (h *JobLogHandler) LogCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.logs)
}

// Clear removes all captured logs (call after flushing)
func (h *JobLogHandler) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logs = make([]LogEntry, 0)
}

// FlushAndClear returns the logs as JSON and clears the buffer
func (h *JobLogHandler) FlushAndClear() ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.logs) == 0 {
		return nil, nil
	}

	data, err := json.Marshal(h.logs)
	if err != nil {
		return nil, err
	}

	h.logs = make([]LogEntry, 0)
	return data, nil
}
