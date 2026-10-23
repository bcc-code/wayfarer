package main

import (
	"log/slog"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLogExportConfig(t *testing.T) {
	cfg := &config.Config{}
	cfg.OTEL.ServiceName = "wayfarer-backend"
	cfg.Server.Environment = "production"
	cfg.Log.ExportURL = "http://127.0.0.1:9880/"
	cfg.Log.ExportLevel = "warn"
	cfg.Log.ExportBufferSize = 5
	cfg.Log.ExportBatchSize = 2
	cfg.Log.ExportFlushInterval = time.Second
	cfg.Log.ExportInstance = "blue"

	got := logExportConfig(cfg)
	assert.Equal(t, "http://127.0.0.1:9880/", got.URL)
	assert.Equal(t, slog.LevelWarn, got.Level)
	assert.Equal(t, 5, got.BufferSize)
	assert.Equal(t, 2, got.BatchSize)
	assert.Equal(t, time.Second, got.FlushInterval)
	assert.Equal(t, "wayfarer-backend", got.Fields["service"])
	assert.Equal(t, "production", got.Fields["env"])
	assert.Equal(t, "blue", got.Fields["instance"])
	assert.NotEmpty(t, got.Fields["host"])

	cfg.Log.ExportInstance = ""
	_, ok := logExportConfig(cfg).Fields["instance"]
	assert.False(t, ok, "no empty instance field")
}
