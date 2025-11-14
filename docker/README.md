# Wayfarer Infrastructure

This directory contains Docker-based infrastructure for the Wayfarer project.

## Services

### Jaeger (Tracing)

Lightweight OpenTelemetry tracing for local debugging. Single container, no persistence needed.

**Location:** `jaeger/`

**Quick Start:**
```bash
cd jaeger
docker compose up -d
```

**Access:** [http://localhost:16686](http://localhost:16686)

See [jaeger/README.md](jaeger/README.md) for detailed documentation.

## Development Workflow

The Wayfarer backend runs natively on your machine (not containerized), while infrastructure services run in Docker:

```
┌─────────────────────────────────────┐
│   Your Machine (macOS)              │
│                                     │
│  ┌──────────────────┐              │
│  │ Wayfarer Backend │              │
│  │  (Go - Native)   │              │
│  │  Port 8080       │              │
│  └────────┬─────────┘              │
│           │                         │
│           │ Sends traces            │
│           │ to localhost:4317       │
│           ▼                         │
│  ┌─────────────────────────┐       │
│  │  Docker: Jaeger         │       │
│  │  - UI :16686            │       │
│  │  - OTLP :4317           │       │
│  └─────────────────────────┘       │
└─────────────────────────────────────┘
```

## Common Operations

### Start tracing
```bash
cd jaeger && docker compose up -d
```

### Stop tracing
```bash
cd jaeger && docker compose down
```

### View logs
```bash
cd jaeger && docker compose logs -f
```

### Check status
```bash
cd jaeger && docker compose ps
```

### Clear traces (restart container)
```bash
cd jaeger && docker compose restart
```

## Resource Management

Jaeger is lightweight (~100MB RAM). When not actively debugging:

```bash
# Stop Jaeger
cd jaeger && docker compose stop

# Start again when needed
cd jaeger && docker compose start
```

## Future Infrastructure

This directory may include:

- **Kafka/Redis** - If event streaming is needed
- **PostgreSQL** - If local database is preferred over remote Neon
- **Prometheus/Grafana** - Additional metrics visualization
- **Testing infrastructure** - Test databases, mock services

## Ports Overview

| Port | Service | Purpose |
|------|---------|---------|
| 16686 | Jaeger UI | Trace visualization |
| 4317 | OTLP gRPC | Receive traces from backend |
| 4318 | OTLP HTTP | Alternative trace endpoint |
| 8080 | Wayfarer Backend | GraphQL API (not in Docker) |

## Troubleshooting

### Ports already in use

Check what's using a port:
```bash
lsof -i :16686
```

### Docker issues

Restart Docker Desktop if containers won't start.

### Services won't start

Check Docker daemon:
```bash
docker ps
```

If no response, restart Docker Desktop.

## Contributing

When adding new infrastructure services:

1. Create a new subdirectory (e.g., `redis/`)
2. Add docker-compose.yaml
3. Document in this README
4. Update port table
5. Add Makefile targets in `/backend/Makefile`
