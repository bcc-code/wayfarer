# BCC Members Integration

## Overview

This document describes the BCC Members API integration for fetching member information during user authentication.

## Members API Implementation Details

### 1. API Methods

The implementation provides 4 main API methods:

- **`Lookup(personID)`** - Fetches a single member by PersonID from `GET /v2/persons/{personID}`
- **`RetrieveByEmails(emails[])`** - Fetches multiple members by email with filtering
- **`GetMembersByIDs(ids[])`** - Batch fetches members with automatic chunking (800 IDs per request)
- **`GetOrganizationsByIDs(uuids[])`** - Batch fetches organizations with automatic chunking (800 UUIDs per request)

All requests are wrapped in a generic request handler that manages authentication, error handling, and circuit breaker execution.

### 2. Authentication & Authorization

Uses **Auth0 Client Credentials OAuth2 Flow**:
- Requests an access token from Auth0 using `client_id` and `client_secret`
- Tokens are cached per audience with 24-hour TTL
- Each Members API request includes `Authorization: Bearer {token}` header
- Token provider is injected via dependency injection pattern

### 3. Data Fields Available

**Member Fields:**
- PersonID
- BirthDate
- Email
- EmailVerified
- DisplayName
- FirstName
- Gender
- Affiliations array (organization relationships)

**Affiliation Fields:**
- Active
- OrgUid
- PersonUid
- Uid
- Type
- ValidFrom
- ValidTo

**Organization Fields:**
- OrgID
- Name (districtName)
- Type
- Uid

The API uses a `fields` parameter to specify what to fetch (supports `*` for all fields).

### 4. Configuration & Environment Variables

- **`MEMBERS_API_DOMAIN`** - Members API base domain (required)
- **`AUTH0_CLIENT_ID`**, **`AUTH0_CLIENT_SECRET`**, **`AUTH0_DOMAIN`** - Auth0 configuration (required for token generation)
- **`AUTH0_MANAGEMENT_AUDIENCE`** - Must include Members API domain

### 5. Go Packages & Dependencies

Key dependencies:
- `github.com/sony/gobreaker v0.5.0` - Circuit breaker
- `github.com/google/uuid v1.6.0` - UUID handling
- `github.com/ansel1/merry/v2 v2.2.1` - Error handling
- `github.com/samber/lo` - Chunking utilities
- `github.com/Code-Hex/go-generics-cache v1.3.1` - Token caching
- Standard library: `net/http`, `encoding/json`, `context`

### Key Implementation Details

- **Circuit Breaker**: 2-second timeout prevents cascading failures
- **HTTP Timeout**: 3 seconds per request
- **Chunking**: Automatically splits large requests into 800-item chunks
- **DataLoader Integration**: Uses graph-gophers/dataloader for efficient GraphQL batching
- **Error Handling**: Comprehensive error wrapping with HTTP codes via merry package
- **Response Format**: All responses wrapped in `{"data": {...}}` envelope

## Implementation Plan

1. Copy members client implementation from brunstadtv project
2. Change database schema: `age` → `birthdate` with calculated age
3. Add members API client to backend dependencies
4. Integrate with login flow to auto-populate user data
5. Update GraphQL schema and resolvers
6. Add configuration for Members API
