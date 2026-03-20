defmodule ElixirBackend.Accounts do
  @moduledoc """
  Context module for user/account business logic.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.Cache
  alias ElixirBackend.Pagination
  alias ElixirBackend.Accounts.{User, UserProject, UserEvent}

  # ── Read ──

  def get_user(id) do
    Cache.fetch_user(id, fn ->
      case Repo.get(User, id) do
        nil -> {:error, :not_found}
        user -> {:ok, user}
      end
    end)
  end

  def get_user!(id), do: Repo.get!(User, id)

  @doc """
  Permission-checked user lookup.
  - Admin/superadmin/m2m: full access
  - Regular user: can only access their own record
  - Returns {:error, :not_found} for permission denial (don't leak existence)
  """
  def get_accessible_user(id, opts \\ []) do
    user_id = opts[:user_id]
    roles = opts[:roles] || []

    cond do
      has_elevated_role?(roles) ->
        get_user(id)

      user_id == id ->
        get_user(id)

      true ->
        {:error, :not_found}
    end
  end

  @doc """
  Paginated user listing with permission checks.
  - Admin/superadmin/m2m: full access with filters
  - Regular user: denied
  """
  def list_users(filter \\ %{}, pagination_opts \\ %{}, opts \\ []) do
    roles = opts[:roles] || []

    if has_elevated_role?(roles) do
      base_query = from(u in User)

      query =
        base_query
        |> apply_filter(filter)

      total_count = Repo.aggregate(query, :count)

      pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

      items =
        query
        |> Pagination.paginate(pagination_opts)
        |> Repo.all()

      {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
    else
      {:error, :unauthorized}
    end
  end

  def me(user_id), do: get_user(user_id)

  @doc "Calculate age from birthdate. Returns nil if no birthdate."
  def calculate_age(nil), do: nil

  def calculate_age(birthdate) do
    today = Date.utc_today()
    age = today.year - birthdate.year

    if Date.compare(
         %Date{year: today.year, month: birthdate.month, day: birthdate.day},
         today
       ) == :gt do
      age - 1
    else
      age
    end
  end

  # ── Write ──

  def assign_user_to_project(user_id, project_id) do
    joined_at = DateTime.utc_now() |> DateTime.truncate(:second)

    %UserProject{}
    |> UserProject.changeset(%{user_id: user_id, project_id: project_id, joined_at: joined_at})
    |> Repo.insert(on_conflict: :nothing, conflict_target: [:user_id, :project_id])
    |> case do
      {:ok, _} -> get_user(user_id)
      error -> error
    end
  end

  def remove_user_from_project(user_id, project_id) do
    query =
      from(up in UserProject,
        where: up.user_id == ^user_id and up.project_id == ^project_id
      )

    Repo.delete_all(query)
    get_user(user_id)
  end

  def assign_user_to_event(user_id, event_id) do
    joined_at = DateTime.utc_now() |> DateTime.truncate(:second)

    %UserEvent{}
    |> UserEvent.changeset(%{user_id: user_id, event_id: event_id, joined_at: joined_at})
    |> Repo.insert(on_conflict: :nothing, conflict_target: [:user_id, :event_id])
    |> case do
      {:ok, _} -> get_user(user_id)
      error -> error
    end
  end

  def lock_user_church(user_id) do
    locked_until =
      DateTime.utc_now()
      |> DateTime.add(6 * 30 * 24 * 3600)
      |> DateTime.truncate(:second)

    with {:ok, user} <- get_user(user_id) do
      result =
        user
        |> User.church_lock_changeset(%{church_locked_until: locked_until})
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.put_user(updated)
        {:ok, updated}
      end
    end
  end

  def unlock_user_church(user_id) do
    with {:ok, user} <- get_user(user_id) do
      result =
        user
        |> User.church_lock_changeset(%{church_locked_until: nil})
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.put_user(updated)
        {:ok, updated}
      end
    end
  end

  # ── Private helpers ──

  defp has_elevated_role?(roles) do
    Enum.any?(roles, &(&1 in ["admin", "superadmin", "m2m"]))
  end

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:query, text}, q when is_binary(text) and text != "" ->
        pattern = "%#{text}%"
        where(q, [u], ilike(u.name, ^pattern) or ilike(u.email, ^pattern))

      {:church_id, church_id}, q when is_binary(church_id) ->
        where(q, [u], u.church_id == ^church_id)

      {:gender, gender}, q when is_binary(gender) ->
        where(q, [u], u.gender == ^gender)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [u], u.id in ^ids)

      {:min_age, min_age}, q when is_integer(min_age) ->
        max_birthdate = Date.utc_today() |> Date.add(-min_age * 365)
        where(q, [u], not is_nil(u.birthdate) and u.birthdate <= ^max_birthdate)

      {:max_age, max_age}, q when is_integer(max_age) ->
        min_birthdate = Date.utc_today() |> Date.add(-(max_age + 1) * 365)
        where(q, [u], not is_nil(u.birthdate) and u.birthdate > ^min_birthdate)

      {:project_id, project_id}, q when is_binary(project_id) ->
        where(
          q,
          [u],
          u.id in subquery(
            from(up in UserProject,
              where: up.project_id == ^project_id,
              select: up.user_id
            )
          )
        )

      {:event_id, event_id}, q when is_binary(event_id) ->
        where(
          q,
          [u],
          u.id in subquery(
            from(ue in UserEvent,
              where: ue.event_id == ^event_id,
              select: ue.user_id
            )
          )
        )

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
