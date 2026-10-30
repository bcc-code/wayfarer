# Admin consent management

**Status:** done (migration 00118 not yet run on dev)
**Changed:** 2026-10-28

## What the admin pages list

`Query.consents` returns the latest _published_ version of each key. That is
right for a client asking what a user must agree to, and wrong for an admin
page: a draft was invisible on the only page that can publish it, and the
`Utkast` badge in the listing template could never render.

`Query.adminConsents(projectId: ID, includeArchived: Boolean)` is the admin
surface instead:

- the latest **live** version of each key (`published_at <= now()`), plus
- every version that is **not live yet** — no `published_at` (draft) or one
  still in the future (scheduled) — as its own row.

So a key with a live v1 and a draft v2 shows both: what users see now, and what
is waiting. Superseded published versions stay hidden.

`projectId` is optional. Omitted, nothing is filtered (the global
`/admin/consents` list). Set, the result is that project's own consents plus
the global ones (`project_id IS NULL`) — the same set
`GetLatestPublishedConsentsForProject` gates users on, which is what
`/admin/projects/:projectId/consents` shows.

`includeArchived` defaults to false; the "Vis arkiverte" switch on both pages
turns it on.

Authorisation is **in the resolver** (`RoleService.IsAdmin`), not
`@requireRole` — that directive is for mutations here. Drafts are unpublished
text, so `adminConsents` is admin/superadmin only. Plain `consents` is
unchanged apart from now excluding archived consents.

The pages render three tables, split client-side by `groupConsentsForAdmin`:
**Utkast** (drafts, shown only when there are any), **Publiserte** (live and
scheduled, badged apart) and **Arkiverte** (behind the switch).

## Archiving, never deleting

`consents.archived_at` (migration 00118) retires a consent.
`Mutation.archiveConsent(id)` / `unarchiveConsent(id)`, admin/superadmin.

There is deliberately **no delete**. `user_consent_history.consent_id` is
`ON DELETE RESTRICT` (migration 00027) because that history is the record of
what each user agreed to; it has to outlive the consent. Archiving takes a
consent out of circulation and leaves every row of that record in place.

What archiving changes:

- it drops out of `pendingConsents` and `me { consentStatus }`, so nobody is
  asked for it again, and a **required** consent stops blocking participation.
  00118 also redefines `consent_blocked_user_ids` to skip archived keys, and
  both mutations run `applyConsentBlockChanges`, so the boards and unit views
  show the unblocked users straight away (and hide them again on unarchive;
  `e2e/consents_blocked_visibility_test.go`);
- `acceptConsent` / `rejectConsent` refuse it (a late answer is a stale client);
- it leaves the default admin listing, and appears under "Arkiverte".

### The rule that matters: a key is live iff its latest published version is unarchived

Every live-consent query filters `archived_at IS NULL` **outside** the
`DISTINCT ON (key)`, never inside it. Filtering inside would make archiving v2
promote v1 back into circulation — users would be asked to agree to the
superseded text. `GetAllLatestPublishedConsents`,
`GetLatestPublishedConsentsForProject` and
`GetMissingConsentsForUserWithRejections` all follow this, and
`e2e/consents_admin_test.go` pins it ("does not resurrect the version
underneath").

`GetLatestPublishedConsentByKey` deliberately does **not** filter: it is the
webhook/registration path, and an external system's report of what a user
answered is worth recording even for a retired consent.

## Cache invalidation (a bug this work uncovered)

`cache.InvalidateConsents()` is called by create, update, archive and
unarchive. It clears three layers, and all three are load-bearing:

| Layer                   | Key                          | Why it goes                                                                                                  |
| ----------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------ |
| Live set per project    | `latestconsents:<projectID>` | What the gate reads                                                                                          |
| Every cached user       | `user:<userID>`              | `model.User` embeds a `ConsentStatus` built from that set (`loaders/user.go`)                                |
| Whole GraphQL responses | `gqlresponse:`               | A mutation only drops the _caller's_ entries; the caller is an admin, the stale entries belong to every user |

Two pre-existing bugs fell out of this: `PrefixLatestConsents` was missing from
`extractPrefixes`, so **every** `DeletePrefix(PrefixLatestConsents)` since the
cache registry was introduced — including the ones in `createConsent` and
`updateConsent` — silently matched nothing; and nothing cleared the embedded
`ConsentStatus` or the response cache, so publishing a consent could take up to
the cache TTL to reach users. Covered by
`TestKeyRegistry_LatestConsentsKeyIsFoundByPrefix` and
`TestCacheWithRegistry_InvalidateConsents`.

## Where the code lives

| Piece                 | File                                                                                                        |
| --------------------- | ----------------------------------------------------------------------------------------------------------- |
| Migration             | `backend/internal/database/migrations/00118_add_consent_archived_at.sql`                                    |
| SQL                   | `backend/internal/database/queries/consents.sql` (`GetAdminConsents`, `ArchiveConsent`, `UnarchiveConsent`) |
| Schema                | `gql/consents.graphqls`                                                                                     |
| Resolvers             | `backend/internal/graph/api/consents.resolvers.go`, helpers in `consents.go`                                |
| Cache                 | `CacheWithRegistry.InvalidateConsents` in `backend/internal/cache/invalidation.go`                          |
| E2E                   | `backend/e2e/consents_admin_test.go`                                                                        |
| Tables UI             | `frontend/layers/admin/app/components/admin/consent/`                                                       |
| Status/grouping rules | `frontend/layers/admin/app/utils/consents.ts`                                                               |
| Pages                 | `frontend/layers/admin/app/pages/admin/consents/`, `.../admin/projects/[projectId]/consents/index.vue`      |

Both list pages are gated by `consents:view`, which is **superadmin** today
(`canAccessConsents`); archive and restore additionally need
`canManageConsents` (admin or superadmin). The project-scoped page follows the
same rule, so a project admin does not see it.

Not covered: `pending_consent_events` rows reference a consent by **key**, not
by id, so archiving does not clear events queued for users who have not
registered yet. They would apply to a later version of the same key.
