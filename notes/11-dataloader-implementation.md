# Dataloader Implementation Guide

## Overview

Dataloaders batch and cache database queries to prevent N+1 query problems in GraphQL resolvers. They use `github.com/graph-gophers/dataloader/v7` with custom caching.

## Architecture

- **Location**: `backend/internal/loaders/`
- **Global instance**: Created once at server startup in `NewLoaders()`
- **Two-tier caching**:
  1. Dataloader's built-in per-request batching
  2. Application-level cache (`github.com/Code-Hex/go-generics-cache`)

## Creating a New Dataloader

### 1. Add cache key function in `backend/internal/cache/keys.go`

```go
const PrefixEntity = "entity:"

func EntityKey(entityID string) string {
    return PrefixEntity + entityID
}
```

### 2. Create batch function in `backend/internal/loaders/entity.go`

```go
func entityByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Entity] {
    return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Entity] {
        // 1. Check cache for each ID
        entityMap := make(map[string]*model.Entity)
        missingIDs := []string{}

        for _, id := range ids {
            cacheKey := cache.EntityKey(id)
            if cached, ok := c.Get(cacheKey); ok {
                if entity, ok := cached.(*model.Entity); ok {
                    entityMap[id] = entity
                    continue
                }
            }
            missingIDs = append(missingIDs, id)
        }

        // 2. Query database only for cache misses
        if len(missingIDs) > 0 {
            rows, err := db.Queries.GetEntitiesByIDs(ctx, missingIDs)
            if err != nil {
                // Return error for all IDs
                results := make([]*dataloader.Result[*model.Entity], len(ids))
                for i := range results {
                    results[i] = &dataloader.Result[*model.Entity]{Error: err}
                }
                return results
            }

            // 3. Convert to GraphQL model and populate cache
            for _, row := range rows {
                entity := &model.Entity{
                    ID:   row.ID,
                    Name: row.Name,
                    // ... other fields
                }

                entityMap[row.ID] = entity
                // Set cache with appropriate TTL
                c.Set(cache.EntityKey(row.ID), entity)
                // Or: c.SetWithTTL(cache.EntityKey(row.ID), entity, 30*time.Minute)
            }
        }

        // 4. Return results in same order as input IDs
        results := make([]*dataloader.Result[*model.Entity], len(ids))
        for i, id := range ids {
            if entity, ok := entityMap[id]; ok {
                results[i] = &dataloader.Result[*model.Entity]{Data: entity}
            } else {
                results[i] = &dataloader.Result[*model.Entity]{
                    Error: fmt.Errorf("entity not found: %s", id),
                }
            }
        }
        return results
    }
}
```

### 3. Register in `backend/internal/loaders/loaders.go`

```go
type Loaders struct {
    // ... existing loaders
    EntityByIDLoader *dataloader.Loader[string, *model.Entity]
}

func NewLoaders(db *database.DB, cache *cache.CacheWithRegistry) *Loaders {
    return &Loaders{
        // ... existing loaders
        EntityByIDLoader: dataloader.NewBatchedLoader(
            entityByIDBatchFunc(db, cache),
            dataloader.WithBatchCapacity[string, *model.Entity](100),
        ),
    }
}
```

## Using Dataloaders in Resolvers

### Example: Loading a single entity

```go
func (r *someResolver) Entity(ctx context.Context, obj *model.Parent) (*model.Entity, error) {
    // Load returns a Thunk that must be called
    thunk := r.Loaders.EntityByIDLoader.Load(ctx, obj.EntityID)
    entity, err := thunk()
    if err != nil {
        return nil, fmt.Errorf("failed to load entity: %w", err)
    }
    return entity, nil
}
```

The dataloader automatically:
- Batches multiple Load() calls within the same request tick
- Caches results for the request lifetime
- Uses application cache for cross-request caching

## Cache TTL Guidelines

- **Frequently changing data**: Use default TTL (15 minutes) via `c.Set()`
- **Rarely changing data**: Use longer TTL via `c.SetWithTTL(key, value, 30*time.Minute)`
- Examples:
  - Users: 15 minutes (default)
  - Churches: 30 minutes (rarely change)
  - Projects: 15 minutes (default)

## Important Patterns

1. **Always preserve order**: Return results in the same order as input IDs
2. **Handle missing entities**: Return explicit errors for not-found cases
3. **Error handling**: If DB query fails, return error for ALL IDs in batch
4. **Cache misses only**: Only query DB for IDs not in cache
5. **Type safety**: Use generics for type-safe loaders

## File Organization

```
backend/internal/loaders/
├── loaders.go           # Loader registry and initialization
├── user.go             # User batch function
├── church.go           # Church batch function
├── project_by_id.go    # Project batch function
└── [entity].go         # New entity batch functions
```

## References

- Example implementations: `backend/internal/loaders/user.go`, `backend/internal/loaders/church.go`
- Cache keys: `backend/internal/cache/keys.go`
- Resolver usage: `backend/internal/graph/api/shared.resolvers.go`
