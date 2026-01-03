# Wayfarer API Endpoints

## Health & Monitoring

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check (returns `{"status": "ok"}`) |
| `/metrics/cache` | GET | Cache metrics (hit/miss stats, keys) |
| `/debug/pprof/*` | GET | Go profiling endpoints |

## Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/token` | GET | OAuth callback - exchanges external JWT for Wayfarer JWT |

Query param: `token` (JWT from Brunstad TV or Auth0)

**Env vars:**
- `JWT_SECRET` - HMAC secret for signing Wayfarer tokens
- `JWT_ISSUER` - JWT issuer claim
- `BRUNSTAD_TV_JWKS_URL` - Brunstad TV JWKS endpoint
- `BRUNSTAD_TV_JWT_ISSUER` - Brunstad TV expected issuer
- `AUTH0_JWKS_URL` - Auth0 JWKS endpoint
- `AUTH0_JWT_ISSUER` - Auth0 issuer
- `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET` - Auth0 credentials

## GraphQL API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/graphql` | POST | GraphQL API endpoint |
| `/graphql` | GET | GraphQL Playground (non-production only) |

Requires JWT Bearer authentication.

**Env vars:**
- `JWT_SECRET` - For validating incoming JWTs
- `JWT_ISSUER` - Expected issuer

## File Upload

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/upload` | POST | Upload files to S3 |

Requires JWT + admin role. Accepts multipart form data (max 30MB).

**Env vars:**
- `AWS_S3_BUCKET` - S3 bucket name
- `AWS_S3_REGION` - AWS region
- `AWS_S3_ACCESS_KEY_ID`, `AWS_S3_SECRET_ACCESS_KEY` - AWS credentials (optional on Cloud Run)
- `AWS_S3_PUBLIC_BASE_URL` - Public base URL for files
- `AWS_S3_ROLE_ARN` - OIDC role ARN for Cloud Run

## Webhooks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/content-events` | POST | External content event webhook |
| `/api/v1/consent-events` | POST | External consent event webhook |

Requires API key via Bearer token.

**Env vars:**
- `EXTERNAL_API_KEYS` - Format: `source1:key1,source2:key2`

## Maintenance

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/maintenance/sync-user-data` | POST | Batch sync user data from Members API |
| `/api/maintenance/sync-user/:user_id` | POST | Sync single user by ID |

Requires API key via Bearer token.

**Env vars:**
- `EXTERNAL_API_KEYS`
- `MEMBERS_API_DOMAIN` - BCC Members API domain

## SSF Integration

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ssf/sync/:slug` | POST | Sync Bible study plan from SSF API |

Auth: `X-Sync-Key` header or `?key=` query param.

**Env vars:**
- `SSF_API_BASE_URL` - SSF API endpoint
- `SSF_API_KEY` - Bearer token for SSF API
- `SSF_SYNC_KEY` - Static key for sync endpoint

## Translations (Phrase TMS)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/translations/webhook` | POST | Phrase TMS webhook (no auth) |
| `/api/translations/export/:collection` | POST | Export collection to Phrase |
| `/api/translations/export/all` | POST | Export all collections to Phrase |

Export endpoints require `X-Export-Key` header or `Authorization: Bearer <key>`.

**Env vars:**
- `PHRASE_ENABLED` - Enable Phrase integration
- `PHRASE_BASE_URL` - Phrase API endpoint
- `PHRASE_USERNAME`, `PHRASE_PASSWORD` - Phrase credentials
- `PHRASE_PROJECT_UID` - Phrase project UUID
- `TRANSLATIONS_EXPORT_KEY` - Static key for export endpoints

## Static Files

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/`, `/_nuxt/*`, `/_fonts/*`, `/images/*` | GET | Static file serving |
| `/{unmatched}` | GET | SPA fallback (serves index.html) |

Only active if configured.

**Env vars:**
- `STATIC_FILES_PATH` - Path to frontend build files
