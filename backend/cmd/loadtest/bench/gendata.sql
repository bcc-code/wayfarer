-- Synthetic dataset at production scale.
-- Row counts from notes/front-page-performance-report.md (measured 2026-03-31).
-- Distributions are ASSUMPTIONS (Neon calibration not authorised); sensitivity
-- to the multi-team case is tested separately in the benchmark.

\timing on
\set ON_ERROR_STOP on

-- id helper: prefix + 26 chars matching ^[0-9A-Z]{26}$
CREATE OR REPLACE FUNCTION mkid(prefix text, n bigint) RETURNS char(28) AS $$
  SELECT (prefix || lpad(upper(to_hex(n)), 26, '0'))::char(28);
$$ LANGUAGE sql IMMUTABLE;

-- ---------------------------------------------------------------- churches 179
INSERT INTO churches (id, name, country, category)
SELECT mkid('CH', g), 'Church ' || g, 'NO',
       (ARRAY['S','L','XL'])[1 + (g % 3)]
FROM generate_series(1, 179) g;

-- ---------------------------------------------------------------- project 1
INSERT INTO projects (
  id, name, description, start_date, end_date,
  color_light_accent, color_light_accent_contrast, color_light_on_accent,
  color_light_background_default, color_light_background_raised, color_light_background_indent,
  color_light_text_default, color_light_text_muted, color_light_text_hint,
  color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
  color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
  color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
  color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
  color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default
) VALUES (
  mkid('PR', 1), 'Ladder to Heaven', 'Benchmark project', now() - interval '60 days', now() + interval '60 days',
  '#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000',
  '#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000','#000000'
);

-- ---------------------------------------------------------------- events 2
INSERT INTO events (id, project_id, name, description, start_date, end_date)
SELECT mkid('EV', g), mkid('PR', 1), 'Event ' || g, 'Benchmark event',
       now() - interval '10 days', now() + interval '10 days'
FROM generate_series(1, 2) g;

-- ---------------------------------------------------------------- superteams 100
INSERT INTO super_teams (id, project_id, name)
SELECT mkid('ST', g), mkid('PR', 1), 'SuperTeam ' || g
FROM generate_series(1, 100) g;

-- ---------------------------------------------------------------- teams 790
INSERT INTO teams (id, project_id, name, join_code, super_team_id)
SELECT mkid('TM', g), mkid('PR', 1), 'Team ' || g, 'JC' || lpad(g::text, 6, '0'),
       mkid('ST', 1 + (g % 100))
FROM generate_series(1, 790) g;

-- ---------------------------------------------------------------- users 13,162
INSERT INTO users (id, members_id, email, name, gender, church_id, birthdate)
SELECT mkid('US', g), 'M' || g, 'user' || g || '@example.test', 'User ' || g,
       (ARRAY['MALE','FEMALE'])[1 + (g % 2)],
       mkid('CH', 1 + (g % 179)),
       (DATE '1995-01-01' + ((g * 7) % 3650))::date
FROM generate_series(1, 13162) g;

-- ---------------------------------------------------------------- user_projects 10,588
INSERT INTO user_projects (user_id, project_id)
SELECT mkid('US', g), mkid('PR', 1) FROM generate_series(1, 10588) g;

-- ---------------------------------------------------------------- user_events ~8,000
INSERT INTO user_events (user_id, event_id)
SELECT mkid('US', g), mkid('EV', 1) FROM generate_series(1, 8000) g;

-- ---------------------------------------------------------------- team_members 6,854
-- 6,300 users in exactly 1 team; 554 users in 2 teams (exercises the 00079 loop)
INSERT INTO team_members (team_id, user_id)
SELECT mkid('TM', 1 + (g % 790)), mkid('US', g) FROM generate_series(1, 6300) g;
INSERT INTO team_members (team_id, user_id)
SELECT mkid('TM', 1 + ((g + 397) % 790)), mkid('US', g) FROM generate_series(1, 554) g
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------- consent 19,979
INSERT INTO user_consent_history (id, user_id, consent_id, action, consent_key, occurred_at)
SELECT mkid('UC', g), mkid('US', 1 + (g % 13162)), mkid('CO', 1),
       CASE WHEN g % 7 = 0 THEN 'REJECTED' ELSE 'ACCEPTED' END,
       'leaderboard_consent',
       now() - (g % 500) * interval '1 hour'
FROM generate_series(1, 19979) g;

-- ---------------------------------------------------------------- score_journal 484,492
-- Trigger OFF for bulk load; aggregates rebuilt afterwards by the schema's own function.
ALTER TABLE score_journal DISABLE TRIGGER trigger_score_journal_leaderboard;

INSERT INTO score_journal (id, project_id, user_id, event_id, points, source_type, created_at)
SELECT mkid('SJ', g), mkid('PR', 1), mkid('US', 1 + (g % 13162)),
       CASE WHEN g % 2 = 0 THEN mkid('EV', 1) ELSE NULL END,
       1 + (g % 50), 'MANUAL',
       now() - (g % 5000) * interval '1 minute'
FROM generate_series(1, 484492) g;

ALTER TABLE score_journal ENABLE TRIGGER trigger_score_journal_leaderboard;

ANALYZE;

-- Build the 8 leaderboard tables using the schema's own regenerate function
SELECT regenerate_leaderboards();

ANALYZE;

-- ---------------------------------------------------------------- report
\echo '=== row counts ==='
SELECT 'churches' t, count(*) FROM churches
UNION ALL SELECT 'users', count(*) FROM users
UNION ALL SELECT 'teams', count(*) FROM teams
UNION ALL SELECT 'super_teams', count(*) FROM super_teams
UNION ALL SELECT 'team_members', count(*) FROM team_members
UNION ALL SELECT 'user_projects', count(*) FROM user_projects
UNION ALL SELECT 'user_events', count(*) FROM user_events
UNION ALL SELECT 'consent', count(*) FROM user_consent_history
UNION ALL SELECT 'score_journal', count(*) FROM score_journal
UNION ALL SELECT 'lb_project_persons', count(*) FROM leaderboard_project_persons
UNION ALL SELECT 'lb_project_teams', count(*) FROM leaderboard_project_teams
UNION ALL SELECT 'lb_project_superteams', count(*) FROM leaderboard_project_superteams
UNION ALL SELECT 'lb_project_churches', count(*) FROM leaderboard_project_churches
UNION ALL SELECT 'lb_event_persons', count(*) FROM leaderboard_event_persons
UNION ALL SELECT 'lb_event_teams', count(*) FROM leaderboard_event_teams
ORDER BY 1;

\echo '=== teams per user distribution ==='
SELECT n_teams, count(*) AS users FROM (
  SELECT user_id, count(*) n_teams FROM team_members GROUP BY user_id
) s GROUP BY n_teams ORDER BY n_teams;

\echo '=== db size ==='
SELECT pg_size_pretty(pg_database_size('wayfarer_bench'));
