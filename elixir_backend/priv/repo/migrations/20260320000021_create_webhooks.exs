defmodule ElixirBackend.Repo.Migrations.CreateWebhooks do
  use Ecto.Migration

  def change do
    create table(:webhooks, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :name, :string, null: false
      add :url, :text, null: false
      add :event_type, :string, size: 50, null: false
      add :include_user, :boolean, null: false, default: false
      add :include_challenge, :boolean, null: false, default: false
      add :include_achievement, :boolean, null: false, default: false
      add :active, :boolean, null: false, default: true
      add :secret, :string, size: 255

      timestamps(type: :utc_datetime)
    end

    create index(:webhooks, [:project_id])
    create index(:webhooks, [:event_type])

    create table(:webhook_logs, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :webhook_id, references(:webhooks, type: :string, on_delete: :delete_all), null: false
      add :event_type, :string, size: 50, null: false
      add :request_payload, :jsonb
      add :response_status_code, :integer
      add :response_body, :text
      add :duration_ms, :integer
      add :error_message, :text
      add :created_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create index(:webhook_logs, [:webhook_id])
    create index(:webhook_logs, [:created_at])
  end
end
