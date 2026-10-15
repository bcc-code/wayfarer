# Required (blocking) consents

**Status:** Phase 1 (the column) done — everything else not started
**Started:** 2026-10-14
**Branch:** feature/required-consents

**Scope call, 2026-10-15:** ship only the column and plumb it to the frontend,
so frontend work can start in parallel. All behaviour — the block, the
participation filter, the reject guard — is left to a backend engineer and is
unchanged below.

## Goal

A consent can be marked **required**. A user who has not accepted a required
consent that applies to the current project **is not a participant in that
project**:

- they cannot use the app — every route sends them to the consent screen, and
  that screen offers no way out except accepting;
- they are not visible in their unit, even when an admin has placed them in one;
- they are not visible in any leaderboard;
- they receive no notifications about the project.

Rejecting a required consent is not possible. Accepting restores everything —
nothing is torn down in the data, so nothing has to be rebuilt.

Today every consent is optional: the user may reject it and carry on, and the
only gating is a soft redirect from the front page during onboarding.

## What exists today (findings)

### Database

`consents` (migrations `00027`, `00028`, `00043`, `00107`):

| Column                     | Meaning                                                   |
| -------------------------- | --------------------------------------------------------- |
| `key` + `version`          | A consent is a key; versions bump per key                 |
| `published_at`             | NULL = draft; status always looks at the latest published |
| `is_remote` / `managed_by` | Accepted in an external system (SSF), not in-app          |
| `project_id` (nullable)    | NULL = applies to every project (`00107`)                 |

`user_consent_history` is append-only, keyed by `consent_key` (not consent id)
so a version bump keeps the user's answer. `action` is `ACCEPTED` | `REJECTED`,
constrained by `check_action`.

### Where "pending" is computed — two independent paths

1. **`ConsentStatus` on `User`** — `backend/internal/loaders/user.go`
   - `fetchConsentDataForUsers` loads `GetLatestPublishedConsentsForProject`
     (cached per project under `cache.LatestConsentsKey(projectID)`) plus
     `GetCurrentUserConsentStatusesByUsers`.
   - `buildConsentStatus` splits into pending / accepted / rejected.
     **Pending means "no history row for the key"** — a rejected consent is
     never pending again.
   - It also **skips LOCAL consents for users under 13** (`calculateAge`).
2. **`Query.pendingConsents`** — `consents.resolvers.go` →
   `GetMissingConsentsForUserWithRejections`, same "no history at all" rule,
   project-scoped. (Despite the name, it does _not_ re-surface rejections.)

Both paths need the new flag; the frontend only uses path 1
(`me { consentStatus { ... } }`).

### Frontend gating today

- `layers/user/app/pages/index.vue:107` — watches `pendingConsents` filtered to
  `ConsentManagementType.Local`, sets `hasCompletedOnboarding = false` and
  `navigateTo({ name: 'settings-consent' })`. **Only the front page does this**,
  so any other route is reachable while consents are pending.
- `layers/user/app/pages/settings/consent.vue` — the onboarding screen. Hides
  the close button and shows a "Gå videre" button gated on
  `hasCompletedOnboarding` (a `useLocalStorage` flag) and
  `hasPendingInternalConsents`.
- `ConsentCard.vue` — accept/reject buttons, remote consents link out instead.
- Root middleware lives in `frontend/app/middleware/` (`01.auth.global.ts`,
  `02.admin-permission.global.ts`). **A layer's `*.global.ts` runs before the
  root ones** (see the comment in `layers/user/app/layouts/default.vue`), so a
  new gate belongs in `frontend/app/middleware/03.*.global.ts`.

### Admin

`layers/admin/app/pages/admin/consents/{index,new,[consentId]}.vue`. Note the
create form does not expose `projectId` even though the mutation accepts it.

## The consent model (from the 2026-10-15 flowchart)

Two consents, doing two different jobs:

```
"Samtykker til å delta i prosjektet"  (PÅKREVD)
        │ Ja                           ← no "Nei" branch: required
        ▼
"Samtykker til at poeng vises i unit og i topplister"   (valgfri)
    ┌───┴────┐
   Nei       Ja
    │         └─► synlig i unit OG topplister, poeng synlige for alle
    └─► synlig i sin unit, men poeng skjult; IKKE synlig i topplister
```

**"Unit" is a Team.** Confirmed in `layers/user/app/pages/standings.vue:14` —
the unit tab keys off `myCurrentProject.myTeam` and renders
`myTeam.memberLeaderboard`. "Topplister" are the config-driven boards in
`myCurrentProject.leaderboards`.

