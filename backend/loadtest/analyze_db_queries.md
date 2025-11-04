# Database Query Analysis for Load Tests

## Issues Found and Fixed

### 1. ✅ Root `projects` Query (FIXED)
**Location:** `internal/graph/api/schema.resolvers.go:396`

**Problem:** The root `projects` query was directly querying the database on every request without any caching.

```graphql
query {
  projects {
    id
  }
}
```

**Solution:** Added cache check before database query with cache key `"projects:all"` and 15-minute TTL.

**Impact:** This query appears in EVERY load test request, so this was causing massive DB load.

---

### 2. ✅ RoleScope.project Field (FIXED)
**Location:** `internal/graph/api/shared.resolvers.go:113`

**Problem:** When querying user roles with project scope, the nested `scope.project` field was returning an error instead of loading the project.

```graphql
query {
  me {
    roles {
      scope {
        project {
          name
          id
          description
        }
      }
    }
  }
}
```

**Solution:** Wired up the `ProjectByIDLoader` to resolve the project field using the cached loader.

**Impact:** Users with project-scoped roles were causing errors or additional DB queries.

---

## How to Verify Cache is Working

### Method 1: Enable Database Query Logging

Add query logging to see what's hitting the database:

```go
// In internal/database/database.go, add query logging
import "github.com/jackc/pgx/v5/tracelog"

// In Connect function, add:
config.Tracer = &tracelog.TraceLog{
    Logger: tracelog.LoggerFunc(func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
        slog.Info("DB Query", "msg", msg, "data", data)
    }),
    LogLevel: tracelog.LogLevelInfo,
}
```

### Method 2: Monitor Cache Hit Rates

Add a metrics endpoint to expose cache statistics:

```go
// In cmd/server/main.go
router.GET("/metrics/cache", func(c *gin.Context) {
    metrics := cacheInstance.Metrics()
    c.JSON(http.StatusOK, gin.H{
        "hits":      metrics.Hits(),
        "misses":    metrics.Misses(),
        "hit_rate":  float64(metrics.Hits()) / float64(metrics.Hits()+metrics.Misses()),
        "cost":      metrics.CostAdded(),
        "evictions": metrics.KeysEvicted(),
    })
})
```

Then during load test:
```bash
# Watch cache hit rate in real-time
watch -n 1 'curl -s http://localhost:8080/metrics/cache | jq'
```

### Method 3: Use Database Connection Pool Stats

Monitor active database connections:

```bash
# While load test is running
watch -n 1 'curl -s http://localhost:8080/health'  # Add DB stats to health endpoint
```

---

## Expected Performance After Fixes

With 34 different users in the load test:

**Before (no cross-request caching):**
- `projects` query: Hits DB on every single request (3000+ queries during test)
- `RoleScope.project`: Error or additional DB queries per user with roles

**After (with caching):**
- `projects` query: Hits DB once, then served from cache for 15 minutes
- `RoleScope.project`: Hits DB once per unique project, then cached
- Expected cache hit rate: 60-80%+ for realistic load test

---

## Remaining Uncached Queries

These queries are NOT yet cached and may still hit the database:

1. **`me.roles.assignedBy.name`** - Loads user who assigned each role
   - Currently cached via `UserByIDLoader` ✅

2. **`me.church`** - Loads user's church
   - Currently cached via `ChurchLoader` ✅

3. **`me.projects`** - Loads user's projects
   - Currently cached via `ProjectsByUserLoader` ✅

4. **`me.roles`** - Loads user's roles
   - Currently cached via `RolesByUserLoader` ✅

All the main queries in the load test are now cached! 🎉

---

## Next Steps for Further Optimization

If you still see high DB load after these fixes:

1. **Check for N+1 queries** - Use DataLoader profiling to identify patterns
2. **Cache computed data** - Leaderboards, scores, etc. (Phase 2)
3. **Implement query complexity limits** - Prevent overly complex queries
4. **Add cache warming** - Pre-populate cache for active projects/events
