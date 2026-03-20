defmodule ElixirBackend.Admin do
  @moduledoc "Context for admin dashboard statistics."

  import Ecto.Query
  alias ElixirBackend.Repo

  def dashboard_stats do
    now = DateTime.utc_now()

    total_users = Repo.aggregate("users", :count)
    total_projects = Repo.aggregate("projects", :count)
    total_challenges = Repo.aggregate("challenges", :count)

    total_points =
      from(s in "score_journal", select: coalesce(sum(s.points), 0))
      |> Repo.one()

    seven_days_ago = DateTime.add(now, -7, :day)

    new_users =
      from(u in "users", where: u.inserted_at >= ^seven_days_ago)
      |> Repo.aggregate(:count)

    active_projects =
      from(p in "projects",
        where: p.start_date <= ^now and (is_nil(p.end_date) or p.end_date >= ^now)
      )
      |> Repo.aggregate(:count)

    {:ok,
     %{
       total_users: total_users,
       total_projects: total_projects,
       total_challenges: total_challenges,
       total_points_awarded: total_points,
       new_users_last_7_days: new_users,
       active_projects_count: active_projects
     }}
  end

  def church_admin_statistics(_user_id) do
    # Stub — requires joining user→church→project with team/score data
    # Full implementation depends on the specific project the church admin manages
    {:ok,
     %{
       church_id: "",
       church_name: "",
       project_id: "",
       project_name: "",
       age_groups: [],
       total_users_in_teams: 0,
       user_scores: [],
       last_updated_at: DateTime.utc_now() |> DateTime.truncate(:second)
     }}
  end
end
