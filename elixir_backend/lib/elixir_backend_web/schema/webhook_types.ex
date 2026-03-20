defmodule ElixirBackendWeb.Schema.WebhookTypes do
  use Absinthe.Schema.Notation
  @moduledoc false
  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  alias ElixirBackend.Webhooks

  enum :webhook_event_type do
    value(:challenge_completed, as: "CHALLENGE_COMPLETED")
    value(:achievement_unlocked, as: "ACHIEVEMENT_UNLOCKED")
    value(:score_updated, as: "SCORE_UPDATED")
    value(:user_joined, as: "USER_JOINED")
    value(:custom, as: "CUSTOM")
  end

  object :webhook do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :url, non_null(:string)
    field :event_type, non_null(:webhook_event_type)
    field :include_user, non_null(:boolean)
    field :include_challenge, non_null(:boolean)
    field :include_achievement, non_null(:boolean)
    field :active, non_null(:boolean)
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)

    field :recent_logs, non_null(list_of(non_null(:webhook_log))) do
      resolve(fn webhook, _, _ ->
        {:ok, Webhooks.get_recent_logs(webhook.id)}
      end)
    end
  end

  object :webhook_log do
    field :id, non_null(:id)
    field :event_type, non_null(:string)
    field :request_payload, :json
    field :response_status_code, :integer
    field :response_body, :string
    field :duration_ms, :integer
    field :error_message, :string
    field :created_at, non_null(:datetime)
  end

  input_object :create_webhook_input do
    field :project_id, non_null(:id)
    field :name, non_null(:string)
    field :url, non_null(:string)
    field :event_type, non_null(:webhook_event_type)
    field :include_user, :boolean
    field :include_challenge, :boolean
    field :include_achievement, :boolean
    field :active, :boolean
    field :secret, :string
  end

  input_object :update_webhook_input do
    field :name, :string
    field :url, :string
    field :event_type, :webhook_event_type
    field :include_user, :boolean
    field :include_challenge, :boolean
    field :include_achievement, :boolean
    field :active, :boolean
    field :secret, :string
  end
end
