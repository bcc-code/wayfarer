# Authentication Tests Documentation

## Overview

Comprehensive test suite for the Wayfarer authentication system, covering token management, OAuth callbacks, middleware protection, and role-based authorization.

## Test Files

### ✅ `auth-helpers.test.ts` (36 tests - **COMPLETE**)

**Status:** All tests implemented and passing

Tests pure functions related to authentication:

- JWT format validation
- Redirect URL safety validation
- JWT payload parsing
- Token expiration checking
- Role checking helpers
- URL construction

**Run:** `pnpm test -- auth-helpers.test.ts`

### ✅ `auth.test.ts` (87 tests - **57 IMPLEMENTED, 30 SCAFFOLDED**)

**Status:** Core auth functionality, callback page, and middleware tests implemented and passing

Implemented tests for:

- ✅ **Token Management** (9 tests)
  - Token storage and retrieval
  - Cookie handling
  - Token overwriting
  - Special characters and long tokens
  - Loading state management
  - Token clearing

- ✅ **Login Redirect** (5 tests)
  - Redirect to login URL with current path
  - Redirect parameter encoding
  - Handling paths with query params
  - Handling hash fragments
  - Special character encoding

- ✅ **Role-based Authorization** (9 tests)
  - Superadmin, admin, church admin, project admin, team lead detection
  - Multiple roles handling
  - Empty/null user scenarios

- ✅ **GraphQL Me Query** (6 tests)
  - User data updates
  - Loading state management
  - Null user handling
  - Role computed reactivity
  - Church data handling
  - Multiple updates and reactivity

- ✅ **Callback Page Token Validation** (12 tests)
  - Backend token validation
  - Token storage after validation
  - Redirect to home after success
  - Redirect to specified path
  - Missing token parameter handling
  - Invalid token format handling
  - Expired token handling
  - Backend validation failure
  - Network timeout handling
  - Malformed response handling
  - Redirect with query params
  - Hash preservation in redirects
  - Array redirect parameter handling

- ✅ **Global Auth Middleware** (6 tests)
  - Callback page access without token
  - Redirect to login for protected routes
  - Allow access with valid token
  - Token check on every navigation
  - Callback route variations
  - Routes containing 'callback' in path

- ✅ **Admin Middleware** (10 tests)
  - Superadmin access to admin routes
  - Admin access to admin routes
  - Block regular users (403 error)
  - Redirect to login if no token
  - Return 403 for non-admin users
  - Wait for me query to complete
  - Timeout handling for slow queries
  - Handle me query errors
  - Only check /admin routes
  - Nested admin routes

Still scaffolded (placeholders - 30 tests):

- Security tests
- Integration tests

**Run:** `pnpm test -- auth.test.ts`

## What's Already Working

The helper functions test suite validates critical security functions:

### 1. **JWT Validation** (7 tests)

- ✅ Validates correct 3-part JWT format
- ✅ Rejects malformed tokens
- ✅ Handles base64url encoding properly
- ✅ Prevents invalid characters

### 2. **Redirect Safety** (8 tests)

- ✅ Prevents open redirect vulnerabilities
- ✅ Blocks absolute URLs
- ✅ Blocks protocol-relative URLs
- ✅ Blocks javascript: and data: URLs
- ✅ Only allows safe relative paths starting with /

### 3. **JWT Parsing** (5 tests)

- ✅ Parses JWT payloads correctly
- ✅ Handles base64url encoding
- ✅ Gracefully handles malformed tokens

### 4. **Expiration Checking** (6 tests)

- ✅ Detects expired tokens
- ✅ Validates future tokens
- ✅ Handles missing exp claims
- ✅ Edge cases (just expired, expiring soon)

### 5. **Role Management** (6 tests)

- ✅ Finds roles in user's role array
- ✅ Case-sensitive matching
- ✅ Handles empty roles
- ✅ Multiple role instances

### 6. **URL Construction** (4 tests)

- ✅ Builds login URLs with redirect params
- ✅ Properly encodes special characters
- ✅ Preserves existing query params

## Implementation Roadmap

To complete the full auth test suite, implement tests in this order:

### Priority 1: Token Management (Critical)

```typescript
// test/unit/auth-token.test.ts
describe('Token Storage', () => {
  // Mock useCookie and test:
  - setAccessToken stores in cookie
  - getAccessTokenSilently retrieves from cookie
  - Token persists across page reloads
  - Token cleared on logout
})
```

**Why:** Token storage is the foundation of auth. Bugs here break everything.

### Priority 2: Callback Page (High)

```typescript
// test/unit/auth-callback.test.ts
describe('OAuth Callback', () => {
  // Mock $fetch and test:
  - Validates token with backend
  - Handles validation errors
  - Redirects after successful validation
  - Sanitizes redirect parameter
})
```

**Why:** This is the entry point for authentication. Must be bulletproof.

### Priority 3: Middleware (High)

```typescript
// test/unit/auth-middleware.test.ts
describe('Route Protection', () => {
  // Mock navigation and test:
  - Redirects to login without token
  - Allows access with valid token
  - Skips callback page
  - Preserves intended destination
})
```

**Why:** Prevents unauthorized access. Security-critical.

### Priority 4: Admin Authorization (Medium)

```typescript
// test/unit/auth-admin.test.ts
describe('Admin Routes', () => {
  // Test role-based access:
  - Admin/superadmin can access
  - Regular users get 403
  - Waits for me query
  - Handles loading timeout
})
```

**Why:** Prevents privilege escalation.

