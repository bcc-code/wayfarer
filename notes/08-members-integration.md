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

## Keeping Wayfarer Users in Sync with Members Data

New user creation only ever happens reactively: `AuthHandler.findOrCreateUser`
(`internal/handlers/auth.go`) creates a Wayfarer `users` row the first time a
person authenticates via Auth0. Nothing polls or watches the Members API for
brand-new people ahead of their first login — the Members API client has no
"list/filter by created-at" method to support that, and no webhook exists for
"member created" (the two existing webhooks, `content-events` and
`consent-events`, both require an *existing* `members_id` in Wayfarer). This is
a deliberate decision, not a gap to fix by default — revisit only if proactive
account creation becomes an actual requirement.

For users that already exist, three entry points re-pull data from the
Members API. As of the changes below, the two maintenance endpoints share one
implementation (`MaintenanceHandler.syncUserProfile`) and apply the fully
extended field set; the GraphQL mutation is still limited to
gender/church/person_uuid and is a candidate for the same treatment later:

| Entry point | Trigger | Fields synced |
|---|---|---|
| `POST /api/maintenance/sync-user-data?limit=N` (`MaintenanceHandler.SyncUserData`) | Manual/API-key, batch, oldest-`updated_at`-first (`GetUsersLeastRecentlySynced`) | email, name, first/last/middle/display name, gender, birthdate, church_id, person_uuid |
| `POST /api/maintenance/sync-user/:user_id` (`MaintenanceHandler.SyncSingleUser`) | Manual/API-key, one user | same, always for the given user |
| GraphQL `syncUser(userId)` (`UserSyncService.SyncUser`) | Admin/superadmin GraphQL call | gender, church_id, person_uuid + SSF content-event backfill |

`SyncUserData` used to select candidates via `WHERE gender = 'UNKNOWN'` —
meaning once a user's gender was known, that user could never be touched
again, even with a stale name/birthdate. It now selects the `limit`
least-recently-`updated_at` users, so repeated calls (e.g. from a cron job)
cycle through the *entire* user base over time. Caveat: `updated_at` is a
general "last modified" column, not a dedicated "last-synced-from-Members"
timestamp — any other write to a user (e.g. an admin edit) also bumps it and
pushes that user to the back of the sync queue. Acceptable for now; a
dedicated `last_synced_at` column would need a migration if this imprecision
becomes a real problem.

The query also filters to `WHERE members_id ~ '^[0-9]+$'`. About half the
rows in the dev DB have a `members_id` that was never a real Members API
person ID — `MEM-<n>` placeholders from `cmd/seed/seeders/users.go`, plus at
least one UUID-shaped one from `cmd/debug`. `Lookup` can't resolve any of
these (the Members API only has integer person IDs), and since a failed
sync never reaches the `UPDATE` that bumps `updated_at`, unfiltered rows like
this would wedge permanently at the front of the least-recently-synced queue
— every batch call would burn its whole `limit` on the same unsyncable rows
forever, and users with real numeric IDs would never get reached.

No cron/scheduler exists anywhere in this repo for `sync-user-data` (checked:
no Terraform/k8s manifests, no scheduled GitHub Actions, no in-process
ticker). It's triggered externally, from outside this codebase.

Shared logic lives in `internal/members/profile.go`:
- `ExtractProfile(member *Member) ProfileFields` — computes email, first/last/middle/display
  name, computed `name`, and a validated `*time.Time` birthdate from a Member record.
  An empty string / nil field means "Members API had no value" — callers should leave
  the existing DB value alone rather than blank it.
- `GenerateDisplayName` / `ParseBirthdate` — same rules used at account creation
  (`auth.go`) and at sync time, so a synced profile can't drift from a freshly
  created one.

The corresponding SQL, `UpdateUserProfileFromMembers`
(`internal/database/queries/users.sql`), uses
`COALESCE(NULLIF(@x::text, ''), x)` (and the date equivalent for birthdate) per
column — an empty/absent value from the Members API is a no-op, never a
blank-out. The older `UpdateUserGenderAndChurch` query is kept for now since
`UserSyncService.syncMemberData` (used by the GraphQL `syncUser` mutation)
still uses it.
