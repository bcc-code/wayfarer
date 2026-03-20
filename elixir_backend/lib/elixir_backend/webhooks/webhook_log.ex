defmodule ElixirBackend.Webhooks.WebhookLog do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "webhook_logs" do
    field :event_type, :string
    field :request_payload, :map
    field :response_status_code, :integer
    field :response_body, :string
    field :duration_ms, :integer
    field :error_message, :string
    field :created_at, :utc_datetime

    belongs_to :webhook, ElixirBackend.Webhooks.Webhook, type: :string
  end

  def changeset(log, attrs) do
    log
    |> cast(attrs, [
      :id,
      :webhook_id,
      :event_type,
      :request_payload,
      :response_status_code,
      :response_body,
      :duration_ms,
      :error_message,
      :created_at
    ])
    |> validate_required([:id, :webhook_id, :event_type])
    |> foreign_key_constraint(:webhook_id)
  end
end
