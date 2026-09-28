-- Award-path optimization for the leaderboard trigger web (bench experiment).
--
-- Problem: update_{team,church,superteam}_leaderboard(5-arg) re-run a
-- COUNT(DISTINCT ...) membership aggregate on EVERY point award — at ~700
-- finalizes/sec that is thousands of index-join aggregates per second and
-- the dominant postgres CPU cost (726% peak in goal_1x_008_run1).
--
-- Fix: member_count on the leaderboard rows is already maintained by the
-- membership triggers (migrations 00080/00081), so the award path can trust
-- the stored value: try a plain UPDATE first (uses stored member_count);
-- only when the row does not exist yet (first score) compute the COUNT and
-- INSERT (keeping ON CONFLICT for the concurrent-first-score race).
--
-- Semantics preserved: identical expressions as the originals; the COUNT
-- still happens exactly once per row lifetime instead of once per award.
-- update_person_leaderboard has no COUNT and is unchanged.
--
-- Apply on the bench box:  psql ... -f optimize_award_path.sql
-- If validated, this should become a proper goose migration.

\set ON_ERROR_STOP on

CREATE OR REPLACE FUNCTION public.update_team_leaderboard(p_team_id character, p_project_id character, p_event_id character, p_points_delta bigint, p_score_at timestamp with time zone)
 RETURNS void
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_member_count INT;
BEGIN
    IF p_project_id IS NOT NULL THEN
        UPDATE leaderboard_project_teams l SET
            total_points = l.total_points + p_points_delta,
            score = calculate_average_score(l.total_points + p_points_delta, l.member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(l.last_score_at, p_score_at), p_score_at)
        WHERE l.project_id = p_project_id AND l.team_id = p_team_id;

        IF NOT FOUND THEN
            SELECT COUNT(*)::INT INTO v_member_count
            FROM team_members tm
            WHERE tm.team_id = p_team_id;

            INSERT INTO leaderboard_project_teams (project_id, team_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (p_project_id, p_team_id, p_points_delta, v_member_count,
                    calculate_average_score(p_points_delta, v_member_count), NOW(), p_score_at)
            ON CONFLICT (project_id, team_id)
            DO UPDATE SET
                total_points = leaderboard_project_teams.total_points + p_points_delta,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_project_teams.total_points + p_points_delta, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_project_teams.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;

    IF p_event_id IS NOT NULL THEN
        UPDATE leaderboard_event_teams l SET
            total_points = l.total_points + p_points_delta,
            score = calculate_average_score(l.total_points + p_points_delta, l.member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(l.last_score_at, p_score_at), p_score_at)
        WHERE l.event_id = p_event_id AND l.team_id = p_team_id;

        IF NOT FOUND THEN
            SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
            FROM team_members tm
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = p_event_id
            WHERE tm.team_id = p_team_id;

            INSERT INTO leaderboard_event_teams (event_id, team_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (p_event_id, p_team_id, p_points_delta, v_member_count,
                    calculate_average_score(p_points_delta, v_member_count), NOW(), p_score_at)
            ON CONFLICT (event_id, team_id)
            DO UPDATE SET
                total_points = leaderboard_event_teams.total_points + p_points_delta,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_event_teams.total_points + p_points_delta, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_teams.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;
END;
$function$;

CREATE OR REPLACE FUNCTION public.update_church_leaderboard(p_church_id character, p_project_id character, p_event_id character, p_points_delta bigint, p_score_at timestamp with time zone)
 RETURNS void
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_member_count INT;
BEGIN
    IF p_project_id IS NOT NULL THEN
        UPDATE leaderboard_project_churches l SET
            total_points = l.total_points + p_points_delta,
            score = calculate_average_score(l.total_points + p_points_delta, l.member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(l.last_score_at, p_score_at), p_score_at)
        WHERE l.project_id = p_project_id AND l.church_id = p_church_id;

        IF NOT FOUND THEN
            SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
            FROM users u
            INNER JOIN user_projects up ON up.user_id = u.id AND up.project_id = p_project_id
            WHERE u.church_id = p_church_id;

            INSERT INTO leaderboard_project_churches (project_id, church_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (p_project_id, p_church_id, p_points_delta, v_member_count,
                    calculate_average_score(p_points_delta, v_member_count), NOW(), p_score_at)
            ON CONFLICT (project_id, church_id)
            DO UPDATE SET
                total_points = leaderboard_project_churches.total_points + p_points_delta,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_project_churches.total_points + p_points_delta, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_project_churches.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;

    IF p_event_id IS NOT NULL THEN
        UPDATE leaderboard_event_churches l SET
            total_points = l.total_points + p_points_delta,
            score = calculate_average_score(l.total_points + p_points_delta, l.member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(l.last_score_at, p_score_at), p_score_at)
        WHERE l.event_id = p_event_id AND l.church_id = p_church_id;

        IF NOT FOUND THEN
            SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
            FROM users u
            INNER JOIN user_events ue ON ue.user_id = u.id AND ue.event_id = p_event_id
            WHERE u.church_id = p_church_id;

            INSERT INTO leaderboard_event_churches (event_id, church_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (p_event_id, p_church_id, p_points_delta, v_member_count,
                    calculate_average_score(p_points_delta, v_member_count), NOW(), p_score_at)
            ON CONFLICT (event_id, church_id)
            DO UPDATE SET
                total_points = leaderboard_event_churches.total_points + p_points_delta,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_event_churches.total_points + p_points_delta, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_churches.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;
END;
$function$;

CREATE OR REPLACE FUNCTION public.update_superteam_leaderboard(p_super_team_id character, p_project_id character, p_event_id character, p_points_delta bigint, p_score_at timestamp with time zone)
 RETURNS void
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_member_count INT;
BEGIN
    IF p_project_id IS NOT NULL THEN
        UPDATE leaderboard_project_superteams l SET
            total_points = l.total_points + p_points_delta,
            score = calculate_average_score(l.total_points + p_points_delta, l.member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(l.last_score_at, p_score_at), p_score_at)
        WHERE l.project_id = p_project_id AND l.super_team_id = p_super_team_id;

        IF NOT FOUND THEN
            SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
            FROM team_members tm
            INNER JOIN teams t ON t.id = tm.team_id
            WHERE t.super_team_id = p_super_team_id
              AND t.project_id = p_project_id;

            INSERT INTO leaderboard_project_superteams (project_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (p_project_id, p_super_team_id, p_points_delta, v_member_count,
                    calculate_average_score(p_points_delta, v_member_count), NOW(), p_score_at)
            ON CONFLICT (project_id, super_team_id)
            DO UPDATE SET
                total_points = leaderboard_project_superteams.total_points + p_points_delta,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_project_superteams.total_points + p_points_delta, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_project_superteams.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;

    IF p_event_id IS NOT NULL THEN
        UPDATE leaderboard_event_superteams l SET
            total_points = l.total_points + p_points_delta,
            score = calculate_average_score(l.total_points + p_points_delta, l.member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(l.last_score_at, p_score_at), p_score_at)
        WHERE l.event_id = p_event_id AND l.super_team_id = p_super_team_id;

        IF NOT FOUND THEN
            SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
            FROM team_members tm
            INNER JOIN teams t ON t.id = tm.team_id
            INNER JOIN events e ON e.project_id = t.project_id AND e.id = p_event_id
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = p_event_id
            WHERE t.super_team_id = p_super_team_id;

            INSERT INTO leaderboard_event_superteams (event_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (p_event_id, p_super_team_id, p_points_delta, v_member_count,
                    calculate_average_score(p_points_delta, v_member_count), NOW(), p_score_at)
            ON CONFLICT (event_id, super_team_id)
            DO UPDATE SET
                total_points = leaderboard_event_superteams.total_points + p_points_delta,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_event_superteams.total_points + p_points_delta, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_superteams.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;
END;
$function$;
