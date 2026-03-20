defmodule ElixirBackend.Achievements do
  @moduledoc """
  Context module for achievement management, awarding, and progress tracking.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Pagination

  alias ElixirBackend.Achievements.{
    Achievement,
    ContentAchievement,
    ContentAchievementItem,
    StreakAchievement,
    QuizAchievement,
    UserAchievement,
    UserContentProgress
  }

  # ── Read ──

  def get_achievement(id) do
    case Repo.get(Achievement, id) do
      nil -> {:error, :not_found}
      achievement -> {:ok, achievement}
    end
  end

  def get_achievement!(id), do: Repo.get!(Achievement, id)

  def list_achievements(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(a in Achievement, order_by: [asc: a.sort_order, asc: a.inserted_at])

    query = apply_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  def get_content_items(achievement_id) do
    query =
      from(ci in ContentAchievementItem,
        where: ci.achievement_id == ^achievement_id,
        order_by: [asc: ci.sort_order]
      )

    Repo.all(query)
  end

  def get_streak_achievement_data(achievement_id) do
    Repo.get(StreakAchievement, achievement_id)
  end

  def get_quiz_achievement_data(achievement_id) do
    Repo.get(QuizAchievement, achievement_id)
  end

  def get_user_achieved_at(achievement_id, user_id) do
    query =
      from(ua in UserAchievement,
        where: ua.achievement_id == ^achievement_id and ua.user_id == ^user_id,
        select: ua.achieved_at
      )

    Repo.one(query)
  end

  def get_user_celebrated_at(achievement_id, user_id) do
    query =
      from(ua in UserAchievement,
        where: ua.achievement_id == ^achievement_id and ua.user_id == ^user_id,
        select: ua.celebrated_at
      )

    Repo.one(query)
  end

  def get_completed_item_count(achievement_id, user_id) do
    query =
      from(ucp in UserContentProgress,
        where: ucp.achievement_id == ^achievement_id and ucp.user_id == ^user_id
      )

    Repo.aggregate(query, :count)
  end

  def get_user_completed_items(achievement_id, user_id) do
    query =
      from(ucp in UserContentProgress,
        join: ci in ContentAchievementItem,
        on:
          ci.achievement_id == ucp.achievement_id and
            ci.external_content_id == ucp.external_content_id,
        where: ucp.achievement_id == ^achievement_id and ucp.user_id == ^user_id,
        select: ci,
        order_by: [asc: ci.sort_order]
      )

    Repo.all(query)
  end

  def get_next_item(achievement_id, user_id) do
    completed_content_ids =
      from(ucp in UserContentProgress,
        where: ucp.achievement_id == ^achievement_id and ucp.user_id == ^user_id,
        select: ucp.external_content_id
      )

    query =
      from(ci in ContentAchievementItem,
        where: ci.achievement_id == ^achievement_id,
        where: ci.external_content_id not in subquery(completed_content_ids),
        order_by: [asc: ci.sort_order],
        limit: 1
      )

    Repo.one(query)
  end

  # ── Write: Create ──

  def create_simple_achievement(attrs) do
    create_achievement(Map.put(attrs, :achievement_type, "SIMPLE"))
  end

  def create_content_achievement(attrs) do
    items = attrs[:items] || []

    Repo.transaction(fn ->
      {:ok, achievement} = create_achievement(Map.put(attrs, :achievement_type, "CONTENT"))

      # Create content_achievements join record
      %ContentAchievement{achievement_id: achievement.id} |> Repo.insert!()

      # Create items
      Enum.each(items, fn item ->
        %ContentAchievementItem{}
        |> ContentAchievementItem.changeset(%{
          id: ULID.new_content_item_id(),
          achievement_id: achievement.id,
          external_content_id: item.external_content_id,
          sort_order: Map.get(item, :sort_order, 0)
        })
        |> Repo.insert!()
      end)

      achievement
    end)
  end

  def create_streak_achievement(attrs) do
    needed_streak = attrs[:needed_streak]
    streak_id = attrs[:streak_id]

    Repo.transaction(fn ->
      {:ok, achievement} = create_achievement(Map.put(attrs, :achievement_type, "STREAK"))

      %StreakAchievement{}
      |> StreakAchievement.changeset(%{
        achievement_id: achievement.id,
        streak_id: streak_id,
        needed_streak: needed_streak
      })
      |> Repo.insert!()

      achievement
    end)
  end

  def create_quiz_achievement(attrs) do
    quiz_id = attrs[:quiz_id]
    min_score = attrs[:min_score_percentage]
    require_completion = Map.get(attrs, :require_completion, true)

    Repo.transaction(fn ->
      {:ok, achievement} = create_achievement(Map.put(attrs, :achievement_type, "QUIZ"))

      %QuizAchievement{}
      |> QuizAchievement.changeset(%{
        achievement_id: achievement.id,
        quiz_id: quiz_id,
        min_score_percentage: min_score,
        require_completion: require_completion
      })
      |> Repo.insert!()

      achievement
    end)
  end

  # ── Write: Update ──

  def update_achievement(id, attrs) do
    with {:ok, achievement} <- get_achievement(id) do
      achievement
      |> Achievement.update_changeset(attrs)
      |> Repo.update()
    end
  end

  def update_content_achievement(id, attrs) do
    items = attrs[:items]

    Repo.transaction(fn ->
      {:ok, achievement} = update_achievement(id, Map.drop(attrs, [:items]))

      if items do
        from(ci in ContentAchievementItem, where: ci.achievement_id == ^id)
        |> Repo.delete_all()

        Enum.each(items, fn item ->
          %ContentAchievementItem{}
          |> ContentAchievementItem.changeset(%{
            id: ULID.new_content_item_id(),
            achievement_id: id,
            external_content_id: item.external_content_id,
            sort_order: Map.get(item, :sort_order, 0)
          })
          |> Repo.insert!()
        end)
      end

      achievement
    end)
  end

  def update_streak_achievement(id, attrs) do
    streak_attrs = Map.take(attrs, [:streak_id, :needed_streak])

    Repo.transaction(fn ->
      {:ok, achievement} = update_achievement(id, Map.drop(attrs, [:streak_id, :needed_streak]))

      if map_size(streak_attrs) > 0 do
        from(sa in StreakAchievement, where: sa.achievement_id == ^id)
        |> Repo.update_all(set: Enum.to_list(streak_attrs))
      end

      achievement
    end)
  end

  def update_quiz_achievement(id, attrs) do
    quiz_attrs = Map.take(attrs, [:quiz_id, :min_score_percentage, :require_completion])

    Repo.transaction(fn ->
      {:ok, achievement} =
        update_achievement(
          id,
          Map.drop(attrs, [:quiz_id, :min_score_percentage, :require_completion])
        )

      if map_size(quiz_attrs) > 0 do
        from(qa in QuizAchievement, where: qa.achievement_id == ^id)
        |> Repo.update_all(set: Enum.to_list(quiz_attrs))
      end

      achievement
    end)
  end

  # ── Write: Delete ──

  def delete_achievement(id) do
    with {:ok, achievement} <- get_achievement(id) do
      Repo.delete(achievement)
    end
  end

  # ── Write: Award/Revoke ──

  def award_achievement(user_id, achievement_id) do
    with {:ok, achievement} <- get_achievement(achievement_id) do
      # Check awardable_from
      if achievement.awardable_from do
        now = DateTime.utc_now()

        if DateTime.compare(now, achievement.awardable_from) == :lt do
          {:error, "achievement is not yet available for awarding"}
        else
          do_award(user_id, achievement)
        end
      else
        do_award(user_id, achievement)
      end
    end
  end

  def revoke_achievement(user_id, achievement_id) do
    query =
      from(ua in UserAchievement,
        where: ua.user_id == ^user_id and ua.achievement_id == ^achievement_id
      )

    case Repo.delete_all(query) do
      {count, _} when count > 0 -> {:ok, true}
      _ -> {:ok, false}
    end
  end

  def mark_celebrated(user_id, achievement_id) do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    query =
      from(ua in UserAchievement,
        where: ua.user_id == ^user_id and ua.achievement_id == ^achievement_id
      )

    case Repo.update_all(query, set: [celebrated_at: now]) do
      {count, _} when count > 0 -> {:ok, true}
      _ -> {:ok, false}
    end
  end

  # ── Write: Content Progress ──

  def mark_content_completed(user_id, external_content_id) do
    # Find all content achievements containing this external content
    achievement_ids =
      from(ci in ContentAchievementItem,
        where: ci.external_content_id == ^external_content_id,
        select: ci.achievement_id
      )
      |> Repo.all()

    now = DateTime.utc_now() |> DateTime.truncate(:second)

    achievements =
      Enum.map(achievement_ids, fn achievement_id ->
        %UserContentProgress{}
        |> UserContentProgress.changeset(%{
          user_id: user_id,
          achievement_id: achievement_id,
          external_content_id: external_content_id,
          completed_at: now
        })
        |> Repo.insert(
          on_conflict: :nothing,
          conflict_target: [:user_id, :achievement_id, :external_content_id]
        )

        # Check if achievement is now complete and auto-award
        maybe_auto_award_content(user_id, achievement_id)

        Repo.get!(Achievement, achievement_id)
      end)

    {:ok, achievements}
  end

  def unmark_content_completed(user_id, external_content_id) do
    achievement_ids =
      from(ci in ContentAchievementItem,
        where: ci.external_content_id == ^external_content_id,
        select: ci.achievement_id
      )
      |> Repo.all()

    Enum.each(achievement_ids, fn achievement_id ->
      from(ucp in UserContentProgress,
        where:
          ucp.user_id == ^user_id and ucp.achievement_id == ^achievement_id and
            ucp.external_content_id == ^external_content_id
      )
      |> Repo.delete_all()
    end)

    achievements =
      Enum.map(achievement_ids, fn aid -> Repo.get!(Achievement, aid) end)

    {:ok, achievements}
  end

  # ── Write: Reorder ──

  def reorder_achievements(project_id, achievement_ids) do
    Repo.transaction(fn ->
      achievement_ids
      |> Enum.with_index()
      |> Enum.each(fn {id, idx} ->
        from(a in Achievement,
          where: a.id == ^id and a.project_id == ^project_id
        )
        |> Repo.update_all(set: [sort_order: idx])
      end)

      from(a in Achievement,
        where: a.project_id == ^project_id,
        order_by: [asc: a.sort_order, asc: a.inserted_at]
      )
      |> Repo.all()
    end)
  end

  def link_to_challenge(achievement_id, challenge_id) do
    with {:ok, achievement} <- get_achievement(achievement_id) do
      achievement
      |> Ecto.Changeset.change(challenge_id: challenge_id)
      |> Repo.update()
    end
  end

  # ── Private ──

  defp create_achievement(attrs) do
    id = ULID.new_achievement_id()

    %Achievement{}
    |> Achievement.changeset(Map.put(attrs, :id, id))
    |> Repo.insert()
  end

  defp do_award(user_id, achievement) do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    %UserAchievement{}
    |> UserAchievement.changeset(%{
      user_id: user_id,
      achievement_id: achievement.id,
      achieved_at: now
    })
    |> Repo.insert(
      on_conflict: :nothing,
      conflict_target: [:user_id, :achievement_id]
    )

    {:ok, achievement}
  end

  defp maybe_auto_award_content(user_id, achievement_id) do
    total_items =
      from(ci in ContentAchievementItem, where: ci.achievement_id == ^achievement_id)
      |> Repo.aggregate(:count)

    completed =
      from(ucp in UserContentProgress,
        where: ucp.achievement_id == ^achievement_id and ucp.user_id == ^user_id
      )
      |> Repo.aggregate(:count)

    if total_items > 0 and completed >= total_items do
      achievement = Repo.get!(Achievement, achievement_id)
      do_award(user_id, achievement)
    end
  end

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, project_id}, q when is_binary(project_id) ->
        where(q, [a], a.project_id == ^project_id)

      {:event_id, event_id}, q when is_binary(event_id) ->
        where(q, [a], a.event_id == ^event_id)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [a], a.id in ^ids)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
