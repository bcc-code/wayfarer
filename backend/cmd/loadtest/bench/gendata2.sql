\timing on
\set ON_ERROR_STOP on

\echo '=== what is already loaded ==='
SELECT 'churches' t, count(*) FROM churches
UNION ALL SELECT 'projects', count(*) FROM projects
UNION ALL SELECT 'events', count(*) FROM events
UNION ALL SELECT 'super_teams', count(*) FROM super_teams
UNION ALL SELECT 'teams', count(*) FROM teams
UNION ALL SELECT 'users', count(*) FROM users
UNION ALL SELECT 'user_projects', count(*) FROM user_projects
UNION ALL SELECT 'user_events', count(*) FROM user_events
UNION ALL SELECT 'team_members', count(*) FROM team_members
UNION ALL SELECT 'consents', count(*) FROM consents
UNION ALL SELECT 'consent_history', count(*) FROM user_consent_history
UNION ALL SELECT 'score_journal', count(*) FROM score_journal
ORDER BY 1;

-- consent definition row (CN prefix)
INSERT INTO consents (id, key, title, body)
VALUES (mkid('CN', 1), 'leaderboard_consent', 'Leaderboard consent', 'Benchmark consent')
ON CONFLICT DO NOTHING;

-- consent history 19,979
INSERT INTO user_consent_history (id, user_id, consent_id, action, consent_key, occurred_at)
SELECT mkid('UC', g), mkid('US', 1 + (g % 13162)), mkid('CN', 1),
       CASE WHEN g % 7 = 0 THEN 'REJECTED' ELSE 'ACCEPTED' END,
       'leaderboard_consent',
       now() - (g % 500) * interval '1 hour'
FROM generate_series(1, 19979) g
ON CONFLICT DO NOTHING;

-- score_journal 484,492, trigger OFF for bulk load
ALTER TABLE score_journal DISABLE TRIGGER trigger_score_journal_leaderboard;

INSERT INTO score_journal (id, project_id, user_id, event_id, points, source_type, created_at)
SELECT mkid('SJ', g), mkid('PR', 1), mkid('US', 1 + (g % 13162)),
       CASE WHEN g % 2 = 0 THEN mkid('EV', 1) ELSE NULL END,
       1 + (g % 50), 'MANUAL',
       now() - (g % 5000) * interval '1 minute'
FROM generate_series(1, 484492) g
ON CONFLICT DO NOTHING;

ALTER TABLE score_journal ENABLE TRIGGER trigger_score_journal_leaderboard;

ANALYZE;

\echo '=== regenerate_leaderboards() (full rebuild of all 8 tables) ==='
SELECT regenerate_leaderboards();

ANALYZE;

\echo '=== final row counts ==='
SELECT 'score_journal' t, count(*) FROM score_journal
UNION ALL SELECT 'consent_history', count(*) FROM user_consent_history
UNION ALL SELECT 'lb_project_persons', count(*) FROM leaderboard_project_persons
UNION ALL SELECT 'lb_project_teams', count(*) FROM leaderboard_project_teams
UNION ALL SELECT 'lb_project_superteams', count(*) FROM leaderboard_project_superteams
UNION ALL SELECT 'lb_project_churches', count(*) FROM leaderboard_project_churches
UNION ALL SELECT 'lb_event_persons', count(*) FROM leaderboard_event_persons
UNION ALL SELECT 'lb_event_teams', count(*) FROM leaderboard_event_teams
UNION ALL SELECT 'lb_event_superteams', count(*) FROM leaderboard_event_superteams
UNION ALL SELECT 'lb_event_churches', count(*) FROM leaderboard_event_churches
ORDER BY 1;

\echo '=== teams per user (drives the 00079 trigger loop) ==='
SELECT n_teams, count(*) AS users FROM (
  SELECT user_id, count(*) n_teams FROM team_members GROUP BY user_id
) s GROUP BY n_teams ORDER BY n_teams;

\echo '=== members per team ==='
SELECT min(c), round(avg(c),1) AS avg, max(c) FROM (
  SELECT team_id, count(*) c FROM team_members GROUP BY team_id
) s;

\echo '=== db size ==='
SELECT pg_size_pretty(pg_database_size('wayfarer_bench'));
