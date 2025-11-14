-- Backfill script for leaderboard tables
-- Run this ONCE after creating the tables and triggers to populate initial data
-- This will calculate scores from existing achievements and score_adjustments

-- ==================== Project Person Leaderboards ====================

INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at)
SELECT
    up.project_id,
    up.user_id,
    COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) AS score,
    NOW() AS updated_at
FROM user_projects up
LEFT JOIN user_achievements ua ON ua.user_id = up.user_id
LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.project_id = up.project_id
LEFT JOIN score_adjustments sa ON sa.entity_type = 'USER' AND sa.entity_id = up.user_id AND sa.project_id = up.project_id
GROUP BY up.project_id, up.user_id
HAVING COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) > 0
ON CONFLICT (project_id, user_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;

-- ==================== Project Team Leaderboards ====================

INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at)
SELECT
    t.project_id,
    t.id AS team_id,
    -- Sum of user achievements from team members
    COALESCE(SUM(a_user.points), 0) +
    -- Plus team-specific achievements
    COALESCE(SUM(a_team.points), 0) +
    -- Plus team score adjustments
    COALESCE(SUM(sa.points), 0) AS score,
    NOW() AS updated_at
FROM teams t
LEFT JOIN team_members tm ON tm.team_id = t.id
LEFT JOIN user_achievements ua ON ua.user_id = tm.user_id
LEFT JOIN achievements a_user ON ua.achievement_id = a_user.id AND a_user.project_id = t.project_id
LEFT JOIN team_achievements ta ON ta.team_id = t.id
LEFT JOIN achievements a_team ON ta.achievement_id = a_team.id AND a_team.project_id = t.project_id
LEFT JOIN score_adjustments sa ON sa.entity_type = 'TEAM' AND sa.entity_id = t.id AND sa.project_id = t.project_id
GROUP BY t.project_id, t.id
HAVING COALESCE(SUM(a_user.points), 0) + COALESCE(SUM(a_team.points), 0) + COALESCE(SUM(sa.points), 0) > 0
ON CONFLICT (project_id, team_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;

-- ==================== Project SuperTeam Leaderboards ====================

INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at)
SELECT
    st.project_id,
    st.id AS super_team_id,
    -- Sum of user achievements from all team members in superteam
    COALESCE(SUM(a_user.points), 0) +
    -- Plus team-specific achievements from teams in superteam
    COALESCE(SUM(a_team.points), 0) +
    -- Plus superteam-specific achievements
    COALESCE(SUM(a_st.points), 0) +
    -- Plus superteam score adjustments
    COALESCE(SUM(sa.points), 0) AS score,
    NOW() AS updated_at
FROM super_teams st
LEFT JOIN teams t ON t.super_team_id = st.id
LEFT JOIN team_members tm ON tm.team_id = t.id
LEFT JOIN user_achievements ua ON ua.user_id = tm.user_id
LEFT JOIN achievements a_user ON ua.achievement_id = a_user.id AND a_user.project_id = st.project_id
LEFT JOIN team_achievements ta ON ta.team_id = t.id
LEFT JOIN achievements a_team ON ta.achievement_id = a_team.id AND a_team.project_id = st.project_id
LEFT JOIN super_team_achievements sta ON sta.super_team_id = st.id
LEFT JOIN achievements a_st ON sta.achievement_id = a_st.id AND a_st.project_id = st.project_id
LEFT JOIN score_adjustments sa ON sa.entity_type = 'SUPER_TEAM' AND sa.entity_id = st.id AND sa.project_id = st.project_id
GROUP BY st.project_id, st.id
HAVING COALESCE(SUM(a_user.points), 0) + COALESCE(SUM(a_team.points), 0) + COALESCE(SUM(a_st.points), 0) + COALESCE(SUM(sa.points), 0) > 0
ON CONFLICT (project_id, super_team_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;

-- ==================== Project Church Leaderboards ====================

INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at)
SELECT
    up.project_id,
    u.church_id,
    COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) AS score,
    NOW() AS updated_at
