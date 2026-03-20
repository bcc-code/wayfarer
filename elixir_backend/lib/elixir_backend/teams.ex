defmodule ElixirBackend.Teams do
  @moduledoc """
  Context module for team and super team business logic.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Cache
  alias ElixirBackend.Pagination
  alias ElixirBackend.Teams.{Team, SuperTeam, TeamMember}

  # ── Teams: Read ──

  def get_team(id) do
    Cache.fetch(Cache.team_key(id), fn ->
      case Repo.get(Team, id) do
        nil -> {:error, :not_found}
        team -> {:ok, team}
      end
    end)
  end

  def get_team!(id), do: Repo.get!(Team, id)

  def get_team_by_join_code(code, project_id) do
    query =
      from(t in Team,
        where: t.join_code == ^code and t.project_id == ^project_id
      )

    case Repo.one(query) do
      nil -> {:ok, nil}
      team -> {:ok, team}
    end
  end

  def list_teams(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(t in Team)

    query = apply_team_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  def get_team_members(team_id) do
    query =
      from(tm in TeamMember,
        where: tm.team_id == ^team_id,
        preload: [:user]
      )

    Repo.all(query)
  end

  # ── Teams: Write ──

  def create_team(project_id, attrs) do
    id = ULID.new_team_id()
    join_code = generate_join_code()

    result =
      %Team{}
      |> Team.changeset(
        attrs
        |> Map.put(:id, id)
        |> Map.put(:project_id, project_id)
        |> Map.put(:join_code, join_code)
      )
      |> Repo.insert()

    with {:ok, team} <- result do
      Cache.del(Cache.teams_by_project_key(project_id))
      {:ok, team}
    end
  end

  def update_team(id, attrs) do
    with {:ok, team} <- get_team(id) do
      result =
        team
        |> Team.update_changeset(attrs)
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.invalidate_team(id, team.project_id)
        {:ok, updated}
      end
    end
  end

  def delete_team(id) do
    with {:ok, team} <- get_team(id) do
      result = Repo.delete(team)

      with {:ok, deleted} <- result do
        Cache.invalidate_team(id, team.project_id)
        {:ok, deleted}
      end
    end
  end

  def join_team(user_id, join_code) do
    query = from(t in Team, where: t.join_code == ^join_code)

    case Repo.one(query) do
      nil ->
        {:error, "invalid join code"}

      team ->
        add_member(team.id, user_id)
        Cache.del(Cache.team_members_key(team.id))
        Cache.invalidate_user_teams(user_id)
        {:ok, team}
    end
  end

  def add_members(team_id, user_ids) do
    joined_at = DateTime.utc_now() |> DateTime.truncate(:second)

    Enum.each(user_ids, fn user_id ->
      %TeamMember{}
      |> TeamMember.changeset(%{team_id: team_id, user_id: user_id, joined_at: joined_at})
      |> Repo.insert(on_conflict: :nothing, conflict_target: [:team_id, :user_id])
    end)

    Cache.del(Cache.team_members_key(team_id))
    Enum.each(user_ids, &Cache.invalidate_user_teams/1)
    get_team(team_id)
  end

  def remove_members(team_id, user_ids) do
    query =
      from(tm in TeamMember,
        where: tm.team_id == ^team_id and tm.user_id in ^user_ids
      )

    Repo.delete_all(query)
    Cache.del(Cache.team_members_key(team_id))
    Enum.each(user_ids, &Cache.invalidate_user_teams/1)
    get_team(team_id)
  end

  def regenerate_join_code(team_id) do
    with {:ok, team} <- get_team(team_id) do
      result =
        team
        |> Ecto.Changeset.change(join_code: generate_join_code())
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.invalidate_team(team_id, team.project_id)
        {:ok, updated}
      end
    end
  end

  def assign_team_lead(team_id, user_id) do
    query =
      from(tm in TeamMember,
        where: tm.team_id == ^team_id and tm.user_id == ^user_id
      )

    if Repo.exists?(query) do
      from(tm in TeamMember,
        where: tm.team_id == ^team_id and tm.user_id == ^user_id
      )
      |> Repo.update_all(set: [is_team_lead: true])

      Cache.del(Cache.team_members_key(team_id))
      get_team(team_id)
    else
      {:error, "user is not a member of the team"}
    end
  end

  # ── SuperTeams: Read ──

  def get_super_team(id) do
    Cache.fetch(Cache.super_team_key(id), fn ->
      case Repo.get(SuperTeam, id) do
        nil -> {:error, :not_found}
        super_team -> {:ok, super_team}
      end
    end)
  end

  def get_super_team!(id), do: Repo.get!(SuperTeam, id)

  def list_super_teams(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(st in SuperTeam)

    query = apply_super_team_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  # ── SuperTeams: Write ──

  def create_super_team(project_id, attrs) do
    id = ULID.new_super_team_id()

    result =
      %SuperTeam{}
      |> SuperTeam.changeset(
        attrs
        |> Map.put(:id, id)
        |> Map.put(:project_id, project_id)
      )
      |> Repo.insert()

    with {:ok, super_team} <- result do
      if team_ids = attrs[:team_ids] do
        assign_teams_to_super_team(super_team.id, team_ids)
      end

      {:ok, super_team}
    end
  end

  def update_super_team(id, attrs) do
    with {:ok, super_team} <- get_super_team(id) do
      result =
        super_team
        |> SuperTeam.update_changeset(attrs)
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.invalidate_super_team(id, super_team.project_id)
        {:ok, updated}
      end
    end
  end

  def delete_super_team(id) do
    with {:ok, super_team} <- get_super_team(id) do
      result = Repo.delete(super_team)

      with {:ok, deleted} <- result do
        Cache.invalidate_super_team(id, super_team.project_id)
        {:ok, deleted}
      end
    end
  end

  def assign_teams_to_super_team(super_team_id, team_ids) do
    query =
      from(t in Team,
        where: t.id in ^team_ids
      )

    Repo.update_all(query, set: [super_team_id: super_team_id])
    Cache.del(Cache.teams_by_superteam_key(super_team_id))
    Enum.each(team_ids, fn tid -> Cache.del(Cache.team_key(tid)) end)

    get_super_team(super_team_id)
  end

  # ── Private helpers ──

  defp add_member(team_id, user_id) do
    joined_at = DateTime.utc_now() |> DateTime.truncate(:second)

    %TeamMember{}
    |> TeamMember.changeset(%{team_id: team_id, user_id: user_id, joined_at: joined_at})
    |> Repo.insert(on_conflict: :nothing, conflict_target: [:team_id, :user_id])
  end

  defp generate_join_code do
    :crypto.strong_rand_bytes(6)
    |> Base.encode32(case: :upper, padding: false)
    |> binary_part(0, 8)
  end

  defp apply_team_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, project_id}, q when is_binary(project_id) ->
        where(q, [t], t.project_id == ^project_id)

      {:super_team_id, st_id}, q when is_binary(st_id) ->
        where(q, [t], t.super_team_id == ^st_id)

      {:church_id, church_id}, q when is_binary(church_id) ->
        where(
          q,
          [t],
          t.id in subquery(
            from(tm in TeamMember,
              join: u in assoc(tm, :user),
              where: u.church_id == ^church_id,
              select: tm.team_id,
              distinct: true
            )
          )
        )

      {:no_super_team, true}, q ->
        where(q, [t], is_nil(t.super_team_id))

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [t], t.id in ^ids)

      _, q ->
        q
    end)
  end

  defp apply_team_filter(query, _), do: query

  defp apply_super_team_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, project_id}, q when is_binary(project_id) ->
        where(q, [st], st.project_id == ^project_id)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [st], st.id in ^ids)

      _, q ->
        q
    end)
  end

  defp apply_super_team_filter(query, _), do: query
end
