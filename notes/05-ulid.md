# ULID Generation

**Date**: 2025-10-31

## Design Pattern

ULIDs (Universally Unique Lexicographically Sortable Identifiers) with 2-character prefixes for type safety and readability.

## ID Format

```
Total length: 28 characters
Format: XX + 26-character ULID
Example: CH01ARZ3NDEKTSV4RRFFQ69G5FAV (Church ID)
```

## Entity Prefixes

All 13 entity types from the database schema:

| Prefix | Entity | Example |
|--------|--------|---------|
| CH | Churches | CH01ARZ3NDEKTSV4RRFFQ69G5FAV |
| US | Users | US01ARZ3NDEKTSV4RRFFQ69G5FAV |
| PR | Projects | PR01ARZ3NDEKTSV4RRFFQ69G5FAV |
| EV | Events | EV01ARZ3NDEKTSV4RRFFQ69G5FAV |
| ST | SuperTeams | ST01ARZ3NDEKTSV4RRFFQ69G5FAV |
| TM | Teams | TM01ARZ3NDEKTSV4RRFFQ69G5FAV |
| SK | Streaks | SK01ARZ3NDEKTSV4RRFFQ69G5FAV |
| CL | Challenges | CL01ARZ3NDEKTSV4RRFFQ69G5FAV |
| AC | Achievements | AC01ARZ3NDEKTSV4RRFFQ69G5FAV |
| RA | Reading Achievement Articles | RA01ARZ3NDEKTSV4RRFFQ69G5FAV |
| LT | Listening Achievement Tracks | LT01ARZ3NDEKTSV4RRFFQ69G5FAV |
| SA | Score Adjustments | SA01ARZ3NDEKTSV4RRFFQ69G5FAV |

## Type-Safe Generator Functions

Each entity type has its own generator function:

```go
ulid.NewChurchID()              // CH...
ulid.NewUserID()                // US...
ulid.NewProjectID()             // PR...
ulid.NewEventID()               // EV...
ulid.NewSuperTeamID()           // ST...
ulid.NewTeamID()                // TM...
ulid.NewStreakID()              // SK...
ulid.NewChallengeID()           // CL...
ulid.NewAchievementID()         // AC...
ulid.NewReadingAchievementID()  // RA...
ulid.NewListeningAchievementID() // LT...
ulid.NewScoreAdjustmentID()     // SA...
```

## Benefits

1. **Type Safety**: Different function for each entity type prevents mixing IDs
2. **Readability**: Prefix makes it clear what entity type an ID belongs to
3. **Debugging**: Easy to identify entity types in logs and errors
4. **Sortable**: ULIDs are lexicographically sortable by timestamp
5. **Unique**: Cryptographically random, monotonic within same millisecond

## Validation Functions

Each entity type has its own validation:

```go
ulid.IsChurchID(id)   // Validates CH prefix and ULID format
ulid.IsUserID(id)     // Validates US prefix and ULID format
// ... etc for all entity types
```

Generic validation:

```go
ulid.IsValidID(id, ulid.PrefixUser)  // Validate with specific prefix
ulid.GetPrefix(id)                    // Extract the 2-char prefix
ulid.GetTimestamp(id)                 // Extract creation timestamp
```

## Thread Safety

The ULID generation uses a mutex to ensure thread-safe concurrent ID generation. The monotonic entropy source maintains ordering even when IDs are generated in the same millisecond.

## Implementation Details

- Uses `github.com/oklog/ulid/v2` library
- Monotonic entropy source for guaranteed ordering
- Mutex protection for concurrent generation
- Timestamp-based for sortability
- Crypto/rand for security

## Testing

Full test coverage includes:
- All 12 ID generator functions
- Prefix validation
- ULID format validation
- Timestamp extraction
- Lexicographic sorting
- Concurrent ID generation (100 goroutines × 100 IDs each)

All tests pass, confirming thread-safety and correctness.

## Usage in Code

```go
// Generate IDs
userID := ulid.NewUserID()
churchID := ulid.NewChurchID()

// Validate IDs
if !ulid.IsUserID(userID) {
    return errors.New("invalid user ID")
}

// Extract information
prefix := ulid.GetPrefix(userID)  // "US"
created := ulid.GetTimestamp(userID)  // time.Time
```

## Next Steps

These ID generators will be used:
1. In sqlc queries (for INSERT statements)
2. In GraphQL resolvers (when creating new entities)
3. In database seeders (for test data)
