# Jaeger Tracing for Wayfarer

Simple OpenTelemetry tracing setup for local debugging using Jaeger all-in-one.

## What is Jaeger?

Jaeger is a lightweight distributed tracing system. The "all-in-one" image includes:
- OTLP collector (receives traces from your backend)
- In-memory storage (traces are stored in RAM)
- Web UI (view and search traces)

Perfect for local development and debugging.

## Quick Start

### Start Jaeger

```bash
cd docker/jaeger
docker compose up -d
```

### Access UI

Open http://localhost:16686

### Stop Jaeger

```bash
docker compose down
```

## Using with Wayfarer Backend

1. **Enable tracing** in `backend/.env`:
   ```bash
   OTEL_ENABLED=true
   OTEL_EXPORTER_ENDPOINT=localhost:4317
   ```

2. **Start backend**:
   ```bash
   cd backend
   make dev
   ```

3. **Make requests** to generate traces

4. **View traces** at http://localhost:16686
   - Select service: `wayfarer-backend`
   - Click "Find Traces"

## Ports

| Port | Purpose |
|------|---------|
| 16686 | Jaeger UI |
| 4317 | OTLP gRPC (for backend) |
| 4318 | OTLP HTTP |

## Notes

- **In-memory storage**: Traces are lost when container restarts
- **No persistence**: Perfect for debugging, not for production
- **Lightweight**: Uses minimal resources (~100MB RAM)
- **No alerting**: Just trace visualization

## Troubleshooting

### No traces appearing

Check Jaeger is running:
```bash
docker compose ps
```

Check backend configuration:
```bash
cd backend && grep OTEL .env
```

View Jaeger logs:
```bash
docker compose logs -f
```

### Clear all traces

Restart container (traces are in memory):
```bash
docker compose restart
```

## Resources

- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [OpenTelemetry with Jaeger](https://opentelemetry.io/docs/instrumentation/go/exporters/)
