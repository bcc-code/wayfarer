# Cache Invalidation Patterns

## Overview

The system uses a two-tier caching system with tag-based invalidation to ensure data consistency across mutations. Cache keys are registered under multiple tags (prefixes) to enable efficient bulk invalidation.

## Cache Architecture

- **Base Cache**: Ristretto (high-performance concurrent cache)
- **Cache with Registry**: Extended version that tracks all cache keys for tag-based bulk invalidation
- **Location**: `backend/internal/cache/`

### Key Pattern Structure

Cache keys follow specific patterns to enable tag-based invalidation:

- **Direct entity keys**: `project:PROJECTID`, `event:EVENTID`, etc.
- **Related entity keys**: `event:project:PROJECTID`, `team:project:PROJECTID`, etc.
- **Query/filter keys**: `projectsfilter:HASH`, `eventsfilter:HASH`, etc.
- **Leaderboard keys**: `leaderboard:project:PROJECTID:entityType:HASH:page`

### Tag Extraction

The `extractPrefixes()` function in `backend/internal/cache/invalidation.go` registers cache keys under multiple prefixes:

- Main entity prefix (e.g., `event:`, `team:`)
- Project tag (extracted from `:project:PROJECTID` pattern)
- Event tag (extracted from `:event:EVENTID` pattern)
- User tag (extracted from user-related patterns)

When `DeletePrefix("project:PROJECTID")` is called, it finds all keys registered under that prefix and deletes them.

## Invalidation Helper Functions

Located in `backend/internal/cache/invalidation.go`:

### InvalidateProject

Invalidates all cache entries related to a project:

```go
func (c *CacheWithRegistry) InvalidateProject(projectID string) {
    // Direct project entity
    c.Delete(ProjectKey(projectID))

    // All related entities (events, teams, challenges, etc. by project)
    c.DeletePrefix("project:" + projectID)

    // All project list/filter queries (any filter combination)
    c.DeletePrefix(PrefixProjectsFilter)
    c.DeletePrefix(PrefixProjectsCount)

    // All leaderboard data for this project
    c.DeletePrefix("leaderboard:project:" + projectID)
    c.DeletePrefix("leaderboard:position:project:" + projectID)
    c.DeletePrefix("leaderboard:count:project:" + projectID)
    c.DeletePrefix("leaderboard:full:project:" + projectID)
}
```

**Why invalidate ALL project filter queries?**
Filter query cache keys are hashed and don't contain the project ID, so we can't selectively invalidate only queries involving the changed project. This is acceptable because:
- Project updates are relatively infrequent
- The performance impact is minimal (cache refills on next query)
- Correctness is more important than cache efficiency

**Why separate leaderboard invalidation?**
Leaderboard keys use the pattern `leaderboard:project:PROJECTID:...` which starts with the `leaderboard:` prefix, not `project:`. The tag extraction would extract `PROJECTID:...` instead of just `PROJECTID`, so these keys are NOT registered under the `project:PROJECTID` tag. They must be explicitly invalidated.

### InvalidateEvent

Invalidates all cache entries related to an event:

```go
func (c *CacheWithRegistry) InvalidateEvent(eventID string) {
    c.Delete(EventKey(eventID))
    c.DeletePrefix("event:" + eventID)
}
```

### InvalidateUser

Invalidates all cache entries related to a user:

- Direct user entity (`user:{userID}`)
- User's project and event lists (`userprojects:{userID}`, `userevents:{userID}`)
- User roles (`userroles:{userID}`)
- All keys tagged with `user:{userID}` (reverse lookups, etc.)
- Challenge enrollments and completions
- Content progress (reading/listening achievements)
- Achievement timestamps (earned and celebrated)
- Streak activity
- User filter/count queries (since user attributes like gender/church affect results)

Called automatically by `UserSyncService.SyncUser()` and `MaintenanceHandler` sync endpoints after updating user data.

### Other Helpers

- `InvalidateTeam(teamID)` - Invalidates team and members
- `InvalidateChallenge(challengeID)` - Invalidates challenge
- `InvalidateAchievement(achievementID)` - Invalidates achievement

## Mutation Patterns

### Creating Child Entities

When creating a child entity (e.g., Event, Challenge), invalidate the parent project to clear list queries:

