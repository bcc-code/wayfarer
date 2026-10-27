package logger

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// ExportConfig configures shipping log records to a VictoriaLogs-compatible
// NDJSON endpoint (VictoriaLogs /insert/jsonline, or a local Vector
// http_server source that forwards there).
type ExportConfig struct {
	URL           string            // POST target for gzip NDJSON batches
	Level         slog.Level        // minimum level exported (on top of the global level)
	BufferSize    int               // queued records; beyond this new records are dropped
	BatchSize     int               // records per POST
	FlushInterval time.Duration     // max time a record waits in a partial batch
	Timeout       time.Duration     // per-request HTTP timeout
	Fields        map[string]string // static fields added to every record (service, host, instance, env)
}

func (c ExportConfig) withDefaults() ExportConfig {
	if c.BufferSize <= 0 {
		c.BufferSize = 10000
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 1000
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = 2 * time.Second
	}
	if c.Timeout <= 0 {
		c.Timeout = 2 * time.Second
	}
	return c
}

const (
	maxBatchBytes     = 1 << 20
	maxBackoff        = 30 * time.Second
	dropReportPeriod  = time.Minute
	initialBackoff    = time.Second
	exportContentType = "application/stream+json"
)

// ExportStats are cumulative exporter counters.
type ExportStats struct {
	Queued  int    `json:"queued"`
	Sent    uint64 `json:"sent"`
	Dropped uint64 `json:"dropped"` // buffer full, or discarded during backoff
	Failed  uint64 `json:"failed"`  // records in batches the endpoint did not accept
}

// Exporter ships log records off the request path. Callers only ever do a
// non-blocking channel send: when the buffer is full the record is dropped
// and counted, so a slow or unreachable log endpoint can never add latency —
// log export is the first thing shed under load. A single goroutine batches,
// gzips and POSTs; failed batches are dropped (durability belongs to the
// local Vector agent, not the app), and while the endpoint is failing the
// flusher backs off and discards instead of retrying.
type Exporter struct {
	cfg    ExportConfig
	client *http.Client
	ch     chan []byte
	done   chan struct{}
	stop   chan struct{}
	once   sync.Once

	sent    atomic.Uint64
	dropped atomic.Uint64
	failed  atomic.Uint64

	// local receives the exporter's own diagnostics (drop reports, failures).
	// It must never be a logger that feeds back into this exporter.
	local slog.Handler
	now   func() time.Time
}

// NewExporter creates an exporter and starts its flusher goroutine.
func NewExporter(cfg ExportConfig, local slog.Handler) *Exporter {
	cfg = cfg.withDefaults()
	e := &Exporter{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		ch:     make(chan []byte, cfg.BufferSize),
		done:   make(chan struct{}),
		stop:   make(chan struct{}),
		local:  local,
		now:    time.Now,
	}
	go e.run()
	return e
}

// Stats returns the exporter counters.
func (e *Exporter) Stats() ExportStats {
	return ExportStats{
		Queued:  len(e.ch),
		Sent:    e.sent.Load(),
		Dropped: e.dropped.Load(),
		Failed:  e.failed.Load(),
	}
}

// Handler returns a slog.Handler that enqueues records to this exporter.
func (e *Exporter) Handler() slog.Handler {
	return &exportHandler{exp: e, level: currentLevel}
}

