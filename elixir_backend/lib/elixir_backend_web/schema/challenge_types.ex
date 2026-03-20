defmodule ElixirBackendWeb.Schema.ChallengeTypes do
  @moduledoc "Absinthe types for challenges: interface, concrete types, inputs, and pagination."
  use Absinthe.Schema.Notation

  import Absinthe.Resolution.Helpers, only: [on_load: 2]

  alias ElixirBackend.Translations
  alias ElixirBackendWeb.Schema.ChallengeResolvers

  # ── Enums ──

  enum :challenge_type_enum do
    value(:simple, as: "SIMPLE")
    value(:quiz, as: "QUIZ")
    value(:external, as: "EXTERNAL")
    value(:plugin, as: "PLUGIN")
  end

  # ── Interface ──

  interface :challenge do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:html)
    field :image_url, :string
    field :button_text, :string
    field :notification_text, non_null(:string)
    field :published_at, :datetime
    field :visible_at, :datetime
    field :started_at, :datetime
    field :end_time, :datetime
    field :requires_team_membership, non_null(:boolean)
    field :requires_super_team_membership, non_null(:boolean)
    field :project, non_null(:project)
    field :event, :event
    field :user_completed_at, :datetime
    field :user_enrolled_at, :datetime
    field :translation_status, non_null(list_of(non_null(:translation_field_status)))

    resolve_type(fn
      %{challenge_type: "SIMPLE"}, _ -> :simple_challenge
      %{challenge_type: "QUIZ"}, _ -> :quiz_challenge
      %{challenge_type: "EXTERNAL"}, _ -> :external_challenge
      %{challenge_type: "PLUGIN"}, _ -> :plugin_challenge
      _, _ -> nil
    end)
  end

  # ── Concrete types ──

  object :simple_challenge do
    interface(:challenge)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:html)
    field :image_url, :string
    field :button_text, non_null(:string)
    field :notification_text, non_null(:string)
    field :published_at, :datetime
    field :visible_at, :datetime
    field :started_at, :datetime
    field :end_time, :datetime
    field :requires_team_membership, non_null(:boolean)
    field :requires_super_team_membership, non_null(:boolean)
    field :allow_self_completion, non_null(:boolean)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn c, _, _ -> {:ok, Translations.translation_status(:challenge, c.id)} end)
    end

    field :project, non_null(:project) do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, c)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :project, c)}
          |> Translations.translate_result(:project, res)
        end)
      end)
    end

    field :event, :event do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :event, c)
        |> on_load(fn loader ->
          case Dataloader.get(loader, ElixirBackend.Repo, :event, c) do
            nil -> {:ok, nil}
            event -> {:ok, event} |> Translations.translate_result(:event, res)
          end
        end)
      end)
    end

    field :user_completed_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_completed_at/3)
    end

    field :user_enrolled_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_enrolled_at/3)
    end
  end

  object :quiz_challenge do
    interface(:challenge)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:html)
    field :image_url, :string
    field :button_text, non_null(:string)
    field :notification_text, non_null(:string)
    field :published_at, :datetime
    field :visible_at, :datetime
    field :started_at, :datetime
    field :end_time, :datetime
    field :requires_team_membership, non_null(:boolean)
    field :requires_super_team_membership, non_null(:boolean)
    # Skipped for spike: quiz field resolver (quiz: Quiz!)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn c, _, _ -> {:ok, Translations.translation_status(:challenge, c.id)} end)
    end

    field :project, non_null(:project) do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, c)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :project, c)}
          |> Translations.translate_result(:project, res)
        end)
      end)
    end

    field :event, :event do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :event, c)
        |> on_load(fn loader ->
          case Dataloader.get(loader, ElixirBackend.Repo, :event, c) do
            nil -> {:ok, nil}
            event -> {:ok, event} |> Translations.translate_result(:event, res)
          end
        end)
      end)
    end

    field :user_completed_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_completed_at/3)
    end

    field :user_enrolled_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_enrolled_at/3)
    end
  end

  object :external_challenge do
    interface(:challenge)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:html)
    field :image_url, :string
    field :button_text, non_null(:string)
    field :notification_text, non_null(:string)
    field :published_at, :datetime
    field :visible_at, :datetime
    field :started_at, :datetime
    field :end_time, :datetime
    field :requires_team_membership, non_null(:boolean)
    field :requires_super_team_membership, non_null(:boolean)
    field :url, non_null(:string)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn c, _, _ -> {:ok, Translations.translation_status(:challenge, c.id)} end)
    end

    field :project, non_null(:project) do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, c)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :project, c)}
          |> Translations.translate_result(:project, res)
        end)
      end)
    end

    field :event, :event do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :event, c)
        |> on_load(fn loader ->
          case Dataloader.get(loader, ElixirBackend.Repo, :event, c) do
            nil -> {:ok, nil}
            event -> {:ok, event} |> Translations.translate_result(:event, res)
          end
        end)
      end)
    end

    field :user_completed_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_completed_at/3)
    end

    field :user_enrolled_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_enrolled_at/3)
    end
  end

  object :plugin_challenge do
    interface(:challenge)

    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:html)
    field :image_url, :string
    field :button_text, :string
    field :notification_text, non_null(:string)
    field :published_at, :datetime
    field :visible_at, :datetime
    field :started_at, :datetime
    field :end_time, :datetime
    field :requires_team_membership, non_null(:boolean)
    field :requires_super_team_membership, non_null(:boolean)
    field :plugin_challenge_id, non_null(:string)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn c, _, _ -> {:ok, Translations.translation_status(:challenge, c.id)} end)
    end

    field :project, non_null(:project) do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :project, c)
        |> on_load(fn loader ->
          {:ok, Dataloader.get(loader, ElixirBackend.Repo, :project, c)}
          |> Translations.translate_result(:project, res)
        end)
      end)
    end

    field :event, :event do
      resolve(fn c, _, %{context: %{loader: loader}} = res ->
        loader
        |> Dataloader.load(ElixirBackend.Repo, :event, c)
        |> on_load(fn loader ->
          case Dataloader.get(loader, ElixirBackend.Repo, :event, c) do
            nil -> {:ok, nil}
            event -> {:ok, event} |> Translations.translate_result(:event, res)
          end
        end)
      end)
    end

    field :user_completed_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_completed_at/3)
    end

    field :user_enrolled_at, :datetime do
      resolve(&ChallengeResolvers.resolve_user_enrolled_at/3)
    end
  end

  # ── Pagination ──

  object :challenge_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:challenge)
  end

  object :challenge_connection do
    field :edges, non_null(list_of(non_null(:challenge_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  object :page_info do
    field :has_next_page, non_null(:boolean)
    field :has_previous_page, non_null(:boolean)
    field :start_cursor, :string
    field :end_cursor, :string
  end

  # ── Input types ──

  input_object :create_challenge_input do
    field :type, non_null(:challenge_type_enum)
    field :name, non_null(:string)
    field :description, :html
    field :image, :string
    field :button_text, :string
    field :notification_text, :string
    field :published_at, :datetime
    field :visible_at, :datetime
    field :end_time, :datetime
    field :requires_team_membership, :boolean
    field :requires_super_team_membership, :boolean
    field :allow_self_completion, :boolean
    field :url, :string
    field :plugin_challenge_id, :string
  end

  input_object :update_challenge_input do
    field :name, :string
    field :description, :html
    field :image, :string
    field :event_id, :id
    field :button_text, :string
    field :notification_text, :string
    field :published_at, :datetime
    field :visible_at, :datetime
    field :started_at, :datetime
    field :end_time, :datetime
    field :requires_team_membership, :boolean
    field :requires_super_team_membership, :boolean
    field :allow_self_completion, :boolean
    field :url, :string
    field :plugin_challenge_id, :string
  end

  input_object :challenge_filter do
    field :project_id, :id
    field :event_id, :id
    field :challenge_type, :challenge_type_enum
    field :ids, list_of(non_null(:id))
    field :published_after, :datetime
    field :published_before, :datetime
  end
end
