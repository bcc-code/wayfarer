defmodule ElixirBackend.Events do
  @moduledoc """
  Context module for event business logic.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Cache
  alias ElixirBackend.Pagination
  alias ElixirBackend.Events.Event
  alias ElixirBackend.Accounts.UserEvent

  # ── Read ──

  def get_event(id) do
    Cache.fetch(Cache.event_key(id), fn ->
      case Repo.get(Event, id) do
        nil -> {:error, :not_found}
        event -> {:ok, event}
      end
    end)
  end

  def get_event!(id), do: Repo.get!(Event, id)

  def list_events(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(e in Event)

    query = apply_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  def my_events(user_id, project_id \\ nil) do
    # Only cache when no project_id filter (common case)
    if is_nil(project_id) do
      Cache.fetch_raw(Cache.events_by_user_key(user_id), fn ->
        do_my_events_query(user_id, nil)
      end)
      |> then(&{:ok, &1})
    else
      {:ok, do_my_events_query(user_id, project_id)}
    end
  end

  defp do_my_events_query(user_id, project_id) do
    query =
      from(e in Event,
        join: ue in UserEvent,
        on: ue.event_id == e.id,
        where: ue.user_id == ^user_id
      )

    query =
      if project_id do
        where(query, [e], e.project_id == ^project_id)
      else
        query
      end

    Repo.all(query)
  end

  # ── Write ──

  def create_event(project_id, attrs) do
    id = ULID.new_event_id()

    result =
      %Event{}
      |> Event.create_changeset(attrs |> Map.put(:id, id) |> Map.put(:project_id, project_id))
      |> Repo.insert()

    with {:ok, event} <- result do
      Cache.del(Cache.events_by_project_key(project_id))
      {:ok, event}
    end
  end

  def update_event(id, attrs) do
    with {:ok, event} <- get_event(id) do
      result =
        event
        |> Event.update_changeset(attrs)
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.invalidate_event(id, event.project_id)
        {:ok, updated}
      end
    end
  end

  def delete_event(id) do
    with {:ok, event} <- get_event(id) do
      result = Repo.delete(event)

      with {:ok, deleted} <- result do
        Cache.invalidate_event(id, event.project_id)
        {:ok, deleted}
      end
    end
  end

  def move_event(id, new_project_id) do
    with {:ok, event} <- get_event(id) do
      old_project_id = event.project_id

      result =
        event
        |> Ecto.Changeset.change(project_id: new_project_id)
        |> Ecto.Changeset.foreign_key_constraint(:project_id)
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.invalidate_event(id, old_project_id)
        Cache.del(Cache.events_by_project_key(new_project_id))
        {:ok, updated}
      end
    end
  end

  def join_event(user_id, event_id) do
    joined_at = DateTime.utc_now() |> DateTime.truncate(:second)

    %UserEvent{}
    |> UserEvent.changeset(%{user_id: user_id, event_id: event_id, joined_at: joined_at})
    |> Repo.insert(on_conflict: :nothing, conflict_target: [:user_id, :event_id])

    Cache.invalidate_user_memberships(user_id)
    get_event(event_id)
  end

  # ── Private helpers ──

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, project_id}, q when is_binary(project_id) ->
        where(q, [e], e.project_id == ^project_id)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [e], e.id in ^ids)

      {:start_date_after, dt}, q when not is_nil(dt) ->
        where(q, [e], e.start_date >= ^dt)

      {:start_date_before, dt}, q when not is_nil(dt) ->
        where(q, [e], e.start_date <= ^dt)

      {:end_date_after, dt}, q when not is_nil(dt) ->
        where(q, [e], e.end_date >= ^dt)

      {:end_date_before, dt}, q when not is_nil(dt) ->
        where(q, [e], e.end_date <= ^dt)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
