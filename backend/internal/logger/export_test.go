package logger

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

// collector is an httptest endpoint that decodes gzip NDJSON batches.
type collector struct {
	mu      sync.Mutex
	batches [][]map[string]any
	status  atomic.Int32
	srv     *httptest.Server
}

func newCollector(t *testing.T) *collector {
	t.Helper()
	c := &collector{}
	c.status.Store(http.StatusOK)
	c.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		gz, err := gzip.NewReader(r.Body)
		if !assert.NoError(t, err) {
			return
		}
		var batch []map[string]any
		sc := bufio.NewScanner(gz)
		for sc.Scan() {
			var m map[string]any
			if assert.NoError(t, json.Unmarshal(sc.Bytes(), &m)) {
				batch = append(batch, m)
			}
		}
		c.mu.Lock()
		c.batches = append(c.batches, batch)
		c.mu.Unlock()
		w.WriteHeader(int(c.status.Load()))
	}))
	t.Cleanup(c.srv.Close)
	return c
}

func (c *collector) records() []map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	var all []map[string]any
	for _, b := range c.batches {
		all = append(all, b...)
	}
	return all
}

func (c *collector) batchCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.batches)
}

func shutdown(t *testing.T, e *Exporter) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, e.Shutdown(ctx))
}

func TestExportHandlerSerializesRecord(t *testing.T) {
	currentLevel.Set(slog.LevelDebug)
	c := newCollector(t)
	exp := NewExporter(ExportConfig{
		URL:    c.srv.URL,
		Level:  slog.LevelDebug,
		Fields: map[string]string{"service": "wayfarer", "instance": "blue"},
	}, nil)

	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled,
	}))

	lgr := slog.New(exp.Handler()).With("req", "r1").WithGroup("db")
	lgr.WarnContext(ctx, "slow query", "ms", 600, "error", errors.New("boom"), "_msg", "shadow")

	shutdown(t, exp)
	recs := c.records()
	require.Len(t, recs, 1)
	r := recs[0]
	assert.Equal(t, "slow query", r["_msg"])
	assert.Equal(t, "warn", r["level"])
	assert.Equal(t, "wayfarer", r["service"])
	assert.Equal(t, "blue", r["instance"])
	assert.Equal(t, "r1", r["req"])
	assert.Equal(t, traceID.String(), r["trace_id"])
	assert.Equal(t, spanID.String(), r["span_id"])
	db, ok := r["db"].(map[string]any)
	require.True(t, ok, "group attrs nest under db")
	assert.Equal(t, float64(600), db["ms"])
	assert.Equal(t, "boom", db["error"])
	_, err := time.Parse(time.RFC3339Nano, r["_time"].(string))
	assert.NoError(t, err)
}

func TestExportHandlerLevels(t *testing.T) {
	currentLevel.Set(slog.LevelDebug)
	exp := NewExporter(ExportConfig{URL: "http://127.0.0.1:1", Level: slog.LevelInfo}, nil)
	defer shutdown(t, exp)
	h := exp.Handler()

	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug), "below export level")
	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))

	currentLevel.Set(slog.LevelError)
	defer currentLevel.Set(slog.LevelInfo)
	assert.False(t, h.Enabled(context.Background(), slog.LevelWarn), "below global level")
}

func TestExporterNeverBlocksWhenEndpointStalls(t *testing.T) {
	currentLevel.Set(slog.LevelInfo)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer srv.Close()
	defer close(release)

	exp := NewExporter(ExportConfig{
		URL:        srv.URL,
		BufferSize: 10,
		BatchSize:  1,
		Timeout:    time.Minute,
	}, nil)
	lgr := slog.New(exp.Handler())

	start := time.Now()
	for i := 0; i < 1000; i++ {
		lgr.Info("spam", "i", i)
	}
	assert.Less(t, time.Since(start), 500*time.Millisecond, "logging must not wait on the endpoint")
	assert.Greater(t, exp.Stats().Dropped, uint64(900))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	assert.ErrorIs(t, exp.Shutdown(ctx), context.DeadlineExceeded, "shutdown honours its deadline")
}

func TestExporterBatchesBySize(t *testing.T) {
	currentLevel.Set(slog.LevelInfo)
	c := newCollector(t)
	exp := NewExporter(ExportConfig{URL: c.srv.URL, BatchSize: 5, FlushInterval: time.Hour}, nil)
	lgr := slog.New(exp.Handler())
	for i := 0; i < 10; i++ {
		lgr.Info("m", "i", i)
	}
	require.Eventually(t, func() bool { return c.batchCount() == 2 }, 2*time.Second, 10*time.Millisecond)
	shutdown(t, exp)
	assert.Len(t, c.records(), 10)
	assert.Equal(t, uint64(10), exp.Stats().Sent)
}

func TestExporterFlushesOnInterval(t *testing.T) {
	currentLevel.Set(slog.LevelInfo)
	c := newCollector(t)
	exp := NewExporter(ExportConfig{URL: c.srv.URL, BatchSize: 1000, FlushInterval: 20 * time.Millisecond}, nil)
	defer shutdown(t, exp)
	slog.New(exp.Handler()).Info("lonely")
	require.Eventually(t, func() bool { return len(c.records()) == 1 }, 2*time.Second, 10*time.Millisecond)
}

func TestExporterBacksOffAfterFailure(t *testing.T) {
	currentLevel.Set(slog.LevelInfo)
	c := newCollector(t)
	c.status.Store(http.StatusServiceUnavailable)
	exp := NewExporter(ExportConfig{URL: c.srv.URL, BatchSize: 1, FlushInterval: time.Hour}, nil)
	lgr := slog.New(exp.Handler())

	lgr.Info("first")
	require.Eventually(t, func() bool { return exp.Stats().Failed == 1 }, 2*time.Second, 10*time.Millisecond)

	// Within the backoff window batches are shed, not sent.
	lgr.Info("second")
	require.Eventually(t, func() bool { return exp.Stats().Dropped == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, 1, c.batchCount())
	assert.Zero(t, exp.Stats().Sent)
	shutdown(t, exp)
}

func TestMultiHandlerRoutesByLevel(t *testing.T) {
	currentLevel.Set(slog.LevelDebug)
	defer currentLevel.Set(slog.LevelInfo)
	c := newCollector(t)
	exp := NewExporter(ExportConfig{URL: c.srv.URL, Level: slog.LevelWarn}, nil)

	local := &recordingHandler{level: slog.LevelDebug}
	lgr := slog.New(newMultiHandler(local, exp.Handler())).With("k", "v")
	lgr.Debug("debug only local")
	lgr.Error("both")

	shutdown(t, exp)
	assert.Equal(t, []string{"debug only local", "both"}, local.msgs)
	recs := c.records()
	require.Len(t, recs, 1)
	assert.Equal(t, "both", recs[0]["_msg"])
	assert.Equal(t, "v", recs[0]["k"])
	assert.True(t, local.hadAttrs, "WithAttrs propagates to every handler")
}

type recordingHandler struct {
	level    slog.Level
	msgs     []string
	hadAttrs bool
}

func (h *recordingHandler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }
func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.msgs = append(h.msgs, r.Message)
	return nil
}
func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { h.hadAttrs = true; return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }
