# Unit Tests

This directory contains unit tests for utility functions and composables.

## Running Tests

```bash
# Run all unit tests once
pnpm test:unit

# Run tests in watch mode (re-runs on file changes)
pnpm test:watch

# Run all tests (unit + nuxt)
pnpm test
```

## Test Coverage

### `formatters.test.ts` - Utility Functions

Tests for string and date formatting utilities:

- **getInitials()** - Extract initials from full names
  - Handles single/multiple names
  - Handles extra spaces
  - Handles mixed case

- **formatDate()** - Format ISO date strings to readable format
  - Formats to "Month Day, Year" format
  - Handles dates with time components
  - Works with past and future dates

- **capitalizeFirst()** - Capitalize first letter, lowercase rest
  - Handles all uppercase/lowercase/mixed case
  - Handles special characters and numbers
  - Useful for formatting enum values like "MALE" → "Male"

### `usePagination.test.ts` - Pagination Composable

Tests for the Relay cursor-based pagination composable:

- **Initialization**
  - Default and custom page sizes
  - Initial cursor support
  - Correct initial state

- **updateConnection()** - Update state from GraphQL response
  - Updates pageInfo and totalCount
  - Handles null/undefined connections

- **Navigation Methods**
  - nextPage() - Navigate forward with cursor
  - previousPage() - Navigate backward with cursor
  - firstPage() - Reset to first page

- **State Management**
  - reset() - Reset all state
  - setPageSize() - Change page size and reset

- **Computed Properties**
  - isFirstPage - Correctly detects first page
  - isLastPage - Correctly detects last page

## Test Structure

Each test file follows this pattern:

```typescript
import { describe, it, expect } from 'vitest'
import { functionToTest } from '../../app/path/to/function'

describe('feature name', () => {
  describe('specific function', () => {
    it('should do something specific', () => {
      // Arrange
      const input = 'test'

      // Act
      const result = functionToTest(input)

      // Assert
      expect(result).toBe('expected')
    })
  })
})
```

## Adding New Tests

1. Create a new `.test.ts` file in this directory
2. Import functions using relative paths: `../../app/...`
3. Use `describe` blocks to group related tests
4. Write clear test descriptions using `it('should ...')`
5. Follow the Arrange-Act-Assert pattern

## Notes

- Tests run in Node environment (not Nuxt environment)
- Use relative imports (`../../app/...`) instead of aliases (`~/...`)
- Vue composables that use Nuxt features should be tested in `test/nuxt/` instead
