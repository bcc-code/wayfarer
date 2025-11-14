# Wayfarer Observability

This document describes the observability setup for the Wayfarer backend using OpenTelemetry and Jaeger.

## Overview

The Wayfarer backend uses **OpenTelemetry** for distributed tracing and **Jaeger** as the trace viewer. This setup provides:

- **Distributed tracing** - Track requests end-to-end
- **Performance insights** - Identify slow queries and bottlenecks
- **Database query tracking** - See every SQL query with timing
- **GraphQL operation tracking** - Monitor resolver performance
- **Simple debugging** - Lightweight, no persistence needed

## Architecture

```
┌─────────────────────────────────────────────┐
│  Wayfarer Backend (Go)                      │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ HTTP Request (Gin)                  │   │
│  │  └─> Span: http.request             │   │
│  │                                      │   │
│  │  ┌──────────────────────────────┐   │   │
│  │  │ GraphQL Query (gqlgen)       │   │   │
│  │  │  └─> Span: graphql.operation │   │   │
│  │  │                               │   │   │
│  │  │  ┌──────────────────────┐    │   │   │
│  │  │  │ Database Query (pgx) │    │   │   │
│  │  │  │  └─> Span: db.query  │    │   │   │
│  │  │  └──────────────────────┘    │   │   │
│  │  └──────────────────────────────┘   │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  Traces sent to ──>  localhost:4317        │
└─────────────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────┐
│  Jaeger (Docker)                             │
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │   Jaeger All-in-One                  │   │
│  │   - OTLP Collector (Port 4317)       │   │
│  │   - In-Memory Storage                │   │
│  │   - Web UI (Port 16686)              │   │
│  └──────────────────────────────────────┘   │
└──────────────────────────────────────────────┘
```

## Quick Start

### 1. Start Jaeger

```bash
cd backend
make jaeger-up
```

Starts immediately. Access the UI at http://localhost:16686

### 2. Enable Tracing in Backend

Create or update `backend/.env`:

```bash
# Enable OpenTelemetry tracing
OTEL_ENABLED=true
OTEL_SERVICE_NAME=wayfarer-backend
OTEL_EXPORTER_ENDPOINT=localhost:4317
OTEL_EXPORTER_INSECURE=true
OTEL_SAMPLING_RATIO=1.0
```

### 3. Run Backend

```bash
make dev
```

### 4. Generate Traffic

Make some GraphQL requests to generate traces:

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $YOUR_TOKEN" \
  -d '{"query":"{ projects(first: 10) { edges { node { id name } } } }"}'
```

### 5. View Traces

Open http://localhost:16686:
1. Select service: `wayfarer-backend`
2. Click "Find Traces"
3. Click on a trace to see details

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_ENABLED` | `false` | Enable/disable OpenTelemetry tracing |
| `OTEL_SERVICE_NAME` | `wayfarer-backend` | Service name in traces |
| `OTEL_SERVICE_VERSION` | `dev` | Service version |
| `OTEL_EXPORTER_ENDPOINT` | `localhost:4317` | OTLP gRPC endpoint |
| `OTEL_EXPORTER_INSECURE` | `true` | Use insecure connection (no TLS) |
| `OTEL_SAMPLING_RATIO` | `1.0` | Sampling ratio (0.0 to 1.0) |

### Sampling Strategies

**Development (100% sampling):**
```bash
OTEL_SAMPLING_RATIO=1.0  # Capture every trace
```

**Production (10% sampling):**
```bash
OTEL_SAMPLING_RATIO=0.1  # Capture 10% of traces
```

**Disabled:**
```bash
OTEL_ENABLED=false  # No tracing overhead
```

## Instrumentation Layers

The backend automatically instruments the following layers when `OTEL_ENABLED=true`:

### 1. HTTP Layer (Gin)
- **Location:** `cmd/server/main.go:176`
- **Middleware:** `otelgin.Middleware()`
- **Captures:**
  - HTTP method, path, status code
  - Request duration
  - Client IP
  - User agent

### 2. GraphQL Layer (gqlgen)
- **Location:** `cmd/server/main.go:169`
- **Middleware:** `otelgqlgen.Middleware()`
- **Captures:**
  - GraphQL operation name
  - Operation type (query/mutation)
  - Field resolvers
  - Arguments
  - Errors

### 3. Database Layer (pgx)
- **Location:** `internal/database/database.go:40`
- **Tracer:** `otelpgx.NewTracer()`
- **Captures:**
  - SQL statement (trimmed)
  - Query duration
  - Row counts
  - Connection info

## Using Traces for Performance Analysis

### Finding Slow Requests

1. Open SigNoz UI at http://localhost:3301
2. Go to "Traces" tab
3. Sort by "Duration" descending
4. Click on a slow trace to see the breakdown

### Analyzing a Trace

