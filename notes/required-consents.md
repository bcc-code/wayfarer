# Required (blocking) consents

**Status:** Phase 1 (column) and Phase 3 (client-side gate) done. Phase 2b
step 1 (hide from other users) implemented 2026-10-23 on
`feat/consents-required-backend` — migration `00116` **written, not yet run**.
Averages, push/bulk and the reject guard are still open; see "Phase 2b
re-survey" and "Step 1".
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

`user_consent_history` is append-only and stores both `consent_key` and
`consent_id`. `action` is `ACCEPTED` | `REJECTED`, constrained by
`check_action`. Answers count **per version** (`consent_id`), since 00117: see
design decision 11.

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
| 11  | Does an answer carry over to a new version?                       | **No — decided 2026-10-29.** A user has not consented to a version they have not seen. Only an answer to the latest published version counts, for every consent (required or not, local or remote): a new version is pending for everyone, also for users who rejected the previous one. `acceptConsent`, `rejectConsent` and `adminSetUserConsent` refuse any but the latest version. Editing a version in place (`updateConsent`, no bump) does not ask again; that is the way to fix typos. Implemented in 00117 (`consent_blocked_user_ids`), `GetCurrentUserConsentStatusesByUsers` / `buildConsentStatus`, `GetMissingConsentsForUserWithRejections`, and the `leaderboard_consent` check of the person boards. Until then answers were matched per key, so a version bump kept them. |

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

> **Superseded 2026-10-22** — the mechanics below assume leaderboards are
> computed at read time from a `project_users` CTE. They are not: see
> [Phase 2b re-survey](#phase-2b-re-survey-2026-10-22) further down. The surface
> list and the under-13 / admin-exempt constraints still hold; the "one view,
> anti-join everywhere" plan does not.

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

#### Phase 2b re-survey (2026-10-22)

Information gathering only — no implementation decided. Requirement restated
by Jan on 2026-10-22, matching decisions #7, #9, #10:

1. A user who has not accepted the required consents must not "exist" for other
   users.
2. Not visible in leaderboards or in their unit; their points must not count
   toward averages.
3. Admins must still see them, so they can place them into units.

##### Finding A — leaderboards are pre-aggregated; the old 2b targets dead code

- Every board is served from 8 stored tables,
  `leaderboard_{project,event}_{persons,teams,superteams,churches}`, through
  the `GetFull*` queries only (`services/leaderboard.go:347`, `:386`, …),
  paginated in Go, `totalCount = len(entries)`.
- `Get{Project,Event}*Leaderboard`, `Count*Leaderboard`, `FindMy*Position` —
  including the `project_users` CTE the old plan targets
  (`leaderboards.sql:4`, `:154`) — have **no callers**. They are still listed in
  `LeaderboardQuerier` (`services/leaderboard.go:27-69`).
- Team / superteam / church `score` is a **stored average**:
  `calculate_average_score(total_points, member_count)` = `ROUND(total/count)`
  (`00080:37`). Both `total_points` and `member_count` are maintained on write:
  - `score_journal` INSERT → `leaderboard_apply_queue` → worker
    (`services/leaderboard_apply.go`, 200 ms tick) → `apply_score_journal_delta`
    (`00101:24`). It fans out to person, current church, each team in the
    project, and each superteam once. Each step recounts `member_count` live.
  - Membership triggers (`00081`: `team_members` :16, `user_projects` :222,
    `user_events` :307) move a user's whole `SUM(score_journal)` in or out.
  - Church change, superteam reassignment and the `leaderboard_excluded`
    toggle trigger nothing.
- `member_count` is defined differently per board:
  - project team = `COUNT(*) team_members` (no `user_projects` check);
  - superteam = distinct members across its teams;
  - project church = users of that church in `user_projects`;
  - event boards additionally require `user_events`.
- `Team.memberLeaderboard` (the unit view) is separate: a live
  `SUM(score_journal)` over `team_members` (`teams.sql:201`), cached per team
  for 15 min.

**Consequence — decisions #8 and #9 now conflict.** #9 (points do not count in
unit / superteam / church averages) cannot be done with a read-time filter. Both
the numerator *and* the `member_count` denominator must drop the user, otherwise
a 0-point "member" still drags the average down. #8 ("filter at read time,
accepting restores instantly") assumed aggregates are computed on read. Options,
all of which touch the write side (**not chosen**):