Consent 1 is what this note builds: `required`, plus the block. **Consent 2 is
parked** (decided 2026-10-15) — it needs a mechanism that does not exist yet,
and it is not on the critical path. They share only the consent screen.

### Parked: consent 2 (points visibility)

Not in scope now. Captured so the analysis is not redone:

- It needs **consent effects** — a consent whose answer changes system
  behaviour. Today a consent is a key, text and an accept/reject history, with
  no link to behaviour. Likely shape: a nullable `consents.effect` enum
  (`POINTS_VISIBILITY`), set by the admin, rather than a hard-coded magic key
  that an admin can rename.
- **`LeaderboardEntry.score` is `Int!`** (`gql/shared.graphqls:147`), so
  "visible in unit, points hidden" has no representation. Cheapest fix is an
  added `scoreVisible: Boolean!` rather than making `score` nullable, which
  would break every existing client.
- Two surfaces, different treatments: hidden from the config-driven boards
  entirely, but still listed in `Team.memberLeaderboard` with the score masked.
- Hidden users' points must **still feed** team / super-team / church
  aggregates. Otherwise rejecting drags your team's total down, and a consent
  you are punished for refusing is not freely given. **This is the opposite of
  decision #9** — and correctly so: consent 2 is refused by a _participant_,
  who keeps counting; a required consent is unanswered by someone who is not a
  participant at all.
- Pagination counts (`CountProjectPersonLeaderboard` et al.) must use the same
  filter as the row queries, or paging breaks.
