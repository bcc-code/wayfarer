# Required (blocking) consents

**Status:** Planned — not started
**Started:** 2026-10-14
**Branch:** feature/required-consents

## Goal

A consent can be marked **required**. A user who has not accepted a required
consent that applies to the current project cannot use the app: every route
sends them to the consent screen, and that screen offers no way out except
accepting. Rejecting a required consent is not possible.

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

## Design decisions

| #   | Question                              | Decision                                                                                                                                                                                                                                                                                                                    |
| --- | ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Is `required` per key or per version? | **Per version**, like `title`/`body`: the latest published version wins, so a consent can be made required by publishing a new version. Unlike `project_id` (which is inherited and locked by `consentProjectForNewVersion`), there is no reason to freeze it — widening it is the whole point.                             |
| 2   | What counts as blocking?              | Latest published consent for the current project, `required = true`, and the user's **latest action for that key is not `ACCEPTED`** — i.e. no history row _or_ a `REJECTED` row. This is a new rule; "pending" elsewhere still means "no history".                                                                         |
| 3   | Users under 13                        | Keep the existing skip: LOCAL consents are not shown to under-13s, so a required local consent does **not** block them. Changing that would lock out every under-13 user with no path forward.                                                                                                                              |
| 4   | Required + remote                     | **Reject at the API**: `createConsent`/`updateConsent` error if `required && isRemote`. A remote consent can only be accepted in the external system, so requiring one would strand the user behind a screen with no accept button. Revisit if SSF consent ever needs to be mandatory — that needs a "re-check" flow first. |
| 5   | Rejecting a required consent          | `rejectConsent` errors. The reject button is hidden client-side too, but the API is the authority.                                                                                                                                                                                                                          |
| 6   | Server-side enforcement               | Phase 2, see below. Phase 1 is a client gate; that is honest UX but not a security boundary.                                                                                                                                                                                                                                |

## Plan

### Phase 1 — backend data + API

- [ ] Migration `00108_add_consent_required.sql`:
      `ALTER TABLE consents ADD COLUMN required BOOLEAN NOT NULL DEFAULT false;`
      plus a `COMMENT ON COLUMN`. Down drops it. **Needs approval before running.**
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

### Phase 2 — server-side enforcement (decide before shipping)

Phase 1 blocks the UI only; a crafted GraphQL call still works. To make the
block real, add a GraphQL middleware that, when the caller has blocking
consents, rejects every operation except an allowlist (`me`, `consents`,
`consent`, `pendingConsents`, `acceptConsent`, settings/language, plus anything
admin). Costs one cached lookup per request (the user cache already holds
`ConsentStatus`, so it is free in practice).

**Open:** is this wanted now, or is the UI gate enough for the first release?

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
- Updating this note's decisions table is part of the work — if #4 or #6 flips
  during implementation, record it here.

## Files to touch

```
backend/internal/database/migrations/00108_add_consent_required.sql   (new)
backend/internal/database/queries/consents.sql
backend/internal/loaders/user.go
backend/internal/loaders/consent.go
backend/internal/graph/api/consents.resolvers.go
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
