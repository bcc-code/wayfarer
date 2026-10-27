# Audit: PR #227 → #232 → #234

**Scope:** the stack originally audited at `origin/main...410f6d16`, plus adjacent code it depends on.
**Date:** 2026-10-22, re-verified 2026-10-27 against `origin/main` `5d901c02`. Resolved findings have been removed, and paths and line numbers match that tree. A2 and A3 were fixed afterwards (#252, #253) and removed on 2026-10-27.

**Open:**

- **The stack itself:** 2 medium (M1, M2), 3 low (L3, L4, L5), 2 nits (N2, N3).
- **Adjacent code:** 7 medium (A6–A12), 11 low (A13–A23), 3 nits. A14 has a fix in review (#255).

## Findings

| ID | Sev | PR | Where | Finding | Fix |
|----|-----|----|-------|---------|-----|
| M1 | Medium | #234 | `frontend/app/composables/useFirestoreRefresh.ts:59` | `await refetch()` has no upper bound. urql's `then` resolves only once `fetching` and `stale` are both false, and `fetchExchange` has no timeout. If a request hangs (half-open mobile connection), `inFlight` stays `true`, so every later ping only sets `queued` and the page stops live-updating until the browser kills the request (minutes) or the page remounts. No test covers a refetch that never settles. | `Promise.race` the refetch against a timeout (e.g. 15s), then clear `inFlight` and run the queued refetch. Add a test with a promise that never resolves. |
| M2 | Medium | #232 | `backend/internal/loaders/loaders.go:116` (default `:79`), `backend/internal/config/config.go:260` | `context.WithoutCancel` means a client that disconnects no longer stops its batch's DB work. A client can send expensive queries and drop the connection, and each batch still holds a pool connection for up to `DATALOADER_BATCH_TIMEOUT` (10s). Before, the first caller disconnecting cancelled the batch. Pool is 25 conns, and no `statement_timeout` is set anywhere in the repo (app code, migrations, `docker/`). This is a deliberate tradeoff, but it makes DB pool exhaustion cheaper for an abusive client. `notes/02-configuration.md` explains the timeout but not the pool-exhaustion tradeoff. | Accept it and document it, or lower the default (2–5s) and/or set a Postgres `statement_timeout` for the app role. Rate limiting at the edge covers most of it. |
| L3 | Low | #234 | `frontend/app/composables/useFirestoreRefresh.ts:48,57` | Spacing uses `Date.now()`. If the device clock jumps backwards, `lastStart + minIntervalMs - Date.now()` can delay the next refetch for as long as the jump. | Use `performance.now()`. |
| L4 | Low | #227 | `backend/cmd/loadtest/bench/betting_box.sh:278-281` | `webhook_url` goes into SQL unescaped (`'$webhook_url'`). The operator supplies it, so there is no remote attack path, but a quote in the URL breaks the statement. `reset-multi` also rewrites real `users.church_id` (restorable from `loadtest_church_backup`). It is only safe on the LT box, and nothing checks the target DB. | Pass it as `psql -v` with `:'webhook_url'`. Optionally refuse to run unless the DB has a load-test marker. |
| L5 | Low | #227, #234 | `backend/scripts/loadtest-lt.sh:25,294`, `backend/cmd/loadtest/README.md:216`, `backend/cmd/loadtest/k6/betting-multisession.js:14,100-101`, `frontend/app/composables/useFirestoreRefresh.ts:20` | Merged code points to `notes/multi-session-loadtest-plan.md` (§4, §5) and to risk ID "R4" defined there. The plan was never committed, so the references dangle. Worse, `notes/betting-edge-cases.md:52` now defines a **different** R4 ("Answers and bets only while the session is OPEN"), so the frontend and k6 references point to the wrong risk. | Commit the plan, or inline the needed context and drop the references. At minimum rename the coalescing risk so it doesn't clash with the betting R4. |
| N2 | Nit | #227 | `backend/Makefile:43-44` | `test-bugrepro` target exists, but no file in the tree carries `//go:build bugrepro`, and `notes/09-unit-tests.md` doesn't document it. The target is undocumented and runs nothing extra. | Document it, or drop the target. |
| N3 | Nit | #232 | `notes/02-configuration.md:33` | `DATALOADER_BATCH_TIMEOUT` is listed between `DB_MAX_OPEN_CONNS` and `DB_MAX_IDLE_CONNS`. | Move it after the `DB_*` entries. |

## Checked, no finding

- **tokengen `-church-prefix`:** parameterised queries (`pgx.NamedArgs`). Church admins get a plain `user` token and are authorized by their `CHURCH_ADMIN` DB row, so the tool grants no extra privileges. `config.json` and `results/` are gitignored. k6 logs don't print tokens.
- **`church_sizes`:** awk validates the bucket input and re-emits it through `%d`, so the unquoted `$sizes` heredoc only ever gets integers.
- **`DELETE FROM churches LIKE 'CH01LOADTESTCHURCH%'`:** `users.church_id` is `ON DELETE RESTRICT`, so a user who wasn't restored makes the reset fail instead of being deleted.
- **Coalescer logout path:** `isAuthReady` watcher + `enabled()` at run time + `dispose` on unmount. There are tests for each of these.
- **Page callbacks:** they now return the `executeQuery` result. The completed-tab callback returns `undefined` when the tab is hidden, which resolves immediately.

## Adjacent code (not changed by the stack)

These are code paths the stack relies on or that the load test drives: shared loaders, Firestore pings, the response cache, quiz-session admin mutations, plugin webhook receivers and the load-test auth. They belong in follow-up PRs, not this stack.

Plugin webhook receivers live in `internal/plugins/ssf` and `internal/plugins/teams`. Both also serve the old `/plugins/ladder-to-heaven/...` paths as aliases.

Paths are relative to `backend/` unless they start with `frontend/`.

**Verified** column:
- ✔ means the claim was checked against the code for this audit.
- "rep." means it was found by a read-only sub-review and has not been re-checked line by line. Confirm it before acting.

### Medium

| ID | Verified | Where | Finding | Fix |
|----|----------|-------|---------|-----|
| A6 | ✔ | `internal/graph/api/quiz_sessions.go:26-70`, `internal/services/bulk/service.go:751-820` (rep.) | **Grant targets are not scoped.**<br>• Grant access checks only "creator or project manager".<br>• `userIds`, teams, super-teams and `allProjectUsers` are accepted unscoped.<br>• Church IDs are filtered to the session's project, but not to the admin's own church. | For church admins, allow only their own `churchId`, and filter targets to members of the session's project. |
| A7 | ✔ | `internal/graph/api/quiz_sessions.resolvers.go:50`, `internal/services/roles.go:329-349` | **`createQuizSession` is not scoped.** It uses `CanCreateTeamInProject`, which is true for a church admin of any church in every project. Combined with A6, that admin can open a session to a whole foreign project. | Require project membership or a project-scoped role. |
| A8 | ✔ | Open/Lock/Reopen: `internal/graph/api/quiz_sessions.resolvers.go:254-260,317-323,508-514`, `internal/handlers/quiz_scheduler.go:114-127,174-187`. Challenges: `internal/graph/api/challenges.resolvers.go:319-322,423-429`. Cache: `internal/cache/invalidation.go:515-541,618-629` | **A ping-driven refetch can get a stale cached response.**<br>• Open, lock, reopen and the challenge pings invalidate only entity caches, so the per-user 30s response cache can serve the old state to the refetch #234 coalesces. | Drop response entries before pinging: fold `DeletePrefix(PrefixGQLResponse)` into `InvalidateQuizSessionAccess` / `InvalidateChallenge`. |
| A9 | ✔ | `cmd/server/main.go:158-175,635-636`, `internal/config/config.go:103-108,310` | **Nothing stops the load-test key being set in production.** When `AUTH0_LOADTEST_PRIVATE_KEY` is set, the server serves `/jwks.json` and tolerates JWKS fetch failures. "Never set in production" is only a comment. | Refuse to start when `ENVIRONMENT=production` (or Cloud Run) and the key is set. |
| A10 | ✔ | `internal/config/config.go:304`, `internal/handlers/auth.go:315-317` | **The Auth0 audience is not checked by default.** `AUTH0_AUDIENCE` defaults to `""`, which skips the `aud` check, so a login.bcc.no token issued to **any** BCC app can be exchanged for a Wayfarer JWT. **Check the prod env.** | Set it in prod, and fail startup in production when it's empty. |
| A11 | rep. | `frontend/app/composables/useFirestoreSync.ts:74-91,268-273,317-326` | **One failed Firebase sign-in means no live updates until a layout remounts.** A failed `authenticate()` is never retried, and the watcher only re-initialises on a null → id change of `me`. | Retry with backoff while `me` is set and not initialised, and again on `online` / `visibilitychange`. |
| A12 | rep. | (Firebase console) | **Firestore rules are not in the repo.** They can't be reviewed or versioned. If client writes are allowed, one user writing `projects/{id}/notifications/*` in a loop makes every client refetch. #234 caps that at about 0.5 req/s per page per client. | Commit `firestore.rules` with `allow write: if false` everywhere and `users/{u}` readable only by `u`. Add a role claim before relying on rules for `admin/*`. |

### Low

| ID | Verified | Where | Finding | Fix |
|----|----------|-------|---------|-----|
| A13 | rep. | `internal/graph/api/response_cache.go:215-218` | **An in-flight request can write a stale response back into the cache after invalidation.** There is no generation check, so a stale entry can live for 30s during open→lock→finish bursts. | Capture the generation before `next(ctx)` and skip `Set` if it moved. |
| A14 | ✔ | `frontend/layers/user/app/pages/challenges/index.vue:36` | **The Completed tab never updates live.** It listens for `CompletedChallengesPageDocument`, which no map in `useFirestoreSync.ts:11-37` dispatches. | Fix in review: #255 dispatches it wherever `ActiveChallengesPageDocument` is dispatched, and adds a test that every listened-for document is dispatched. |
| A15 | rep. | `frontend/app/composables/useFirestoreSync.ts:142-144,176-178,211-213` | **A listener that errors is never resubscribed.** Its key stays in `listeners`, so later subscribe calls skip it. | Delete the key in the error callback and resubscribe. |
| A16 | rep. | `internal/graph/api/achievements.resolvers.go:763,1319,1325,1710,1759`, `internal/graph/api/challenges.resolvers.go:804,811,871,1011`, `internal/graph/api/quizzes.resolvers.go:1949`, `internal/services/bulk/service.go:417,425,494,565,678,1058,1110`, `internal/plugins/teams/team_name_changed_handler.go:195` | **Firestore and push notifications start one goroutine per user, unbounded.** Bet-result notifications are already capped at 16 (`quiz_session_betting.go:301`); every site listed here is unbounded. | Route all fan-out through one bounded worker, like `services/push` (`pushSendConcurrency`). |
| A17 | rep. | `cmd/server/main.go:624` (only `/auth` limited), `/graphql` at `:755`; `internal/plugins/ssf/content_event_handler.go:76`, `internal/plugins/teams/team_name_changed_handler.go:63` | **No request body size limits.** `/graphql` and the public plugin routes read bodies of any size, and the plugin handlers `io.ReadAll` before the signature check. `userIds` lists are uncapped. | Use `http.MaxBytesReader` on those routes and cap list sizes. |
| A18 | rep. | `internal/graph/api/webhooks.resolvers.go:36-46,59-68`, `internal/services/webhooks/service.go:30-32,195,337` | **Webhook SSRF (superadmin only).** URLs aren't validated, redirects are followed, and up to 10KB of the response is stored and readable via `recentLogs` / `testWebhook`. | https only; block private, loopback and link-local IPs at dial time; refuse redirects. |
| A19 | rep. | `internal/webhook/signature.go:12-29` | **No replay protection.** The timestamp is never checked. Both receivers return 409 on a duplicate score journal entry, so a replay only causes repeated work. | Reject timestamps older than about 5 minutes, and add a delivery ID. |
| A20 | rep. | `internal/services/webhooks/service.go:222-245`, `internal/database/queries/webhooks.sql:71-100` | **PII is kept indefinitely in `webhook_logs`.** Payloads include email, name and `members_id`, and there is no retention or delete query. | Add a retention job and redact user data. |
| A21 | rep. | `internal/graph/api/bulk_jobs.resolvers.go:32-44`, `internal/services/bulk/service.go:76-89,136-142` | **`bulkJob(id)` has no ownership check.** Any authenticated user can read any job's status and counts. Related: `retryBulkJob` re-runs any job's stored input with no ownership or project check, and is gated only by an admin role. | Restrict to the creator or a project manager. |
| A22 | rep. | `internal/graph/api/quiz_sessions.go:25-70`, `internal/services/bulk/service.go:847-895`, `internal/services/bulk/processor.go:129-143` | **Grant jobs accept any session state and hide partial failures.** One bad `userId` fails a 500-row batch, but the job is marked COMPLETED. `failureCount` is stored. Whether grants into non-OPEN sessions are wrong is a product question. | Validate the session at execution time, and filter IDs through a join. |
| A23 | rep. | `internal/firebase/service.go`, `internal/handlers/auth_session.go:111` | **Firebase refresh tokens are never revoked.** `RevokeRefreshTokens` is called nowhere, including on logout and session cleanup. Impact is small, because the docs hold only a timestamp. | Call `RevokeRefreshTokens(uid)` on user deletion or session revoke. |

### Nits

- `NotifyProjectEvents` writes `projects/{id}/notifications/events`, which no client subscribes to (`internal/graph/api/events.resolvers.go:164,234`).
- The admin feedback page discards its `subscribeAdmin` cleanup, so the listener lives until logout (`frontend/layers/admin/app/pages/admin/feedback/index.vue:11-20`).
- `firebaseToken.expiresIn` always reports 3600 for cached tokens. Tokens are cached for 30 min (`schema.resolvers.go:89-93`) or 55 min (`services/firebase_token_warmer.go:21,269-272`).
