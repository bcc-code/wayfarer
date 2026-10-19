# Leaderboard full rebuild vs. incremental apply

The 8 leaderboard tables (`leaderboard_{project,event}_{persons,teams,superteams,churches}`)
are maintained two ways that must end in the same state:

- **Incremental apply**: every `score_journal` INSERT queues a delta
  (`leaderboard_apply_queue`, migration 00101). The worker drains it via
  `drain_leaderboard_apply_queue` → `apply_score_journal_delta`. Membership
  triggers (00081) move a user's existing points when they join or leave a
  team, project or event.
- **Full rebuild**: `regenerate_leaderboards()` truncates all 8 tables and
  rebuilds them from `score_journal`. Only bench/loadtest scripts and tests call
  it; production scoring never does.

`TestRegenerateLeaderboardsMatchesIncrementalApply`
(`backend/e2e/leaderboard_regenerate_test.go`) seeds data, drains the queue,
snapshots all 8 tables, rebuilds and requires identical rows.

## Fixed in 00111 (2026-10-19)

Found as "event leaderboards differ from `regenerate_leaderboards()` on seed
data" (PR #218 review notes). The rebuild disagreed with the incremental apply
in two ways:

1. **SuperTeams double-counted multi-team users.** The rebuild joined
   `team_members × score_journal`, so a user in two teams of one superteam was
   summed twice. The incremental apply counts them once (`processed_super_teams`).
   On `cmd/seed -seed 42` this was the only difference: 1 project and 3 event
   superteam rows. Event superteam scores still showed 0 on both sides, because
   the seeder creates no `user_events` (member_count 0).
2. **Net scores of zero or below were dropped** (`HAVING SUM(points) > 0`).
   The incremental apply keeps them, and lost bet stakes (negative entries
   carrying the challenge's `event_id`) make them real.

The existing `TestLeaderboardMultiTeamWithSuperTeam` rebuild subtest passed
despite bug 1: it read through the leaderboard cache filled before the rebuild.
It now also checks the table directly.

## Known remaining differences (not fixed)

- **Event points without event participation.** `apply_score_journal_delta`
  and the rebuild add a member's event points to their event team, superteam and
  church whether or not they are in `user_events`. The 00081 event trigger
  instead adds the points on join and removes them on leave, and `member_count`
  counts only participants. Scoring and then calling `joinEvent` double-counts
  incrementally. Leaving an event (no API path) or a team while not in the event
  diverges the other way. Which rule is intended is an open product decision.
- **Joining a second team of the same superteam** re-adds the user's points
  to the superteam via the 00081 team trigger. The incremental apply and the
  rebuild count them once.
- **`cmd/loadtest/bench/gendata*.sql`** disables `trigger_score_journal_leaderboard`
  for the bulk load. Since 00101 that trigger only handles UPDATE/DELETE. The
  INSERT trigger `trigger_score_journal_enqueue_apply` stays active, so the
  bulk rows are queued and applied again on top of the rebuild once a worker
  drains the queue.
- The rebuild does not touch `leaderboard_apply_queue`. Rows still queued when
  it runs are applied again afterwards.
