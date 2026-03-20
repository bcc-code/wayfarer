defmodule ElixirBackend.Projects do
  @moduledoc """
  Context module for project business logic.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Pagination
  alias ElixirBackend.Projects.Project
  alias ElixirBackend.Accounts.UserProject

  # ── Read ──

  def get_project(id) do
    case Repo.get(Project, id) do
      nil -> {:error, :not_found}
      project -> {:ok, project}
    end
  end

  def get_project!(id), do: Repo.get!(Project, id)

  def list_projects(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(p in Project)

    query = apply_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  def my_projects(user_id) do
    query =
      from(p in Project,
        join: up in UserProject,
        on: up.project_id == p.id,
        where: up.user_id == ^user_id,
        order_by: [desc: p.start_date]
      )

    {:ok, Repo.all(query)}
  end

  # ── Write ──

  def create_project(attrs) do
    id = ULID.new_project_id()

    %Project{}
    |> Project.create_changeset(Map.put(attrs, :id, id))
    |> Repo.insert()
  end

  def update_project(id, attrs) do
    with {:ok, project} <- get_project(id) do
      project
      |> Project.update_changeset(attrs)
      |> Repo.update()
    end
  end

  def delete_project(id) do
    with {:ok, project} <- get_project(id) do
      Repo.delete(project)
    end
  end

  def archive_project(id) do
    with {:ok, project} <- get_project(id) do
      project
      |> Ecto.Changeset.change(archived: true)
      |> Repo.update()
    end
  end

  def join_project(user_id, project_id) do
    joined_at = DateTime.utc_now() |> DateTime.truncate(:second)

    %UserProject{}
    |> UserProject.changeset(%{user_id: user_id, project_id: project_id, joined_at: joined_at})
    |> Repo.insert(on_conflict: :nothing, conflict_target: [:user_id, :project_id])

    get_project(project_id)
  end

  # ── Private helpers ──

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [p], p.id in ^ids)

      {:archived, archived}, q when is_boolean(archived) ->
        where(q, [p], p.archived == ^archived)

      {:start_date_after, dt}, q when not is_nil(dt) ->
        where(q, [p], p.start_date >= ^dt)

      {:start_date_before, dt}, q when not is_nil(dt) ->
        where(q, [p], p.start_date <= ^dt)

      {:end_date_after, dt}, q when not is_nil(dt) ->
        where(q, [p], p.end_date >= ^dt)

      {:end_date_before, dt}, q when not is_nil(dt) ->
        where(q, [p], p.end_date <= ^dt)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