```go
func (r *mutationResolver) CreateEvent(ctx context.Context, projectID string, input model.CreateEventInput) (*model.Event, error) {
    // ... create event in database ...

    // Invalidate project cache to reflect new event in lists
    r.Cache.InvalidateProject(projectID)

    return event, nil
}
```

**Why?** The parent project's cached list of events (key: `event:project:PROJECTID`) becomes stale when a new event is added.

### Deleting Child Entities

When deleting a child entity, invalidate both the entity itself AND the parent project:

```go
func (r *mutationResolver) DeleteEvent(ctx context.Context, id string) (bool, error) {
    // Load existing event to get project ID
    existingEvent, err := r.Loaders.EventByIDLoader.Load(ctx, id)()
    // ... delete from database ...

    // Invalidate cache
    r.Cache.InvalidateEvent(id)
    r.Cache.InvalidateProject(existingEvent.ProjectID)

    return true, nil
}
```

**Why?** Both the event itself and the parent project's event list need to be invalidated.

### Updating Entities

When updating an entity, invalidate the entity itself:

```go
func (r *mutationResolver) UpdateEvent(ctx context.Context, id string, input model.UpdateEventInput) (*model.Event, error) {
    // ... update in database ...

    // Invalidate cache
    r.Cache.InvalidateEvent(id)

    return event, nil
}
```

**Why?** The entity's cached data becomes stale. The `InvalidateEvent()` call will also invalidate related entities through the tag system.

### Moving Entities Between Parents

When moving an entity between parents (e.g., MoveEvent), invalidate BOTH the old and new parents:

```go
func (r *mutationResolver) MoveEvent(ctx context.Context, id string, newProjectID string) (*model.Event, error) {
    // Get old project ID
    existingEvent, err := r.Loaders.EventByIDLoader.Load(ctx, id)()
    oldProjectID := existingEvent.ProjectID

    // ... move event in database ...

    // Invalidate cache for both old and new projects
    r.Cache.InvalidateEvent(id)
    r.Cache.InvalidateProject(oldProjectID)
    r.Cache.InvalidateProject(newProjectID)

    return event, nil
}
```

**Why?** Both the old project's event list and the new project's event list become stale.

## Current Implementation Status

### Implemented Mutations

These mutations have correct cache invalidation:

- `CreateProject` - N/A (no parent to invalidate)
- `UpdateProject` - Invalidates project, project lists, and leaderboards
- `DeleteProject` - N/A (not implemented)
- `CreateEvent` - Invalidates parent project
- `UpdateEvent` - Invalidates event
- `DeleteEvent` - Invalidates event and parent project
- `MoveEvent` - Invalidates event and both old/new projects
- `CreateChallenge` - Invalidates parent project
- `UpdateChallenge` - Invalidates challenge
- `DeleteChallenge` - Invalidates challenge and parent project

### Not Yet Implemented

These mutations are not implemented yet (panic) but should follow the same patterns when implemented:

- `CreateTeam` / `DeleteTeam` - Should invalidate parent project
- `CreateSuperTeam` / `DeleteSuperTeam` - Should invalidate parent project
- `CreateAchievement` / `DeleteAchievement` - Should invalidate parent project
- `CreateStreak` / `DeleteStreak` - Should invalidate parent project

## Testing Cache Invalidation

When implementing new mutations:

1. Run `make gqltest` in `./backend/` to verify GraphQL functionality
2. Manually test that:
   - Creating/deleting entities updates parent lists
   - Updating entities reflects changes in queries
   - Filter queries return updated results
   - Leaderboards reflect score changes

## Common Pitfalls

1. **Forgetting to invalidate parent on Create/Delete** - Most common issue. Always invalidate the parent project when creating or deleting child entities.

2. **Not invalidating filter queries** - When updating top-level entities (like projects), remember to invalidate all filter queries since they're hashed.

3. **Missing leaderboard invalidation** - Leaderboard keys have a different prefix pattern and must be explicitly invalidated.

4. **Not loading entity before delete** - Always load the entity first to get parent IDs needed for invalidation.

## References

- Cache implementation: `backend/internal/cache/`
- Invalidation helpers: `backend/internal/cache/invalidation.go`
- Cache keys: `backend/internal/cache/keys.go`
- Mutation examples: `backend/internal/graph/api/schema.resolvers.go`
