package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// New creates a new slog.Logger based on the environment and log level
// In production, uses JSON format. In development, uses colored human-readable format.
func New(environment string, level slog.Level) *slog.Logger {
	var handler slog.Handler

	if environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = newDevHandler(level)
	}

	return slog.New(handler)
}

// ParseLevel converts a string log level to slog.Level
func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// devHandler is a custom slog handler for development with format: [Time] SEVERITY Msg (key=value)
// Levels are colored: DEBUG (no color), INFO (blue), WARN (yellow), ERROR (red)
type devHandler struct {
	level slog.Level
}

func newDevHandler(level slog.Level) *devHandler {
	return &devHandler{level: level}
}

func (h *devHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *devHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.Format("15:04:05")

	// Color codes
	var levelColor string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = "" // No color
	case slog.LevelInfo:
		levelColor = "\033[34m" // Blue
	case slog.LevelWarn:
		levelColor = "\033[33m" // Yellow
	case slog.LevelError:
		levelColor = "\033[31m" // Red
	default:
		levelColor = ""
	}
	resetColor := "\033[0m"

	// Build the log line with colored level
	var err error
	if levelColor != "" {
		_, err = fmt.Fprintf(os.Stdout, "[%s] %s%s%s %s", timeStr, levelColor, r.Level, resetColor, r.Message)
	} else {
		_, err = fmt.Fprintf(os.Stdout, "[%s] %s %s", timeStr, r.Level, r.Message)
	}
	if err != nil {
		return err
	}

	// Add attributes
	if r.NumAttrs() > 0 {
		if _, err = fmt.Fprint(os.Stdout, " ("); err != nil {
			return err
		}
		first := true
		r.Attrs(func(a slog.Attr) bool {
			if !first {
				if _, err = fmt.Fprint(os.Stdout, ", "); err != nil {
					return false
				}
			}
			first = false
			if _, err = fmt.Fprintf(os.Stdout, "%s=%v", a.Key, a.Value); err != nil {
				return false
			}
			return true
		})
		if err != nil {
			return err
		}
		if _, err = fmt.Fprint(os.Stdout, ")"); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(os.Stdout)
	return err
}

func (h *devHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, return the same handler
	// In a full implementation, you'd store these attrs and include them in Handle
	return h
}

func (h *devHandler) WithGroup(name string) slog.Handler {
	// For simplicity, return the same handler
	// In a full implementation, you'd track the group name
	return h
}
