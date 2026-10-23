package main

import (
	"os"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/logger"
)

// logExportConfig maps the app config to the log exporter's settings. The
// static fields become VictoriaLogs stream fields (service, host, instance,
// env), so they stay low-cardinality.
func logExportConfig(cfg *config.Config) logger.ExportConfig {
	fields := map[string]string{
		"service": cfg.OTEL.ServiceName,
		"env":     cfg.Server.Environment,
	}
	if host, err := os.Hostname(); err == nil {
		fields["host"] = host
	}
	if cfg.Log.ExportInstance != "" {
		fields["instance"] = cfg.Log.ExportInstance
	}
	return logger.ExportConfig{
		URL:           cfg.Log.ExportURL,
		Level:         logger.ParseLevel(cfg.Log.ExportLevel),
		BufferSize:    cfg.Log.ExportBufferSize,
		BatchSize:     cfg.Log.ExportBatchSize,
		FlushInterval: cfg.Log.ExportFlushInterval,
		Fields:        fields,
	}
}
