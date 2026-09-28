# Auth Sessions (Wayfarer-owned tokens)

Users sign in with BCC (login.bcc.no / Auth0) once. The backend exchanges
that token for a Wayfarer token pair and from then on renews it itself, with
no Auth0 involvement. Design page: https://claude.ai/artifact/9JjKez4ctAceW5VTsQR1u9

## Tokens

| | Access token | Refresh token |
|---|---|---|
| Format | HS256 JWT with a `kid` header | `wfr_<session id>_<32 random bytes, base64url>` (76 chars) |
| Lifetime | 7 days (`JWT_ACCESS_TOKEN_TTL`) | ~6 months, sliding (`JWT_REFRESH_TOKEN_TTL`) |
| Claims / storage | `user_id`, `user_roles`, `sid`, `sub`, `iss`, `aud=wayfarer`, `iat`, `exp` | SHA-256 hash in `auth_sessions` |
| Checked | Every request, in memory (`internal/authtoken`) | Only by `POST /auth/refresh` |

Verification costs about 5 µs and does no I/O (see `BenchmarkParse` in `internal/authtoken`).

## Flow

1. The frontend gets a BCC token and calls `POST /auth/exchange {token}`.
   - `AuthHandler.authenticateExternalToken` validates it: RS256 only, issuer, and audience when `AUTH0_AUDIENCE` is set.
   - It then finds or creates the user.
   - `AuthSessionService.StartSession` inserts an `auth_sessions` row.
2. Requests send `Authorization: Bearer <access_token>`. `middleware.JWTAuth` puts `user_id`, `user_roles` and `session_id` in the context.
3. `POST /auth/refresh {refresh_token}` loads the session by the ID embedded in the token (primary key).
   - It then loads roles and signs the new access token **before** committing anything.
   - Last comes a compare-and-swap `RotateAuthSession` (`WHERE id = … AND refresh_token_hash = <presented>`). In one statement (a CTE), the swap stores the new hash, keeps the old one in `prev_refresh_token_hash`, records it in `auth_session_retired_tokens`, and moves `expires_at` to now + TTL.
   - A failure before the swap leaves the presented token valid, so the client just retries.
   - If the presented token is the one rotated out within the last 30 s, or the swap loses a concurrent race, the response is **409 `token_rotated`**: another tab won, so the client reloads the stored tokens.
   - A token whose hash is in `auth_session_retired_tokens` for that session (e.g. A after an attacker rotated A→B→C) **revokes the session** and returns **401**.
   - Any other token naming the session is forged and gets **401 with no side effects**. The session ID is not secret (it is also the access token's `sid` claim), so it is never trusted on its own.
4. `POST /auth/logout {refresh_token}` revokes the session if the token is its current one or one of its retired ones, so a token rotated moments ago still works. Forged or unknown tokens are ignored (204 either way, idempotent).

The `/auth` group caps request bodies at 16 KiB (`middleware.MaxBodyBytes`, 413 when exceeded). The BCC token may be at most 8 KiB and the refresh token at most 128 chars; malformed refresh tokens are rejected without a DB query.

On the frontend, refresh and logout share one Web Lock. Logout waits for a pending refresh before clearing tokens. A refresh response that arrives after the stored session changed (cleared or replaced) is discarded. Proactive refresh starts once the access token is older than a day or half its lifetime, whichever is sooner.

`GET /token` (legacy, 24h JWT without a session) still works during the transition.

## Revocation

- **Session revocation:** `Logout` and `RevokeUserSessions` kill the refresh token. The access token already issued stays valid until it expires, at most 7 days.
- **Emergency list:** `AUTH_REVOKED_USERS=USxxx[:RFC3339 cutoff],...`, normally empty and parsed at startup into `JWTConfig.RevokedUsers`.
  - Without a cutoff, every token and refresh for that user is rejected. With a cutoff, tokens issued before it and sessions created before it are rejected, which forces a new BCC login.
  - The per-request cost is one map lookup.
  - Changes need a restart.
  - An unparseable cutoff blocks the user entirely.
- Roles are embedded in the token and refreshed on each refresh. `@requireRole` re-reads DB roles, but some resolvers (scoring, teams, upload, plugins) trust the token's roles. A demoted user can therefore keep elevated roles in those paths until the next refresh (≤7 days). Use the emergency list if that matters.

## Secret rotation

Set `JWT_SECRET_PREVIOUS` to the old secret and `JWT_SECRET` to the new one. Tokens are verified with the secret their `kid` names; tokens without a `kid` (legacy, m2m) use `JWT_SECRET`. Once 7 days have passed, drop `JWT_SECRET_PREVIOUS`.

m2m tokens have no `kid`, so they have to be re-issued when the secret rotates.

## Maintenance

`POST /api/maintenance/cleanup-auth-sessions` (API key) deletes sessions that expired or were revoked more than 30 days ago. It is scheduled weekly as the ansible job `cleanup-auth-sessions`.

## Code

- `backend/internal/authtoken/`: sign/parse, kid, audience, emergency check
- `backend/internal/services/auth_session.go`: sessions, rotation, reuse detection
- `backend/internal/handlers/auth_session.go`: `/auth/*` endpoints
- `backend/internal/database/queries/auth_sessions.sql`, migration `00105_add_auth_sessions.sql` (tables `auth_sessions`, `auth_session_retired_tokens`)
- `frontend/app/composables/useAuth.ts` + `frontend/app/utils/authSession.ts`: storage, cross-tab refresh
