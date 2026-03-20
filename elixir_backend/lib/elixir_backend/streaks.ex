defmodule ElixirBackend.Streaks do
  @moduledoc """
  Context module for streak management and activity tracking.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Cache
  alias ElixirBackend.Pagination
  alias ElixirBackend.Streaks.{Streak, StreakRelevantDay, UserStreakActivity}

  # ── Read ──

  def get_streak(id) do
    Cache.fetch(Cache.streak_key(id), fn ->
      case Repo.get(Streak, id) do
        nil -> {:error, :not_found}
        streak -> {:ok, streak}
      end
    end)
  end

  def get_streak!(id), do: Repo.get!(Streak, id)

  def list_streaks(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(s in Streak)

    query = apply_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  def get_relevant_days(streak_id) do
    query =
      from(rd in StreakRelevantDay,
        where: rd.streak_id == ^streak_id,
        order_by: [asc: rd.start_date]
      )

    Repo.all(query)
  end

  def get_streak_status(streak_id, user_id) do
    relevant_days = get_relevant_days(streak_id)
    today = Date.utc_today()

    # Count consecutive days backwards from today where:
    # 1. The day is within a relevant range
    # 2. The user has activity for that day
    activity_dates = get_user_activity_dates(streak_id, user_id)
    count_consecutive_days(today, relevant_days, activity_dates, 0)
  end

  def get_listened_days(streak_id, user_id, last_n) do
    relevant_days = get_relevant_days(streak_id)
    activity_dates = get_user_activity_dates(streak_id, user_id)

    # Get the last N days from relevant ranges, most recent first
    all_relevant_dates =
      relevant_days
      |> Enum.flat_map(fn rd ->
        Date.range(rd.start_date, rd.end_date) |> Enum.to_list()
      end)
      |> Enum.sort({:desc, Date})
      |> Enum.take(last_n)
      |> Enum.sort(Date)

    Enum.map(all_relevant_dates, fn date ->
      %{date: date, active: MapSet.member?(activity_dates, date)}
    end)
  end

  # ── Write ──

  def create_streak(attrs) do
    id = ULID.new_streak_id()
    relevant_days_input = attrs[:relevant_days] || []

    result =
      Repo.transaction(fn ->
        streak_attrs =
          attrs
          |> Map.put(:id, id)
          |> Map.drop([:relevant_days])

        streak =
          %Streak{}
          |> Streak.changeset(streak_attrs)
          |> Repo.insert!()

        _days = insert_relevant_days(streak.id, relevant_days_input)

        streak
      end)

    with {:ok, streak} <- result do
      if streak.project_id, do: Cache.del(Cache.streaks_by_project_key(streak.project_id))
      {:ok, streak}
    end
  end

  def update_streak(id, attrs) do
    with {:ok, streak} <- get_streak(id),
         {:ok, updated} <- do_update_streak(id, streak, attrs) do
      Cache.invalidate_streak(id, streak.project_id)
      {:ok, updated}
    end
  end

  defp do_update_streak(id, streak, attrs) do
    Repo.transaction(fn ->
      updated_streak =
        streak
        |> Streak.update_changeset(Map.drop(attrs, [:relevant_days]))
        |> Repo.update!()

      if Map.has_key?(attrs, :relevant_days) do
        from(rd in StreakRelevantDay, where: rd.streak_id == ^id) |> Repo.delete_all()
        insert_relevant_days(id, attrs.relevant_days)
      end

      updated_streak
    end)
  end

  def delete_streak(id) do
    with {:ok, streak} <- get_streak(id) do
      result = Repo.delete(streak)

      with {:ok, deleted} <- result do
        Cache.invalidate_streak(id, streak.project_id)
        {:ok, deleted}
      end
    end
  end

  def record_activity(streak_id, user_id, date \\ nil) do
    date = date || Date.utc_today()
    created_at = DateTime.utc_now() |> DateTime.truncate(:second)

    %UserStreakActivity{}
    |> UserStreakActivity.changeset(%{
      user_id: user_id,
      streak_id: streak_id,
      activity_date: date,
      created_at: created_at
    })
    |> Repo.insert(on_conflict: :nothing, conflict_target: [:user_id, :streak_id, :activity_date])
  end

  # ── Private ──

  defp insert_relevant_days(streak_id, days_input) do
    Enum.map(days_input, fn day ->
      id = ULID.new_streak_relevant_day_id()

      %StreakRelevantDay{}
      |> StreakRelevantDay.changeset(%{
        id: id,
        streak_id: streak_id,
        start_date: day.start,
        end_date: day.end
      })
      |> Repo.insert!()
    end)
  end

  defp get_user_activity_dates(streak_id, user_id) do
    query =
      from(a in UserStreakActivity,
        where: a.streak_id == ^streak_id and a.user_id == ^user_id,
        select: a.activity_date
      )

    Repo.all(query) |> MapSet.new()
  end

  defp count_consecutive_days(date, relevant_days, activity_dates, count) do
    if date_in_relevant_range?(date, relevant_days) do
      if MapSet.member?(activity_dates, date) do
        count_consecutive_days(Date.add(date, -1), relevant_days, activity_dates, count + 1)
      else
        count
      end
    else
      # Skip non-relevant days and keep looking backwards
      # But if we've gone past all relevant ranges, stop
      if any_relevant_range_contains_or_before?(date, relevant_days) do
        count_consecutive_days(Date.add(date, -1), relevant_days, activity_dates, count)
      else
        count
      end
    end
  end

  defp date_in_relevant_range?(date, relevant_days) do
    Enum.any?(relevant_days, fn rd ->
      Date.compare(date, rd.start_date) in [:gt, :eq] and
        Date.compare(date, rd.end_date) in [:lt, :eq]
    end)
  end

  defp any_relevant_range_contains_or_before?(date, relevant_days) do
    Enum.any?(relevant_days, fn rd ->
      Date.compare(date, rd.start_date) in [:gt, :eq]
    end)
  end

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, project_id}, q when is_binary(project_id) ->
        where(q, [s], s.project_id == ^project_id)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [s], s.id in ^ids)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