- **(a) Blocked = non-member on write.** The triggers, `apply_score_journal_delta`
  and the `member_count` recounts treat a blocked user as absent. Accept, reject,
  and publishing a new required consent then have to move that user's points in
  or out (like a team join or leave). Publishing a required consent sweeps
  every affected user. `regenerate_leaderboards()` (`00113:12`) must apply the
  same rule, or rebuild ≠ incremental.
- **(b) Recompute aggregates on read** for team / superteam / church. This
  gives up the stored-table design and its loadtest gains.
- **(c) Denormalised blocked flag** (e.g. on `user_projects`) that the triggers
  read. Same sweep problem as (a), plus drift risk.

The **persons board** can still be filtered on read. Its
`leaderboard_consent` check (`leaderboards.sql:132-147`, `:594-609`) is the
precedent and the natural place, since it already sits in the only live query.

##### Finding B — admin and user read the same fields and the same caches

- The admin UI reuses the user-facing fields:
  - `AdminTeamPage` (`admin/projects/[projectId]/teams/[teamId].vue`) reads
    `team(id){ members, memberLeaderboard }`;
  - `MyChurchUnitsPage` (`admin/my-church/units.vue`) reads
    `myCurrentProject.myChurchTeams{ members }` plus `users(filter)`.
- The full-board cache key `leaderboard:full:{ctx}:{id}:{entity}:{filtersHash}`
  (`cache/keys.go:1136`), the per-team member-leaderboard cache, and the 30 s
  `ResponseCache` (`graph/api/response_cache.go:76-80`) are not keyed by who is
  asking.

So a filter in a shared resolver either hides blocked users from admins too,
or leaks them to users through the cache. It needs separate admin fields, or
cache keys that include the role.

- Accept and reject only clear `cache.UserKey` (`consents.resolvers.go:151`,
  `:217`, `:428`). They do not clear board or member-leaderboard caches.

##### Finding C — surfaces where one user sees another (non-admin)

| Surface                                                          | Where                                                                                | Gate today                                                         |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| Persons boards (+ `me`, `nearestChurchRivals`, configured boards) | `GetFull{Project,Event}PersonLeaderboard`                                            | any user; `leaderboard_consent` unless team/superteam-filtered      |
| Team / superteam / church boards                                 | stored aggregates (Finding A)                                                        | any user                                                           |
| Unit view                                                        | `Team.memberLeaderboard` `teams.resolvers.go:1539` → `teams.sql:201`                 | **none** — any user, any team; no consent filter                    |
| Unit roster                                                      | `Team.members` `teams.resolvers.go:1496`                                             | `CanManageTeam` (admin, project_admin, church_admin, **team_lead**) |
| Team stats                                                       | `Team.averageAge` (`teams.sql:276`); `teams(filter)` `minMembers`/`maxMembers`        | none — reveals member counts and ages                              |
| Superteam roster                                                 | `SuperTeam.members` `teams.resolvers.go:1328`                                        | m2m, admin, project_admin                                          |
| User lookup                                                      | `user(id)`, `users(filter)`                                                          | `CanAccessUser` / role-gated                                       |
| Quiz session creator                                             | `QuizSession.createdBy` `quiz_sessions.resolvers.go:879`                             | none                                                               |
| Counts                                                           | `completionCount`, `awardedUserCount`, `activityTrend.activeUsers`, `QuizSession.submissionCount` | none, but only used by admin pages                     |

Audiences (another user is not the viewer here, but per decision #6 a blocked
user must not be reached):

- **Push** — `resolveTargetUserIDs` (`services/push/service.go:354`): `userIds`,
  `teamIds`, `projectIds`, `eventIds`, `allUsers`. Plus automatic per-user pushes
  for achievement, challenge enrolment and bet result
  (`services/push/translated_notifications.go` callers).
- **Bulk** — `ResolveEnrollmentTarget` (`services/bulk/service.go:265`):
  `userIds`, `churchInProject`, `teamIds`, `superTeamIds`, `allProjectMembers`.
  Also `GrantQuizSessionAccess`, `ProcessBulkAwardAchievement` and team score
  adjustments.
- **Webhooks / plugin** — `quiz_session_finished` sends every participant's
  user id, members id and score. The Ladder to Heaven `team-name-changed`
  handler awards points to every team member (`GetUserIDsInTeams`).

Not checked: `JoinTeam` response (`teams.resolvers.go:26`) and
`resolveBulkAwardTarget`.

##### Open questions for Jan (decide before implementing)

1. **Aggregates**: option (a), (b) or (c) above? Or relax #9 so blocked users'
   points still count, which would allow a pure read-time filter.
