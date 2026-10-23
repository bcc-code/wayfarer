package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// New creates a new slog.Logger based on the environment and log level
// In production, uses JSON format. In development, uses colored human-readable format.
// currentLevel is shared by every handler, including the clones WithAttrs and
// WithGroup return, so SetLevel reaches loggers that already exist.
var currentLevel = new(slog.LevelVar)

func New(environment string, level slog.Level) *slog.Logger {
	currentLevel.Set(level)
	return slog.New(localHandler(environment))
}

// NewWithExport is New plus asynchronous shipping of every record to
// exportCfg.URL. The returned Exporter must be shut down to flush its buffer.
func NewWithExport(environment string, level slog.Level, exportCfg ExportConfig) (*slog.Logger, *Exporter) {
	currentLevel.Set(level)
	local := localHandler(environment)
	exp := NewExporter(exportCfg, local)
	return slog.New(newMultiHandler(local, exp.Handler())), exp
}

func localHandler(environment string) slog.Handler {
	if environment == "production" {
		return newGCPHandler(currentLevel)
	}
	return newDevHandler(currentLevel)
}

// SetLevel changes the verbosity of every logger New has returned.
func SetLevel(level slog.Level) {
	currentLevel.Set(level)
}

// Level reports the verbosity currently in force.
func Level() slog.Level {
	return currentLevel.Level()
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
	level *slog.LevelVar
	attrs []slog.Attr // from WithAttrs, already prefixed with any open groups
	group string      // dotted prefix from WithGroup
}

func newDevHandler(level *slog.LevelVar) *devHandler {
	return &devHandler{level: level}
}

func (h *devHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
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
	attrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	attrs = append(attrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, h.prefixed(a))
		return true
	})
	if len(attrs) > 0 {
		if _, err = fmt.Fprint(os.Stdout, " ("); err != nil {
			return err
		}
		for i, a := range attrs {
			if i > 0 {
				if _, err = fmt.Fprint(os.Stdout, ", "); err != nil {
					return err
				}
			}
			if _, err = fmt.Fprintf(os.Stdout, "%s=%v", a.Key, a.Value); err != nil {
				return err
			}
		}
		if _, err = fmt.Fprint(os.Stdout, ")"); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(os.Stdout)
	return err
}

func (h *devHandler) prefixed(a slog.Attr) slog.Attr {
	if h.group != "" {
		a.Key = h.group + "." + a.Key
	}
	return a
}

func (h *devHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	for _, a := range attrs {
		newAttrs = append(newAttrs, h.prefixed(a))
	}
	return &devHandler{level: h.level, attrs: newAttrs, group: h.group}
}

func (h *devHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	group := name
	if h.group != "" {
		group = h.group + "." + name
	}
	return &devHandler{level: h.level, attrs: h.attrs, group: group}
}
