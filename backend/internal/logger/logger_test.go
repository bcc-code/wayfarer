package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGCPSeverity(t *testing.T) {
	tests := []struct {
		level    slog.Level
		expected string
	}{
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARNING"},
		{slog.LevelError, "ERROR"},
		{slog.LevelDebug - 4, "DEBUG"}, // Below debug
		{slog.LevelError + 4, "ERROR"}, // Above error (still ERROR)
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := gcpSeverity(tt.level)
			if result != tt.expected {
				t.Errorf("gcpSeverity(%v) = %q, want %q", tt.level, result, tt.expected)
			}
		})
	}
}

func TestGCPHandlerEnabled(t *testing.T) {
	handler := newGCPHandler(slog.LevelWarn)

	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Expected DEBUG to be disabled when level is WARN")
	}
	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Expected INFO to be disabled when level is WARN")
	}
	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("Expected WARN to be enabled when level is WARN")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Error("Expected ERROR to be enabled when level is WARN")
	}
}

func TestGCPHandlerHandle(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler := newGCPHandler(slog.LevelInfo)
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 123456000, time.UTC)

	record := slog.NewRecord(testTime, slog.LevelError, "test message", 0)
	record.AddAttrs(slog.String("key", "value"), slog.Int("count", 42))

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := strings.TrimSpace(buf.String())

	// Parse JSON
	var entry map[string]any
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify fields
	if entry["severity"] != "ERROR" {
		t.Errorf("severity = %v, want ERROR", entry["severity"])
	}
	if entry["message"] != "test message" {
		t.Errorf("message = %v, want 'test message'", entry["message"])
	}
	if entry["time"] != "2024-01-15T10:30:45.123456Z" {
		t.Errorf("time = %v, want '2024-01-15T10:30:45.123456Z'", entry["time"])
	}
	if entry["key"] != "value" {
		t.Errorf("key = %v, want 'value'", entry["key"])
	}
	if entry["count"] != float64(42) { // JSON numbers are float64
		t.Errorf("count = %v, want 42", entry["count"])
	}
}

func TestGCPHandlerWithAttrs(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler := newGCPHandler(slog.LevelInfo)
	handlerWithAttrs := handler.WithAttrs([]slog.Attr{
		slog.String("service", "test-service"),
	})

	testTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	record := slog.NewRecord(testTime, slog.LevelInfo, "with attrs", 0)
	record.AddAttrs(slog.String("extra", "data"))

	err := handlerWithAttrs.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := strings.TrimSpace(buf.String())

	// Parse JSON
	var entry map[string]any
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify pre-existing attr is present
	if entry["service"] != "test-service" {
		t.Errorf("service = %v, want 'test-service'", entry["service"])
	}
	if entry["extra"] != "data" {
		t.Errorf("extra = %v, want 'data'", entry["extra"])
	}
}

func TestGCPHandlerWithGroup(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler := newGCPHandler(slog.LevelInfo)
	handlerWithGroup := handler.WithGroup("request")
	handlerWithGroupAndAttrs := handlerWithGroup.WithAttrs([]slog.Attr{
		slog.String("method", "GET"),
	})

	testTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	record := slog.NewRecord(testTime, slog.LevelInfo, "with group", 0)
	record.AddAttrs(slog.String("path", "/api/test"))

	err := handlerWithGroupAndAttrs.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := strings.TrimSpace(buf.String())

	// Parse JSON
	var entry map[string]any
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify group nesting
	request, ok := entry["request"].(map[string]any)
	if !ok {
		t.Fatalf("request group not found or not a map: %v", entry)
	}
	if request["method"] != "GET" {
		t.Errorf("request.method = %v, want 'GET'", request["method"])
	}
	if request["path"] != "/api/test" {
		t.Errorf("request.path = %v, want '/api/test'", request["path"])
	}
}

func TestGCPHandlerWithEmptyGroup(t *testing.T) {
	handler := newGCPHandler(slog.LevelInfo)
	handlerWithEmptyGroup := handler.WithGroup("")

	// Should return the same handler
	if handlerWithEmptyGroup != handler {
		t.Error("WithGroup(\"\") should return the same handler")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo}, // Default
		{"", slog.LevelInfo},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	// Test production logger
	prodLogger := New("production", slog.LevelInfo)
	if prodLogger == nil {
		t.Error("New(production) returned nil")
	}

	// Test development logger
	devLogger := New("development", slog.LevelInfo)
	if devLogger == nil {
		t.Error("New(development) returned nil")
	}
}
