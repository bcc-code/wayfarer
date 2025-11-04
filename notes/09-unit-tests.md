# Unit Tests

## Test Coverage Summary

### ✅ High Priority Tests (Completed)

#### 1. **Chunk Utility Function** (`internal/utils/slice_test.go`)
- **Coverage**: 100%
- **Test Cases**: 11 total
  - Empty slice handling
  - Slices smaller/equal/larger than chunk size
  - Edge cases (size 0, negative, size 1)
  - Real-world scenario: 1600 items with size 800 (Members API use case)
  - Generic type support (int and string slices)

**Key Test:**
```go
TestChunk/large_slice_with_size_800_(Members_API_use_case)
// Validates correct chunking for Members API batch requests
```

#### 2. **Age Calculation Resolver** (`internal/graph/api/shared_test.go`)
- **Coverage**: Age resolver fully tested
- **Test Cases**: 13 total
  - Nil/empty birthdate handling
  - Invalid date format error handling
  - Correct age calculation
  - Birthday not yet occurred this year (age - 1)
  - Birthday already occurred this year (age)
  - Edge cases: born today, yesterday, 100 years ago
  - Leap year birthdays (Feb 29, 2000)
  - Year boundary dates (Jan 1, Dec 31)

**Key Test:**
```go
TestUserResolver_Age/birthday_not_yet_occurred_this_year
// Ensures age decrements if birthday hasn't passed
```

#### 3. **Gender Normalization** (`internal/handlers/auth_test.go`)
- **Coverage**: 100% of normalizeGender function
- **Test Cases**: 15 total
  - MALE/FEMALE uppercase
  - male/female lowercase
  - M/F abbreviations
  - Mixed case handling
  - Leading/trailing whitespace
  - Unknown values default to MALE
  - Empty string default

**Key Test:**
```go
TestNormalizeGender/with_leading/trailing_spaces
// Validates proper trimming before normalization
```

## Test Results

```bash
✅ internal/utils        - 100.0% coverage - PASS
✅ internal/graph/api    - Age resolver tested - PASS
✅ internal/handlers     - Gender normalization tested - PASS
```

All 39 test cases passing across 3 packages.

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/utils/ -v
go test ./internal/graph/api/ -v -run TestUserResolver_Age
go test ./internal/handlers/ -v -run TestNormalizeGender

# Run with coverage
go test ./internal/utils/ ./internal/graph/api/ ./internal/handlers/ -cover
```

## Future Test Candidates

### Medium Priority
1. **Auth Handler Integration** (`findOrCreateUser` logic)
   - Members API success path
   - Members API failure fallback
   - Invalid person_id handling

2. **Token Cache Behavior** (`internal/auth0/token.go`)
   - Cache hits/misses
   - Token expiration
   - Concurrent access

3. **Members Client** (`internal/members/`)
   - Mock HTTP responses
   - Circuit breaker activation
   - Error handling

### Lower Priority (Integration Tests Better)
- Full authentication flow
- Database interactions
- GraphQL query execution
- DataLoader behavior

## Test Dependencies

The tests use:
- `github.com/stretchr/testify/assert` - Assertions
- Standard library `testing` package
- No external mocks (pure unit tests)

## Notes

- Age calculation uses the same logic as the resolver (DRY principle)
- Tests are deterministic and don't rely on system time for most cases
- Edge cases like leap years and birthday timing are thoroughly covered
- Gender normalization handles all realistic input variations
