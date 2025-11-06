# Unit Tests

This directory contains unit tests for utility functions and composables.

## Running Tests

```bash
# Run all unit tests once
pnpm test:unit

# Run tests in watch mode (re-runs on file changes)
pnpm test:watch

# Run all tests
pnpm test
```

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

## Directory Structure

You might organize your tests as follows:

```
test/
├── e2e/
│ └── ssr.test.ts
├── unit/
│ └── utils.test.ts
```

You can of course opt for any test structure, but keeping the Nuxt runtime environment separated from Nuxt end-to-end tests is important for test stability.

## Adding New Tests

1. Create a new `.test.ts` file in this directory
2. Import functions using relative paths: `../../app/...`
3. Use `describe` blocks to group related tests
4. Write clear test descriptions using `it('should ...')`
5. Follow the Arrange-Act-Assert pattern

## Testing Techniques

### Mocking Timers

For date-dependent tests, use vitest's fake timers:

```typescript
import { beforeEach, afterEach, vi } from 'vitest'

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2024-06-15'))
})

afterEach(() => {
  vi.useRealTimers()
})
```

## Notes

- Tests run in Node environment
- Use relative imports (`../../app/...`) instead of aliases (`~/...`)
- All tests run quickly (~200ms for the full suite)
- Focus on testing utility functions and business logic
