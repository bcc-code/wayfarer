-- name: GetDashboardStats :one
SELECT
    (SELECT COUNT(*) FROM users)::int AS total_users,
    (SELECT COUNT(*) FROM projects WHERE archived = false)::int AS total_projects,
    (SELECT COUNT(*) FROM challenges)::int AS total_challenges,
    COALESCE((SELECT SUM(points) FROM score_journal), 0)::bigint AS total_points_awarded,
    (SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '7 days')::int AS new_users_last_7_days,
    (SELECT COUNT(*) FROM projects WHERE archived = false AND start_date <= NOW() AND end_date >= NOW())::int AS active_projects_count;
