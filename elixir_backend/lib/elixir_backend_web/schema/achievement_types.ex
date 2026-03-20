defmodule ElixirBackendWeb.Schema.AchievementTypes do
  use Absinthe.Schema.Notation
  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  alias ElixirBackend.Achievements

  # ── Interface ──

  interface :achievement do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, non_null(:string)
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :project, non_null(:project)
    field :event, :event
    field :achieved_at, :datetime
    field :celebrated_at, :datetime

    resolve_type(fn
      %{achievement_type: "SIMPLE"}, _ -> :simple_achievement
      %{achievement_type: "CONTENT"}, _ -> :content_achievement
      %{achievement_type: "STREAK"}, _ -> :streak_achievement
      %{achievement_type: "QUIZ"}, _ -> :quiz_achievement
      _, _ -> :simple_achievement
    end)
  end

  # ── Concrete Types ──

  object :simple_achievement do
    interface(:achievement)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, non_null(:string)
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)
    field :event, :event, resolve: dataloader(ElixirBackend.Repo)

    field :achieved_at, :datetime do
      resolve(&resolve_achieved_at/3)
    end

    field :celebrated_at, :datetime do
      resolve(&resolve_celebrated_at/3)
    end

    field :image_pending_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_pending}} end)
    end

    field :image_completed_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_completed}} end)
    end
  end

  object :content_achievement do
    interface(:achievement)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, non_null(:string)
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)
    field :event, :event, resolve: dataloader(ElixirBackend.Repo)

    field :achieved_at, :datetime do
      resolve(&resolve_achieved_at/3)
    end

    field :celebrated_at, :datetime do
      resolve(&resolve_celebrated_at/3)
    end

    field :image_pending_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_pending}} end)
    end

    field :image_completed_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_completed}} end)
    end

    field :items, non_null(list_of(non_null(:content_item))) do
      resolve(fn achievement, _, _ ->
        {:ok, Achievements.get_content_items(achievement.id)}
      end)
    end

    field :user_completed_items, non_null(list_of(non_null(:content_item))) do
      resolve(fn achievement, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, []}
          user_id -> {:ok, Achievements.get_user_completed_items(achievement.id, user_id)}
        end
      end)
    end

    field :next_item, :content_item do
      resolve(fn achievement, _, %{context: context} ->
        case context[:current_user_id] do
          nil ->
            items = Achievements.get_content_items(achievement.id)
            {:ok, List.first(items)}

          user_id ->
            {:ok, Achievements.get_next_item(achievement.id, user_id)}
        end
      end)
    end

    field :total_items, non_null(:integer) do
      resolve(fn achievement, _, _ ->
        {:ok, length(Achievements.get_content_items(achievement.id))}
      end)
    end

    field :completed_item_count, non_null(:integer) do
      resolve(fn achievement, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, 0}
          user_id -> {:ok, Achievements.get_completed_item_count(achievement.id, user_id)}
        end
      end)
    end
  end

  object :streak_achievement do
    interface(:achievement)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, non_null(:string)
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)
    field :event, :event, resolve: dataloader(ElixirBackend.Repo)

    field :achieved_at, :datetime do
      resolve(&resolve_achieved_at/3)
    end

    field :celebrated_at, :datetime do
      resolve(&resolve_celebrated_at/3)
    end

    field :image_pending_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_pending}} end)
    end

    field :image_completed_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_completed}} end)
    end

    field :needed_streak, non_null(:integer) do
      resolve(fn achievement, _, _ ->
        case Achievements.get_streak_achievement_data(achievement.id) do
          nil -> {:ok, 0}
          sa -> {:ok, sa.needed_streak}
        end
      end)
    end

    field :streak, non_null(:streak) do
      resolve(fn achievement, _, _ ->
        case Achievements.get_streak_achievement_data(achievement.id) do
          nil -> {:error, "streak data not found"}
          sa -> ElixirBackend.Streaks.get_streak(sa.streak_id)
        end
      end)
    end
  end

  object :quiz_achievement do
    interface(:achievement)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, non_null(:string)
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)
    field :event, :event, resolve: dataloader(ElixirBackend.Repo)

    field :achieved_at, :datetime do
      resolve(&resolve_achieved_at/3)
    end

    field :celebrated_at, :datetime do
      resolve(&resolve_celebrated_at/3)
    end

    field :image_pending_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_pending}} end)
    end

    field :image_completed_object, non_null(:image) do
      resolve(fn a, _, _ -> {:ok, %{url: a.image_completed}} end)
    end

    field :quiz, :quiz do
      resolve(fn _achievement, _, _ ->
        # Quiz type not yet implemented
        {:ok, nil}
      end)
    end

    field :min_score_percentage, :integer do
      resolve(fn achievement, _, _ ->
        case Achievements.get_quiz_achievement_data(achievement.id) do
          nil -> {:ok, nil}
          qa -> {:ok, qa.min_score_percentage}
        end
      end)
    end

    field :require_completion, non_null(:boolean) do
      resolve(fn achievement, _, _ ->
        case Achievements.get_quiz_achievement_data(achievement.id) do
          nil -> {:ok, true}
          qa -> {:ok, qa.require_completion}
        end
      end)
    end
  end

  # ── Supporting Types ──

  # Placeholder until Phase 7 (Quizzes)
  object :quiz do
    field :id, non_null(:id)
  end

  object :content_item do
    field :id, non_null(:id)
    field :sort_order, non_null(:integer)

    field :external_content, non_null(:external_content) do
      resolve(fn item, _, _ ->
        ElixirBackend.ExternalContent.get_content(item.external_content_id)
      end)
    end
  end

  object :image do
    field :url, non_null(:string)
    field :width, :integer
    field :height, :integer
    field :blurhash, :string
  end

  object :recalculate_result do
    field :awarded_count, non_null(:integer)
    field :user_ids, non_null(list_of(non_null(:id)))
  end

  # ── Input Types ──

  input_object :create_simple_achievement_input do
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, :string
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :project_id, non_null(:id)
    field :event_id, :id
    field :challenge_id, :id
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
  end

  input_object :create_content_achievement_input do
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, :string
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :project_id, non_null(:id)
    field :event_id, :id
    field :challenge_id, :id
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :items, non_null(list_of(non_null(:content_item_input)))
  end

  input_object :create_streak_achievement_input do
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, :string
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :project_id, non_null(:id)
    field :event_id, :id
    field :challenge_id, :id
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :needed_streak, non_null(:integer)
    field :streak_id, non_null(:id)
  end

  input_object :create_quiz_achievement_input do
    field :name, non_null(:string)
    field :description_pending, non_null(:string)
    field :description_completed, non_null(:string)
    field :notification_text, :string
    field :image_pending, non_null(:string)
    field :image_completed, non_null(:string)
    field :project_id, non_null(:id)
    field :challenge_id, :id
    field :points, non_null(:integer)
    field :hidden, non_null(:boolean)
    field :awardable_from, :datetime
    field :quiz_id, non_null(:id)
    field :min_score_percentage, :integer
    field :require_completion, non_null(:boolean)
  end

  input_object :update_achievement_input do
    field :name, :string
    field :description_pending, :string
    field :description_completed, :string
    field :notification_text, :string
    field :image_pending, :string
    field :image_completed, :string
    field :event_id, :id
    field :challenge_id, :id
    field :points, :integer
    field :hidden, :boolean
    field :awardable_from, :datetime
  end

  input_object :update_content_achievement_input do
    field :name, :string
    field :description_pending, :string
    field :description_completed, :string
    field :notification_text, :string
    field :image_pending, :string
    field :image_completed, :string
    field :event_id, :id
    field :challenge_id, :id
    field :points, :integer
    field :hidden, :boolean
    field :awardable_from, :datetime
    field :items, list_of(non_null(:content_item_input))
  end

  input_object :update_streak_achievement_input do
    field :name, :string
    field :description_pending, :string
    field :description_completed, :string
    field :notification_text, :string
    field :image_pending, :string
    field :image_completed, :string
    field :event_id, :id
    field :challenge_id, :id
    field :points, :integer
    field :hidden, :boolean
    field :awardable_from, :datetime
    field :needed_streak, :integer
    field :streak_id, :id
  end

  input_object :update_quiz_achievement_input do
    field :name, :string
    field :description_pending, :string
    field :description_completed, :string
    field :notification_text, :string
    field :image_pending, :string
    field :image_completed, :string
    field :event_id, :id
    field :points, :integer
    field :hidden, :boolean
    field :awardable_from, :datetime
    field :quiz_id, :id
    field :min_score_percentage, :integer
    field :require_completion, :boolean
  end

  input_object :content_item_input do
    field :external_content_id, non_null(:id)
  end

  input_object :achievement_filter do
    field :project_id, :id
    field :event_id, :id
    field :ids, list_of(non_null(:id))
  end

  # Pagination
  object :achievement_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:achievement)
  end

  object :achievement_connection do
    field :edges, non_null(list_of(non_null(:achievement_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  # ── Shared Resolvers ──

  defp resolve_achieved_at(achievement, _, %{context: context}) do
    case context[:current_user_id] do
      nil -> {:ok, nil}
      user_id -> {:ok, Achievements.get_user_achieved_at(achievement.id, user_id)}
    end
  end

  defp resolve_celebrated_at(achievement, _, %{context: context}) do
    case context[:current_user_id] do
      nil -> {:ok, nil}
      user_id -> {:ok, Achievements.get_user_celebrated_at(achievement.id, user_id)}
    end
  end
end