// Shutdown flushes what is queued, giving up when ctx expires.
func (e *Exporter) Shutdown(ctx context.Context) error {
	e.once.Do(func() { close(e.stop) })
	select {
	case <-e.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *Exporter) enqueue(line []byte) {
	select {
	case e.ch <- line:
	default:
		e.dropped.Add(1)
	}
}

func (e *Exporter) run() {
	defer close(e.done)

	ticker := time.NewTicker(e.cfg.FlushInterval)
	defer ticker.Stop()
	reportTicker := time.NewTicker(dropReportPeriod)
	defer reportTicker.Stop()

	var (
		batch         = make([][]byte, 0, e.cfg.BatchSize)
		batchBytes    int
		backoff       time.Duration
		retryAt       time.Time
		lastDropCount uint64
	)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if !retryAt.IsZero() && e.now().Before(retryAt) {
			// Endpoint is failing: shed instead of piling up retries.
			e.dropped.Add(uint64(len(batch)))
		} else if err := e.send(batch); err != nil {
			e.failed.Add(uint64(len(batch)))
			backoff = min(max(backoff*2, initialBackoff), maxBackoff)
			retryAt = e.now().Add(backoff)
			e.logLocal(slog.LevelWarn, "log export failed", "error", err, "records", len(batch), "backoff", backoff)
		} else {
			e.sent.Add(uint64(len(batch)))
			backoff, retryAt = 0, time.Time{}
		}
		batch = batch[:0]
		batchBytes = 0
	}

	for {
		select {
		case line := <-e.ch:
			batch = append(batch, line)
			batchBytes += len(line) + 1
			if len(batch) >= e.cfg.BatchSize || batchBytes >= maxBatchBytes {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-reportTicker.C:
			if d := e.dropped.Load(); d > lastDropCount {
				e.logLocal(slog.LevelWarn, "log export dropped entries", "dropped", d-lastDropCount, "total_dropped", d)
				lastDropCount = d
			}
		case <-e.stop:
			// Drain what is already queued; ignore backoff for a last attempt.
			retryAt = time.Time{}
			for {
				select {
				case line := <-e.ch:
					batch = append(batch, line)
					if len(batch) >= e.cfg.BatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (e *Exporter) send(batch [][]byte) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	for _, line := range batch {
		if _, err := gz.Write(line); err != nil {
			return err
		}
		if _, err := gz.Write([]byte{'\n'}); err != nil {
			return err
		}
	}
	if err := gz.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, e.cfg.URL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", exportContentType)
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (e *Exporter) logLocal(level slog.Level, msg string, args ...any) {
	if e.local == nil || !e.local.Enabled(context.Background(), level) {
		return
	}
	r := slog.NewRecord(e.now(), level, msg, 0)
	r.Add(args...)
	_ = e.local.Handle(context.Background(), r)
}

// exportHandler serializes records to VictoriaLogs JSON lines (_msg, _time,
// level, static stream fields, trace IDs, attributes) and enqueues them.
type exportHandler struct {
	exp    *Exporter
	level  *slog.LevelVar
	attrs  []groupedAttr // from WithAttrs, each with the groups open at the time
	groups []string
}

type groupedAttr struct {
	groups []string
	attr   slog.Attr
}

func (h *exportHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level() && level >= h.exp.cfg.Level
}

func (h *exportHandler) Handle(ctx context.Context, r slog.Record) error {
	entry := make(map[string]any, 8+len(h.attrs)+r.NumAttrs())

	for _, ga := range h.attrs {
		addAttrToMap(entry, ga.groups, ga.attr)
	}
	r.Attrs(func(a slog.Attr) bool {
		addAttrToMap(entry, h.groups, a)
		return true
	})

	// Reserved fields last so attributes cannot shadow them.
	for k, v := range h.exp.cfg.Fields {
		entry[k] = v
	}
	entry["_msg"] = r.Message
	entry["_time"] = r.Time.UTC().Format(time.RFC3339Nano)
	entry["level"] = levelName(r.Level)
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		entry["trace_id"] = sc.TraceID().String()
		entry["span_id"] = sc.SpanID().String()
	}

	line, err := json.Marshal(entry)
	if err != nil {
		// Unmarshalable attribute: still ship the message.
		line, err = json.Marshal(map[string]any{
			"_msg":         r.Message,
			"_time":        entry["_time"],
			"level":        entry["level"],
			"export_error": err.Error(),
		})
		if err != nil {
			return err
		}
	}
	h.exp.enqueue(line)
	return nil
}

func (h *exportHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]groupedAttr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	for _, a := range attrs {
		newAttrs = append(newAttrs, groupedAttr{groups: h.groups, attr: a})
	}
	return &exportHandler{exp: h.exp, level: h.level, attrs: newAttrs, groups: h.groups}
}

func (h *exportHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups), len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups = append(newGroups, name)
	return &exportHandler{exp: h.exp, level: h.level, attrs: h.attrs, groups: newGroups}
}

func levelName(level slog.Level) string {
	switch {
	case level < slog.LevelInfo:
		return "debug"
	case level < slog.LevelWarn:
		return "info"
	case level < slog.LevelError:
		return "warn"
	default:
		return "error"
	}
}

// multiHandler fans a record out to several handlers, each gated by its own
// Enabled check.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: hs}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: hs}
}