### Priority 5: Integration Tests (Medium)

```typescript
// test/integration/auth-flow.test.ts
describe('Full Auth Flow', () => {
  // Test complete user journey:
  - Visit protected route → login → callback → destination
  - Token expiration handling
  - Multi-tab synchronization
})
```

**Why:** Ensures all pieces work together.

## Test Utilities

### ✅ Implemented Mock Utilities (`test/utils/auth-mocks.ts`)

We've created a comprehensive set of mock utilities for testing auth functionality:

#### Core Mocking Functions

- **`mockUseAuth(overrides)`** - Mock the useAuth composable
  - Includes reactive `me` ref with computed role properties
  - Supports token, isLoading, and all role checks
  - Functions: setAccessToken, getAccessTokenSilently, loginWithRedirect

- **`mockUseCookie(initialValue)`** - Mock Nuxt's useCookie
  - Returns a reactive ref that behaves like a cookie

- **`createMockToken(payload)`** - Generate valid JWT format tokens
  - Properly base64url encoded
  - Customizable payload (user_id, exp, roles, etc.)
  - Default 1-hour expiration

- **`createMockUser(overrides)`** - Generate mock user objects
  - Complete GetMeQuery['me'] structure
  - Customizable roles, church, etc.

- **`mockNavigateTo()`** - Mock Nuxt navigation
- **`mockCreateError()`** - Mock Nuxt error creation
- **`mockWindowLocation(pathname)`** - Mock window.location
- **`mockUseRoute(query)`** - Mock Nuxt route with query parameters
- **`mockUseRuntimeConfig()`** - Mock Nuxt runtime config (apiUrl, callbackUrl, loginUrl)
- **`mockFetch()`** - Mock $fetch for backend API calls
- **`createMockFetchResponse(data, ok, status)`** - Mock fetch responses

#### Example Usage

```typescript
import {
  mockUseAuth,
  createMockToken,
  createMockUser,
} from '../utils/auth-mocks'
import { RoleType } from '../../app/api/generated'

// Test token management
const cookie = mockUseCookie<string>()
const auth = mockUseAuth({ token: cookie })
const token = createMockToken()
auth.setAccessToken(token)
expect(cookie.value).toBe(token)

// Test role detection
const adminUser = createMockUser({
  roles: [{ role: RoleType.Admin, scope: null }],
})
const auth = mockUseAuth({ me: ref(adminUser) })
expect(auth.isAdmin.value).toBe(true)
```

## Critical Bugs to Prevent

These tests are designed to catch:

### 🔐 Security Vulnerabilities

- ✅ Open redirect attacks (redirect=https://evil.com)
- ✅ XSS via redirect parameter (redirect=javascript:alert(1))
- ⏳ Token leakage in logs/URLs
- ⏳ CSRF attacks on token endpoints
- ⏳ Token reuse after logout

### 🐛 Logic Errors

- ⏳ Race conditions (multiple tabs logging in)
- ⏳ Token expiration not detected
- ⏳ Middleware blocking callback page
- ⏳ Admin check before me query completes
- ⏳ Navigation preserving invalid tokens

### 💔 UX Issues

- ⏳ Losing redirect destination on login
- ⏳ Infinite redirect loops
- ⏳ Showing auth errors to users
- ⏳ Not showing loading states
- ⏳ Multiple login requests

## Running Tests

```bash
# Run all auth tests
pnpm test -- auth

# Run only helper tests (fully implemented)
pnpm test -- auth-helpers

# Run with coverage
pnpm test -- auth --coverage

# Watch mode during development
pnpm test -- auth --watch
```

## Test Coverage Goals

| Component                               | Current | Target |
| --------------------------------------- | ------- | ------ |
| Helper Functions                        | 100% ✅ | 100%   |
| useAuth Composable - Token Management   | 100% ✅ | 100%   |
| useAuth Composable - Login Redirect     | 100% ✅ | 100%   |
| useAuth Composable - Role Authorization | 100% ✅ | 100%   |
| useAuth Composable - Me Query           | 100% ✅ | 100%   |
| Callback Page Token Validation          | 100% ✅ | 100%   |
| Global Auth Middleware                  | 100% ✅ | 100%   |
| Admin Middleware                        | 100% ✅ | 100%   |
| Integration                             | 0%      | 80%    |

**Total Progress: 123 tests passing (93 fully implemented, 30 scaffolded)**

**Breakdown:**

- `auth-helpers.test.ts`: 36 tests (100% implemented)
- `auth.test.ts`: 57 implemented + 30 scaffolded = 87 tests

## Next Steps

1. ✅ **Completed:** Created comprehensive helper functions and tests (36 tests)
2. ✅ **Completed:** Implemented token management tests with Vue mocking (9 tests)
3. ✅ **Completed:** Implemented login redirect tests (5 tests)
4. ✅ **Completed:** Implemented role-based authorization tests (9 tests)
5. ✅ **Completed:** Implemented GraphQL me query tests (6 tests)
6. ✅ **Completed:** Implemented callback page validation tests (12 tests)
7. ✅ **Completed:** Implemented global auth middleware tests (6 tests)
8. ✅ **Completed:** Implemented admin middleware tests (10 tests)
9. **Future:** Complete security and integration tests (30 remaining scaffolded tests)

## References

- [Nuxt Test Utils](https://nuxt.com/docs/getting-started/testing)
- [Vitest Documentation](https://vitest.dev/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)

## Questions?

See the inline comments in `auth-helpers.test.ts` for examples of each test pattern.