Each trace shows:
- **Timeline view** - Visual breakdown of time spent
- **Span details** - Attributes like SQL statements, GraphQL operations
- **Errors** - Stack traces if any errors occurred

### Common Patterns to Look For

#### N+1 Query Problem

**Symptom:** Many sequential database queries in a trace

**Example trace structure:**
```
http.request (2.5s)
└─ graphql.operation: projects (2.5s)
   ├─ db.query: GetProjects (50ms)
   ├─ db.query: GetProjectPersonLeaderboard (200ms)
   ├─ db.query: GetProjectPersonLeaderboard (180ms)
   ├─ db.query: GetProjectPersonLeaderboard (190ms)
   └─ ... (10 more similar queries)
```

**Solution:** Use DataLoaders or batch queries

#### Slow Database Query

**Symptom:** Single db.query span taking most of the time

**Example:**
```
http.request (3.2s)
└─ graphql.operation: leaderboard (3.2s)
   └─ db.query: GetProjectPersonLeaderboard (3.1s)
```

**Solution:** Check query plan, add indices, optimize query

#### External API Latency

**Symptom:** Time spent outside database/GraphQL

**Example:**
```
http.request (1.5s)
└─ graphql.operation: me (1.5s)
   ├─ http.client: members-api (1.3s)
   └─ db.query: GetUser (50ms)
```

**Solution:** Add caching, use circuit breakers, reduce API calls

### Filtering Traces

Use Jaeger UI to filter:

- **Service:** Dropdown at top
- **Operation:** Select specific GraphQL operations
- **Tags:** Search by `http.method`, `db.statement`, etc.
- **Min/Max Duration:** Find slow requests
- **Lookback:** Time range (last 1h, 24h, etc.)

## Understanding Traces

### Trace View

Each trace shows:
- **Timeline** - Visual waterfall of spans
- **Spans** - Individual operations with duration
- **Tags** - Metadata (SQL statements, GraphQL ops, etc.)
- **Logs** - Events within spans

## Troubleshooting

### No traces appearing

**Check 1:** Is Jaeger running?
```bash
make jaeger-ps
```

**Check 2:** Is tracing enabled?
```bash
grep OTEL_ENABLED .env
```

**Check 3:** Check Jaeger logs
```bash
make jaeger-logs
```

### High overhead

If tracing is causing performance issues:

1. **Reduce sampling:**
   ```bash
   OTEL_SAMPLING_RATIO=0.1  # Sample 10%
   ```

2. **Disable query logging:**
   ```bash
   DB_LOG_QUERIES=false
   ```

3. **Temporarily disable:**
   ```bash
   OTEL_ENABLED=false
   ```

### Missing database queries

Database queries only appear if:
1. OTEL is enabled (`OTEL_ENABLED=true`)
2. `DB_LOG_QUERIES=false` (tracelog replaces OTEL tracer)

To see all queries in traces, ensure `DB_LOG_QUERIES=false`.

### Traces disappear after restart

Jaeger uses in-memory storage for debugging. Traces are lost on restart. This is intentional for local development.

## Best Practices

### Development

- ✅ Enable 100% sampling (`OTEL_SAMPLING_RATIO=1.0`)
- ✅ Keep Jaeger running while debugging (`make jaeger-up`)
- ✅ Review traces for slow operations
- ✅ Restart Jaeger to clear old traces (`make jaeger-restart`)

### Important Notes

- 🔄 **In-memory only** - Traces lost on restart (by design)
- 🐛 **For debugging** - Not for production monitoring
- 💾 **No persistence** - No disk storage needed
- ⚠️ **Don't log sensitive data** - Traces may contain SQL queries

### Security Considerations

- 🔒 Don't include passwords or tokens in span attributes
- 🔒 Use secure connections in production (`INSECURE=false`)
- 🔒 Limit access to SigNoz UI
- 🔒 Sanitize SQL queries to remove sensitive data

## Advanced Usage

### Custom Spans

To add custom spans in your code:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func YourFunction(ctx context.Context) error {
    tracer := otel.Tracer("wayfarer-backend")
    ctx, span := tracer.Start(ctx, "YourFunction")
    defer span.End()

    // Add custom attributes
    span.SetAttributes(
        attribute.String("user.id", userID),
        attribute.Int("batch.size", len(items)),
    )

    // Your code here

    return nil
}
```

### Span Events

Add events to existing spans:

```go
span.AddEvent("cache.miss",
    trace.WithAttributes(
        attribute.String("cache.key", key),
    ),
)
```

### Recording Errors

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

## Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [OpenTelemetry Spans Guide](https://signoz.io/blog/opentelemetry-spans/)
- [Jaeger GitHub](https://github.com/jaegertracing/jaeger)

## Related Documentation

- [Profiling Guide](./profiling.md) - CPU and memory profiling with pprof
- [Database Optimization](../internal/database/README.md) - Query optimization tips
- [Cache Strategy](../internal/cache/README.md) - Caching best practices