- Open, if it is ever picked up: what is the default before a user answers, and
  does the under-13 LOCAL-consent exemption (decision #3) leave minors
  permanently hidden from leaderboards?

## Design decisions

| #   | Question                                                          | Decision                                                                                                                                                                                                                                                                                                                                  |
| --- | ----------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Is `required` per key or per version?                             | **Per version**, like `title`/`body`: the latest published version wins, so a consent can be made required by publishing a new version. Unlike `project_id` (which is inherited and locked by `consentProjectForNewVersion`), there is no reason to freeze it — widening it is the whole point.                                           |
| 2   | What counts as blocking?                                          | Latest published consent for the current project, `required = true`, and the user's **latest action for that key is not `ACCEPTED`** — i.e. no history row _or_ a `REJECTED` row. This is a new rule; "pending" elsewhere still means "no history".                                                                                       |
| 3   | Users under 13                                                    | Keep the existing skip: LOCAL consents are not shown to under-13s, so a required local consent does **not** block them. Changing that would lock out every under-13 user with no path forward.                                                                                                                                            |
| 4   | Required + remote                                                 | **Reject at the API**: `createConsent`/`updateConsent` error if `required && isRemote`. A remote consent can only be accepted in the external system, so requiring one would strand the user behind a screen with no accept button. Revisit if SSF consent ever needs to be mandatory — that needs a "re-check" flow first.               |
| 5   | Rejecting a required consent                                      | `rejectConsent` errors. The reject button is hidden client-side too, but the API is the authority.                                                                                                                                                                                                                                        |
| 6   | Server-side enforcement                                           | **In scope (decided 2026-10-15,** reversing the 2026-10-14 "UI only" call**).** Driving case: we must not push notifications to someone who has not accepted. Note this needs **two** mechanisms, not one — an inbound request guard and an outbound audience filter. See phase 2.                                                        |
| 7   | What does "blocked" mean elsewhere?                               | **Not a participant, therefore invisible.** Clarified 2026-10-15: a user who has not accepted a required consent "cannot participate in the project" — so they are **not visible in their unit** (even when placed in one) and **not visible in any leaderboard**, on top of receiving no notifications. One rule, applied at every read. |
| 8   | Does membership data get removed?                                 | **No.** The `user_projects` / `team_members` rows stay — a blocked user can still be _placed_ in a unit by an admin, they just do not show. Filtering at read time means accepting restores them instantly, with no re-join and no lost team assignment.                                                                                  |
| 9   | Do a blocked user's points still count toward their unit's total? | **No — confirmed 2026-10-15.** Their score is excluded from the unit aggregate, and from the super-team and church aggregates. This is the _opposite_ of the rule for consent 2 (parked), deliberately — see below.                                                                                                                       |
| 10  | Invisible to whom?                                                | **To other users, not to admins.** Confirmed 2026-10-15: a blocked user "should basically not exist in the project for other users, until they accept". Admin surfaces keep showing them — the admin user list, `AdminUserConsents.vue`, unit management — or a blocked user becomes impossible to support.                               |

## Plan

### Phase 1 — backend data + API ✅ done 2026-10-15

Shipped exactly the field, no behaviour. `required` defaults to `false`, so
every existing consent behaves as it did.

- [x] Migration `00109_add_consent_is_required.sql`. **Was 00108, renumbered:**
      `main` stops at 107, but the shared dev database had 108 applied on
      2026-10-08 from the unmerged quiz answer-description branch. goose keys
      off the version number, so a second 00108 is silently treated as already
      applied — `make migrate` reports success and the column never appears.
      Check every branch, not just `main`, before picking a number here.
- [x] `is_required` added to every consent select, to `CreateConsent`'s insert and
      to `UpdateConsent`. Update uses `COALESCE(sqlc.narg('is_required'), is_required)`
      — a bool has no "empty string means keep" sentinel like the text fields
      next to it, so omitting it is a real no-op rather than a silent `false`.
- [x] `gql/consents.graphqls`: `Consent.isRequired: Boolean!`, plus an `isRequired`
      arg on `createConsent` and `updateConsent`.
- [x] `IsRequired` carried through all six `model.Consent` construction sites
      (4 in `consents.resolvers.go`, `loaders/user.go`, `loaders/consent.go`).
- [x] `make generate`, `make fmt`, `make test` — all green.
- [x] `e2e/consents_required_test.go` — defaults to false; round-trips; an
      update that omits it keeps the stored value; an update can change it.
      4/4 passing.
- [x] `pnpm codegen` — `Consent.isRequired` and both mutation args are in
      `frontend/app/api/generated.ts`.

Named `is_required` / `isRequired` (2026-10-15) rather than `required`, to
avoid a bare `required` colliding across the stack.

**Not done, deliberately:** `blockingConsents`, the reject-a-required-consent
guard, and the `is_required && isRemote` validation (decisions #4 and #5). The
flag is currently just a label — it blocks nothing.

### Phase 1b — the rest of the backend API (not started)

- [x] ~~Migration~~ — done, see above. **Still needs approval to run against a
      real database.**
- [ ] `backend/internal/database/queries/consents.sql`: add `required` to the
      select list of every `consents` query (`GetConsentByID`,
      `GetConsentsByIDs`, `GetLatestPublishedConsentByKey`,
      `GetAllLatestPublishedConsents`, `GetLatestPublishedConsentsForProject`,
      `GetLatestConsentByKey`, `GetMissingConsentsForUserWithRejections`),
      to `CreateConsent`'s insert, and to `UpdateConsent`.
      `UpdateConsent` uses the "empty string keeps the old value" trick for
      strings, which has no bool equivalent — use
      `required = COALESCE(sqlc.narg('required'), required)` so omitting it is
      a no-op.
- [ ] New query `GetBlockingConsentsForUser` — required, published, scoped to
      the project, where the latest `user_consent_history` row for the key is
      missing or `REJECTED`. (Or compute it in `buildConsentStatus` from data
      already loaded — preferred, since the loader has both halves in memory.)
- [ ] `make generate`.
- [ ] `gql/consents.graphqls`: - `Consent.required: Boolean!` - `ConsentStatus.blockingConsents: [Consent!]!` - `createConsent(..., required: Boolean)`,
      `updateConsent(..., required: Boolean)`
- [ ] `gql/users.graphqls` — no change (`consentStatus` already returns the type).
- [ ] `backend/internal/loaders/user.go`: - carry `Required` into the `model.Consent` built in
      `fetchConsentDataForUsers` - `buildConsentStatus`: build `blockingConsents` using decision #2, with
      the same under-13 skip as the other buckets.
- [ ] `consents.resolvers.go`: - `CreateConsent` / `UpdateConsent`: pass `required`, validate #4 - `RejectConsent`: error when `consent.Required` - `AdminSetUserConsent`: error when setting `REJECTED` on a required
      consent (same rule, admin path) - carry `Required` into every `model.Consent` built in this file
      (`Consents`, `PendingConsents`, create, update)
- [ ] `backend/internal/loaders/consent.go` — carry `Required` through
      `LoadConsentWithTranslation` / `ApplyTranslationToConsent`.
- [ ] Caches: `PrefixLatestConsents` is already invalidated on create/update,
      and `cache.UserKey` on accept/reject. Nothing new needed — but check that
      `AdminSetUserConsent` and the consent webhook invalidate the _target_
      user's key (it does).

### Phase 2 — server-side enforcement

In scope as of 2026-10-15. **"Server-side enforcement" is two separate
mechanisms**, and the notification case needs the second one, not the first:

#### 2a — Inbound request guard

A GraphQL middleware that, when the caller has blocking consents, rejects every
operation except an allowlist (`me`, `consents`, `consent`, `pendingConsents`,
`acceptConsent`, settings/language, plus anything admin). Costs one cached
lookup per request — the user cache already holds `ConsentStatus`.

This closes the "crafted GraphQL call bypasses the UI" hole. It does **nothing**
for notifications: a push is not a request the blocked user makes, so this
middleware never runs on that path.

- [ ] Middleware in `backend/internal/middleware/`, wired in `cmd/server/main.go`
- [ ] Allowlist by operation name; deny with a typed error the frontend can
      recognise and turn into a redirect to the consent screen
- [ ] M2M and admin tokens exempt

#### 2b — The participation filter

Per decision #7 a blocked user is not a participant, so they must disappear
from every read that answers "who is in this project / unit". Today **nothing
is consent-aware**; all of these start from `user_projects` or `team_members`:

| Surface             | Where                                                                                         |
| ------------------- | --------------------------------------------------------------------------------------------- |
| Push audiences      | `services/push/service.go` `resolveAudience` (~350) — teams, projects, events, all-subscribed |
| Bulk job targeting  | `services/bulk/service.go` (~290) — church+project, teams, super-teams, all project members   |
| Person leaderboards | `GetProjectPersonLeaderboard` + `GetFull…` + `Count…` — the `project_users` CTE               |
| Unit leaderboard    | `Team.MemberLeaderboard` (`teams.resolvers.go:1539`)                                          |
| Team rosters        | `team_members` reads behind the unit view                                                     |
| Aggregate boards    | Team / super-team / church boards — their score must drop out (decision #9)                   |
| Rivals              | `resolveNearestChurchRivals` (`helpers.go:30`), `Leaderboard.me` neighbours                   |
| Unit roster         | `Team.members` (`gql/teams.graphqls:12`) and anything counting it                             |

That is roughly a dozen queries once the paged / full / count variants are
counted, so **do not hand-filter each one**:

- [ ] Define the predicate **once**, as a `(user_id, project_id)` set — e.g. a
      view `blocked_project_users` — that each query anti-joins with
      `NOT EXISTS`. **Not** a view wrapping `user_projects`: the aggregate
      boards never read that table. `GetProjectTeamLeaderboard` goes
      `teams → team_members → users → score_journal` (`leaderboards.sql:187`),
      so the predicate must compose with a `team_members` join just as well.
      One definition, no drift, and sqlc reads views fine.
- [ ] Only then wire it into the surfaces above. The person leaderboard needs
      it in exactly one place — the `project_users` CTE — which covers the
      whole query.
- [ ] Counts must use the same predicate as the rows, or pagination breaks.
- [ ] Direct `UserIDs` push targets: filter those too. **Recommend yes** — an
      admin picking users by hand should not reach someone the system treats as
      a non-participant.

**Admin reads are exempt** (decision #10): the filter belongs on user-facing
resolvers only. Pushing it down into shared SQL indiscriminately would blank
the admin UI as well, which is where a blocked user has to stay visible.

**The under-13 exemption (decision #3) must hold here too.** A looser "has no
ACCEPTED row" predicate would silently make every under-13 user invisible and
unreachable. Same rule as `buildConsentStatus`, defined once.

**Performance is a real risk.** The leaderboard queries are already tuned
(`MATERIALIZED` CTEs, the "huge performance win" comment at
`leaderboards.sql:4`) and this repo has a loadtest history to protect — see
`notes/loadtest-*.md` and `notes/optimization-history.md`. The predicate needs
a per-user "latest action per key" lookup, which is not free. Plan: define it
as a view, **measure against the existing loadtests**, and only if it regresses
fall back to a denormalised `user_projects.consent_blocked` flag maintained on
accept / reject / consent publish. Do not start with the denormalised column —
it must be swept on every publish of a required consent, and it will drift.

### Phase 3 — user frontend

- [ ] `pnpm codegen` after the schema lands.
- [ ] New `frontend/app/middleware/03.consent.global.ts` (root, so it runs
      _after_ auth): - skip `/login`, `/auth0-callback`, `/logout-callback`, `/settings/consent`
      and all `/admin/**` routes - read blocking consents from a shared composable; if non-empty,
      `navigateTo({ name: 'settings-consent' })`
- [ ] New composable `useBlockingConsents()` — a single `BlockingConsents`
      query (`me { consentStatus { blockingConsents { ... } } }`) held in
      `useState`, awaited by the middleware, refetched after an accept. Paused
      until `useAuthReady()`, like `useCurrentProjectQuery`.
- [ ] `layers/user/app/pages/settings/consent.vue`: - hide the close button and the "Gå videre" button while blocking
      consents remain (today both key off the `hasCompletedOnboarding`
      localStorage flag — blocking must win over it) - sort blocking consents to the top
- [ ] `ConsentCard.vue` — hide the reject button when `consent.required`;
      make a required-and-unaccepted card visually distinct.
- [ ] `layers/user/app/pages/index.vue` — drop the ad-hoc `pendingInternal`
      redirect once the middleware covers it; keep the remote-consent banner.
- [ ] i18n: new keys under `consent.*` (e.g. `requiredNotice`,
      `requiredBlockTitle`) in `frontend/i18n/locales/nb.json` + `en_us.json`;
      the rest come from Phrase.

### Phase 4 — admin frontend

- [ ] `new.vue` — `required` toggle in the zod schema + form, passed to the
      mutation. Disable it when `isRemote` is on (decision #4).
- [ ] `[consentId].vue` — same toggle, wired to `updateConsent`.
- [ ] `index.vue` — "Påkrevd" badge in the consent list.
- [ ] `AdminUserConsents.vue` — reject action unavailable for required consents.

### Phase 5 — tests

- [ ] `backend/internal/loaders/user_test.go` — `buildConsentStatus`:
      required+no-history → blocking; required+REJECTED → blocking;
      required+ACCEPTED → not blocking; optional+REJECTED → not blocking;
      under-13 + required local → not blocking.
- [ ] `backend/internal/graph/api/consents_test.go` — `rejectConsent` on a
      required consent errors; `createConsent` with `required && isRemote`
      errors.
- [ ] `backend/e2e/consents_required_test.go` — accept clears the block;
      project scoping still applies (a required consent on project B does not
      block a user on project A).
- [ ] `frontend/test/unit/` — middleware redirects when blocking, lets admin
      and the consent route through.
- [ ] `services/push/service_test.go` — a user with a blocking consent is
      dropped from the audience; an under-13 user with a required local consent
      is **not** dropped; an accepted user still receives.
- [ ] `services/bulk/service_test.go` — same filter on bulk targeting.
- [ ] Leaderboard tests — a blocked user is absent from the person board, from
      `Team.MemberLeaderboard`, and from the row counts; accepting puts them
      back without a re-join.
- [ ] Loadtest the person leaderboard before/after the participation filter
      (`notes/loadtest-*.md` has the baselines).
- [ ] Middleware test — allowlisted operations pass, others rejected, m2m exempt.
- [ ] `make fmt` + `make test` in `backend/`.

## Risks / gotchas

- **`DISTINCT ON (key)` + `ORDER BY key, version DESC`** picks the latest
  _published_ version. Marking an old version required has no effect; the admin
  UI should make that obvious.
- **Rejected-then-required**: a user who rejected the consent when it was
  optional is blocked the moment it is published as required. That is intended,
  but it is a support surface — the history row stays, so the audit trail shows
  why.
- **The under-13 skip is silent** (`buildConsentStatus`): an admin marking a
  consent required will not see that a slice of users is exempt. Worth a note in
  the admin UI.
- **`hasCompletedOnboarding` is localStorage**, so it is per-device and
  clearable. It must never be able to override a blocking consent.
- **The audience filter is easy to forget.** Every new query that builds a
  recipient list from membership tables needs it. Extracting the predicate into
  one helper (used by `buildConsentStatus`, push and bulk) is the only thing
  that keeps this from rotting.
- Updating this note's decisions table is part of the work — if #4 or #6 flips
  during implementation, record it here.

## Files to touch

```
backend/internal/database/migrations/00109_add_consent_is_required.sql (new)
backend/internal/database/queries/consents.sql
backend/internal/loaders/user.go
backend/internal/loaders/consent.go
backend/internal/graph/api/consents.resolvers.go
backend/internal/middleware/<consent guard>.go                        (new, 2a)
backend/internal/services/push/service.go                             (2b)
backend/internal/services/bulk/service.go                             (2b)
backend/cmd/server/main.go                                            (wiring)
gql/consents.graphqls
frontend/app/middleware/03.consent.global.ts                          (new)
frontend/layers/user/app/composables/useBlockingConsents.ts           (new)
frontend/layers/user/app/pages/settings/consent.vue
frontend/layers/user/app/pages/index.vue
frontend/layers/user/app/components/consent/ConsentCard.vue
frontend/layers/admin/app/pages/admin/consents/{index,new,[consentId]}.vue
frontend/layers/admin/app/components/admin/user/AdminUserConsents.vue
frontend/i18n/locales/{nb,en_us}.json
```