FROM users u
INNER JOIN user_projects up ON up.user_id = u.id
LEFT JOIN user_achievements ua ON ua.user_id = u.id
LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.project_id = up.project_id
LEFT JOIN score_adjustments sa ON sa.entity_type = 'USER' AND sa.entity_id = u.id AND sa.project_id = up.project_id
WHERE u.church_id IS NOT NULL
GROUP BY up.project_id, u.church_id
HAVING COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) > 0
ON CONFLICT (project_id, church_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;

-- ==================== Event Person Leaderboards ====================

INSERT INTO leaderboard_event_persons (event_id, user_id, score, updated_at)
SELECT
    ue.event_id,
    ue.user_id,
    COALESCE(SUM(a.points), 0) AS score,
    NOW() AS updated_at
FROM user_events ue
LEFT JOIN user_achievements ua ON ua.user_id = ue.user_id
LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.event_id = ue.event_id
GROUP BY ue.event_id, ue.user_id
HAVING COALESCE(SUM(a.points), 0) > 0
ON CONFLICT (event_id, user_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;

-- ==================== Event Team Leaderboards ====================

INSERT INTO leaderboard_event_teams (event_id, team_id, score, updated_at)
SELECT
    e.id AS event_id,
    t.id AS team_id,
    -- Sum of user achievements from team members
    COALESCE(SUM(a_user.points), 0) +
    -- Plus team-specific achievements
    COALESCE(SUM(a_team.points), 0) AS score,
    NOW() AS updated_at
FROM events e
CROSS JOIN teams t
INNER JOIN team_members tm ON tm.team_id = t.id
INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = e.id
LEFT JOIN user_achievements ua ON ua.user_id = tm.user_id
LEFT JOIN achievements a_user ON ua.achievement_id = a_user.id AND a_user.event_id = e.id
LEFT JOIN team_achievements ta ON ta.team_id = t.id
LEFT JOIN achievements a_team ON ta.achievement_id = a_team.id AND a_team.event_id = e.id
WHERE t.project_id = e.project_id
GROUP BY e.id, t.id
HAVING COALESCE(SUM(a_user.points), 0) + COALESCE(SUM(a_team.points), 0) > 0
ON CONFLICT (event_id, team_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;

-- ==================== Event SuperTeam Leaderboards ====================

INSERT INTO leaderboard_event_superteams (event_id, super_team_id, score, updated_at)
SELECT
    e.id AS event_id,
    st.id AS super_team_id,
    -- Sum of user achievements from all team members in superteam
    COALESCE(SUM(a_user.points), 0) +
    -- Plus team-specific achievements from teams in superteam
    COALESCE(SUM(a_team.points), 0) +
    -- Plus superteam-specific achievements
    COALESCE(SUM(a_st.points), 0) AS score,
    NOW() AS updated_at
FROM events e
CROSS JOIN super_teams st
INNER JOIN teams t ON t.super_team_id = st.id
INNER JOIN team_members tm ON tm.team_id = t.id
INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = e.id
LEFT JOIN user_achievements ua ON ua.user_id = tm.user_id
LEFT JOIN achievements a_user ON ua.achievement_id = a_user.id AND a_user.event_id = e.id
LEFT JOIN team_achievements ta ON ta.team_id = t.id
LEFT JOIN achievements a_team ON ta.achievement_id = a_team.id AND a_team.event_id = e.id
LEFT JOIN super_team_achievements sta ON sta.super_team_id = st.id
LEFT JOIN achievements a_st ON sta.achievement_id = a_st.id AND a_st.event_id = e.id
WHERE st.project_id = e.project_id
GROUP BY e.id, st.id
HAVING COALESCE(SUM(a_user.points), 0) + COALESCE(SUM(a_team.points), 0) + COALESCE(SUM(a_st.points), 0) > 0
ON CONFLICT (event_id, super_team_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;

-- ==================== Event Church Leaderboards ====================

INSERT INTO leaderboard_event_churches (event_id, church_id, score, updated_at)
SELECT
    ue.event_id,
    u.church_id,
    COALESCE(SUM(a.points), 0) AS score,
    NOW() AS updated_at
FROM users u
INNER JOIN user_events ue ON ue.user_id = u.id
LEFT JOIN user_achievements ua ON ua.user_id = u.id
LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.event_id = ue.event_id
WHERE u.church_id IS NOT NULL
GROUP BY ue.event_id, u.church_id
HAVING COALESCE(SUM(a.points), 0) > 0
ON CONFLICT (event_id, church_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at;
