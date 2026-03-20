defmodule ElixirBackend.Repo.Migrations.CreatePushNotifications do
  use Ecto.Migration

  def change do
    create table(:push_subscriptions, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :endpoint, :text, null: false
      add :p256dh_key, :text, null: false
      add :auth_key, :text, null: false
      add :user_agent, :text

      timestamps(type: :utc_datetime)
    end

    create index(:push_subscriptions, [:user_id])
    create unique_index(:push_subscriptions, [:endpoint])

    create table(:push_notification_preferences, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :notification_type, :string, size: 50, null: false
      add :enabled, :boolean, null: false, default: true

      timestamps(type: :utc_datetime)
    end

    create unique_index(:push_notification_preferences, [:user_id, :notification_type])

    create table(:push_notification_log, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :type, :string, size: 50, null: false
      add :title, :string, null: false
      add :body, :text
      add :url, :text
      add :target_criteria, :jsonb, default: "{}"
      add :sent_count, :integer, default: 0
      add :delivered_count, :integer, default: 0
      add :failed_count, :integer, default: 0
      add :created_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create index(:push_notification_log, [:type])
    create index(:push_notification_log, [:created_at])
  end
end