2. **Who counts as "admin"** for decision #10? Placing users into units happens
   in `admin/my-church/units.vue`, which church admins use. `Team.members` is
   also visible to `team_lead`, an ordinary participant. Do church_admin,
   project_admin and team_lead see blocked users?
3. **What exactly hides a user**: #2 (latest action on the required key is not
   `ACCEPTED`), keeping the under-13 LOCAL exemption (#3)? Any predicate added
   to SQL must reproduce `buildConsentStatus` exactly, or minors vanish.
4. **Other exceptions** to "mandatory"? For example, imported members who have
   never logged in — they have never answered anything, so at camp start whole
   units would look empty to their teammates. Remote consents are already
   excluded (#4).
5. **Relation to `leaderboard_consent`**: keep both? It already hides
   non-consenting users from persons boards, but not when the board is filtered
   by team or superteam (including `myTeam`), and not from aggregates.
6. **Direct `userIds` push / bulk targets** chosen by an admin: filter or not?
   The old note recommended yes.

##### Side findings — not in scope, for separate issues

- `Query.scoreJournal(projectId, userId)` has no authorization at all. There is
  no `@requireRole` and no check in `graph/api/score_journal.go`, so any
  logged-in user can read any user's point history. The frontend does not
  call it.
- `Team.memberLeaderboard` is ungated (any user, any team), and
  `GetTeamMemberLeaderboard` selects `email` and `birthdate`. They are not
  exposed through `LeaderboardEntry`, but they are loaded and cached.
- The leaderboard filters `gender`, `country` and `churchCategory` go into the
  cache key but are never passed to SQL, so they are silently ignored.
- Possible double count. Membership triggers move the full
  `SUM(score_journal)`, including deltas still in `leaderboard_apply_queue`,
  which the worker then applies again. Found by reading the code only; no test
  covers it.
- The Ladder to Heaven distribution sum (`super_teams.sql:99`) joins
  `user_roles`, so a team with several leads would have its sum multiplied.
  Not verified.

#### Requirement clarified 2026-10-23

Jan's use case:

1. A local church admin creates units.
2. The admin adds person X to unit Y.
3. X has not accepted the consent to join the project.
4. X is still visible to admins: as a member of unit Y, of the church, etc.
5. X is not visible to any other user: not in a leaderboard, not in the unit,
   **not even to the unit leader**.
6. X's points and data are not calculated at all — not in the unit's average
   score, average age, and so on.

This settles three open questions:

- **#1 (aggregates):** do not relax #9.
- **#2 (who is "admin"):** admin, superadmin, project_admin and church_admin
  see X. A team_lead does **not**. Today `Team.members` is gated by
  `CanManageTeam`, which admits team leads, so it has to filter blocked users
  for a viewer who is only a team lead.
- **Calculated numbers:** X counts in nothing, for admins too. Admins see X
  *listed*, but averages, ages and counts exclude X everywhere, so there is no
  need for admin-only numbers or role-aware caches.

What it covers beyond step 1:

- Stored team / superteam / church scores (project and event): both
  `total_points` and `member_count`.
- `GetTeamAverageAge` (`teams.sql:285`): the `Team.averageAge` field, and the
  automatic `leaderboard_excluded` rule (average age ≥ 36), which is only
  recomputed on member add/remove (`teams.resolvers.go:487`, `:581`). If
  blocked users are excluded from the age, accept/reject must recompute it.
- `teams(filter)` `minMembers`/`maxMembers` member counts (`teams.sql:10`,
  `:35`).
- External: Ladder to Heaven superteam distribution (team totals and member
  counts) and the `quiz_session_finished` webhook (every participant's score).

#### Step 1 — hide from other users ✅ implemented 2026-10-23

Scope agreed with Jan on 2026-10-23: only hide. A user who has not accepted
the project's required consents is missing from the person leaderboards and
the unit view.

Deliberately **not** in step 1:

- Their points still count in team, superteam and church averages
  (Finding A).
- Push and bulk audiences are not filtered.
- There is no reject guard.

On the admin team page a blocked member is still listed via `Team.members`,
but shows **0 points and no rank**. Jan accepted this for step 1, to be
improved later.

**The rule lives in one SQL function:** `consent_blocked_user_ids(project_id)`
in migration `00116_add_consent_blocked_user_ids.sql`. Jan chose this on
2026-10-23 over inline copies, because later steps (averages in the triggers,
push/bulk audiences, real points for admins) need the same rule. It mirrors
`buildConsentStatus` and the frontend's `blockingConsentIds`:

- latest published version per key, global or the project's own; `is_required`
  is tested *after* picking the latest version;
- local only;
- the user's latest action for the key is missing or `REJECTED`;
- under-13 users (by year difference) are exempt.

`users.birthdate` is `NOT NULL` (`00006`), so there is no missing-birthdate
case. Candidates come from `user_projects`: both team-join paths
(`JoinTeam`, `AddTeamMembers`) enroll the user in the project.

**Where the function is used**, each time as a `MATERIALIZED` CTE `blocked`
that the query anti-joins:

- `GetFullProjectPersonLeaderboard` and `GetFullEventPersonLeaderboard`
  (`leaderboards.sql`). These are the only live person-board queries, so
  `me`, `nearestChurchRivals`, configured boards and `totalCount` follow
  automatically. The unused `FindMy*` / `Get*` / `Count*` variants are
  untouched.
- `GetTeamMemberLeaderboard` (`teams.sql`), which backs `Team.memberLeaderboard`.

The filter sits in `WHERE`, so `DENSE_RANK` / `RANK` stay contiguous. The CTE
is materialized because, on test data, the inlined function produced a
nested-loop anti-join that would re-run the set per row.

**sqlc gotchas:**

- sqlc does not know a function's `RETURNS TABLE` columns. Alias them in the
  query: `consent_blocked_user_ids(...) AS b(user_id)`.
- `make generate | tail` hides a failure; use `set -o pipefail`.

**Cache invalidation** (`consents.resolvers.go`) uses
`InvalidateLeaderboardScores()`:

- on `acceptConsent`, `rejectConsent` and `adminSetUserConsent` when the
  consent is required;
- always on `createConsent` and `updateConsent`.

Known gaps:

- Accepting an *older, non-required* version of a key whose latest version is
  required unblocks the user without invalidating, so it is stale until the
  TTL expires (5 min boards, 15 min unit).
- A future `published_at` takes effect with no write. The TTL covers it.
- **Invalidation burst.** `InvalidateLeaderboardScores()` also clears the
  whole 30 s GraphQL response cache, and it now runs on every accept of a
  required consent, with no throttle. The worst case is when a required
  consent is first published and everyone accepts at once, before camp, when
  no scores are flowing. A naive "once per second" guard would drop
  invalidations and leave a user who just accepted hidden. A proper fix needs
  a trailing flush. Not done.
- **Per-instance cache.** The cache is in-memory, with no cross-instance
  broadcast. "Visible straight away" holds on the instance that handled the
  accept. Other instances catch up on TTL or on their own next invalidation.

**Deploy order:** migrations are run by `cmd/migrate`, not at server startup.
Apply `00116` **before** deploying this code: every person leaderboard and the
unit view call `consent_blocked_user_ids` and error without it. The same
applies to the shared dev database. Re-check that `00116` is still free on
every branch right before merging.

**Tests:** `backend/e2e/consents_blocked_visibility_test.go` checks:

- accepted, never-answered, rejected-after-accepting, under-13 and turns-13
  users;
- project and event boards and the unit view, with contiguous ranks and
  `totalCount`;
- admins still see everyone in `Team.members`;
- another project's required consent, a required remote consent, and a
  required v1 relaxed by v2 do not hide anyone;
- accepting shows the user straight away with warm caches. This fails if the
  accept invalidation is removed (checked).

**Not done:**

- Loadtest against the `notes/loadtest-*.md` baselines.
- Running `00116` on a real database (needs approval).

#### Step 1b — hidden from unit leaders ✅ implemented 2026-10-26

Use-case step 5: a unit leader must not see X either. The app already hid X
from leaders, because `StandingsUnit.vue` and the lead picker read
`memberLeaderboard`, which step 1 filters. Two API paths still exposed X to a
lead:

- `Team.members`, gated by `CanManageTeam`, which admits team leads.
- `users(filter: { teamId })`, where `checkUserPermissions` lets a lead list
  their own team.

Jan chose approach A (2026-10-26): filter those two for a viewer whose access
comes only from being the team lead, instead of removing team leads from the
gates (B).

- `RoleService.CanManageTeamAsAdmin` is `CanManageTeam` without the team-lead
  clause: admin, superadmin, the team's project admin, or the church admin of
  a member's (or the creator's) church. `CanManageTeam`'s call order is
  unchanged, so its mock-based tests still hold.
- `Team.members`: `seesHidden` = m2m or `CanManageTeamAsAdmin`. Otherwise
  access requires `HasRoleInTeam(team_lead)`, and
  `GetConsentBlockedUserIDsInTeam` (`teams.sql`) is filtered out in Go
  (`teamMembersFor`, `teams.go`; was `teamMembersWithout` before step 2). The
  members loader cache stays shared.
- `users()`: `permissionContext.HideConsentBlocked` is set when a lead is not
  `CanManageTeamAsAdmin`. The hidden IDs go into
  `GetUsersFilteredCursor`/`CountUsersFiltered` as `@excludeids`, and into the
  cache key (`addExcludeIDsToCacheKey`), so an accept changes the key with no
  invalidation and a lead's result is never served to an admin.
- Tests:
  - `users_test.go`: `HideConsentBlocked` for a plain lead versus an admin lead.
  - `roles_test.go`: `CanManageTeamAsAdmin` is false for a team lead only, and
    must not even check the lead role; it is true for the church admin of a
    member's church.
  - e2e: church admin sees all members in both places; the lead sees only the
    visible ones, with a matching `totalCount`. Verified to fail with the old
    checks.

#### Step 2 — group boards computed from person totals ✅ implemented 2026-10-26

Approach (a) (agreed 2026-10-23: blocked users as non-members in every
writer of the stored group totals, plus full rebuilds) was dropped on
2026-10-26 for computing the group boards on read. The stored totals have five
writers, each with its own copy of the member rules, and drift already
(`notes/leaderboard-regenerate.md`). Separate refactor PR
(`refactor/group-leaderboards-from-persons`), no consent logic: the six group
`GetFull*` queries average the stored person totals on read, with exactly the
rebuild's member rules. Performance (EXPLAIN on `gendata.sql`, k6 ramp) still
to be measured before merging.

#### Step 3 — counts nowhere (use-case step 6) ✅ implemented 2026-10-26

Branch `feat/consents-exclude-from-averages`, stacked on step 2. A blocked user
counts nowhere, for admins too:

- Group boards (`leaderboards.sql`): one `MATERIALIZED` blocked CTE per board
  (event boards: the event's project). Blocked users count neither in the
  points nor in the member count, nor for a team's `churchId` filter.
- `GetTeamAverageAge` (`teams.sql`): `Team.averageAge` and the automatic
  `leaderboard_excluded` rule (≥ 36, `leaderboardExcludedMinAverageAge`,
  `teams.go`) leave blocked members out.
- `teams`/`superteams` `minMembers`/`maxMembers` (`teams.sql`,
  `super_teams.sql`): the member count leaves blocked users out. The blocked
  set is a `MATERIALIZED` CTE computed once per query, and only when a member
  filter is set (a one-time filter; otherwise it is never executed).
- Small sets check blocked status per member: `NOT EXISTS (SELECT 1 FROM
  consent_blocked_user_ids(project) b WHERE b.user_id = … OFFSET 0)`. Postgres
  inlines the function; `OFFSET 0` keeps the check a per-member subplan
  (index lookups), because as an anti-join the planner may compute the whole
  project's set instead. Used by `GetTeamAverageAge`,
  `GetConsentBlockedUserIDsInTeam` (rewritten to this form) and
  `GetConsentBlockedProjectsOfUsers`.
- Measured in a throwaway container (2026-10-26, generic plans as pgx uses
  them; 8k users, 800 teams): per-member check about 0.1–0.4 ms; `teams()`
  without a member filter 4.4 ms, with one 25 ms. A first version (a
  per-team `LATERAL` count gated in its `ON`) took 13 s: the gate did not
  skip it, and the planner turned the per-member check into a whole-project
  anti-join per team. Correlated `consent_blocked_user_ids` checks must keep
  `OFFSET 0`.
- `TeamMember.countsTowardScores: Boolean!` (new): false for a blocked
  member, whom only admins see. The admin unit card (`AdminUnitCard.vue`)
  still computes average age, the 36+ badge and the member count in the
  browser from the listed members, X included; it should use this field.
  Agreed as a separate frontend PR.
- The age rule after consent changes: accept, reject, `adminSetUserConsent`
  (required consents only), `createConsent` and `updateConsent` snapshot the
  blocked (project, user) pairs before and after
  (`GetConsentBlockedProjectsOfUsers`, `GetConsentBlockedUsersInProjects`,
  `consents.sql`). Only the teams of users whose status changed are
  re-checked (`applyConsentBlockChanges`, `consents.go`), because the rule
  overwrites a hand-set `leaderboard_excluded`, which a project-wide sweep
  would reset everywhere. The duplicated rule in add/remove team members is
  now one helper (`refreshTeamLeaderboardExclusion`, `teams.go`), unchanged in
  behaviour.
- Tests: `e2e/consents_excluded_from_counts_test.go`, all through GraphQL:
  blocked, accept, reject, a relaxing new version, re-requiring it, admin
  accept; a hand-excluded control team must keep its flag. Each rule was
  broken on purpose and a subtest failed. Unit tests for `teamMembersFor`,
  `consentBlockChanges`, `teamsOfChangedUsers`.

Expected behaviour (not a gap):

- A consent whose `published_at` is in the future starts blocking when it
  is published, and a user turning 13 becomes subject to it, with no
  mutation. Boards, averages and counts follow at once (they read live). The
  stored `leaderboard_excluded` flag, as before, is recomputed only on its
  triggers (membership changes, and now consent changes).
- Publishing a required consent changes the status of every project member,
  so the age rule re-runs for every unit of the project (seen in the manual
  test, 2026-10-27). Two effects, both accepted by Jan (2026-10-28):
  - A hand-set `leaderboard_excluded` on a unit under 36 is reset to false;
    admins must set it again. Membership changes already did this.
  - A 36+ unit that never went through the rule (`joinTeam` does not run it;
    neither does seeding) is excluded from then on. This applies the rule as
    intended. In the 2026-10-27 production copy no unit was excluded by hand,
    but the generated projects held 140–1,136 such 36+ units each.

Still open from the use case: the Ladder to Heaven distribution and the
`quiz_session_finished` webhook (they send data out; needs a product
decision), push/bulk audiences, the API reject guard, and the admin unit card
(above).

### Phase 3 — user frontend ✅ done 2026-10-15 (client-side gate only)

- [x] `app/graphql/queries/blocking-consents.gql` — the gate's own minimal
      query; `isRequired` added to the consent and profile page queries.
- [x] `app/middleware/03.consent.global.ts` — root, so it runs after
      `01.auth`. Skips the auth paths, `/settings/consent` and `/admin/**`.
      Blocks on a required LOCAL consent that is pending **or rejected**.
      Fails open on a query error, so offline never locks anyone out.
- [x] `app/plugins/0.urql.ts` exports `getUrqlClient()`. Middleware runs
      outside component setup, so the Vue-provided client is unreachable;
      returning it as a Nuxt `provide` instead collapses every other plugin's
      injection type to `unknown` (it broke `useAnalytics`).
- [x] `settings/consent.vue` — close button and "Gå videre" hidden while
      blocking. Blocking wins over `hasCompletedOnboarding`.
- [x] `ConsentCard.vue` — no reject button on a required consent. A rejected
      one keeps its "Endre" button, which is the recovery path.
- [x] `test/unit/middleware-consent-global.test.ts` — 10 cases, all passing.

**Still client-side only.** Phase 2 (the API guard and the participation
filter) is untouched: a crafted GraphQL call still works, and a blocked user is
still visible in units, leaderboards and notification audiences.

#### Consolidation, 2026-10-16

The show/hide decision had grown to four predicates across three files, with
the blocking rule written twice (`.some()` in the middleware, a `Set` in the
page) and `index.vue` writing localStorage from a render-time watcher.

- [x] `app/utils/consentGate.ts` — `blockingConsentIds()` and
      `unansweredLocalConsents()`, pure and structural over the consent-status
      shape. Not a composable: route middleware cannot call one, and the two
      callers read different queries with different selections. Lives in root
      `app/utils` without a `PINNED_TO_ROOT` entry because
      `test/unit/shared-root.test.ts` exempts anything a root file references.
- [x] The middleware is the single redirect owner, with two tiers: required
      redirects from any route, unanswered-optional only from `/`. The query is
      now `ConsentGate`, not `BlockingConsents`, since it serves both.
- [x] `index.vue`'s `pendingInternal` watcher deleted; the remote-consent
      banner stays. (`stale` went unused with it.)
- [x] `consent.vue` calls the shared functions instead of its own copies.
- [x] Gate test extended to 12 cases, covering both tiers.

**Left for later (item 3):** `hasCompletedOnboarding` still does two jobs —
journey state and presentation switch — and is still per-device localStorage,
now written from the middleware rather than a watcher. Replacing it with a
derived `mode` (`'gate' | 'onboarding' | 'settings'`) needs one decision first:
whether a reload mid-onboarding should resume onboarding presentation, which is
the only thing the persistence buys.

### Phase 3 — original plan (superseded)

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
