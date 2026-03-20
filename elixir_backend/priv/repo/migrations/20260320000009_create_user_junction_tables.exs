defmodule ElixirBackend.Repo.Migrations.CreateUserJunctionTables do
  use Ecto.Migration

  def change do
    create table(:user_projects, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :joined_at, :utc_datetime, null: false
    end

    create unique_index(:user_projects, [:user_id, :project_id])
    create index(:user_projects, [:project_id])

    create table(:user_events, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :event_id, references(:events, type: :string, on_delete: :delete_all), null: false
      add :joined_at, :utc_datetime, null: false
    end

    create unique_index(:user_events, [:user_id, :event_id])
    create index(:user_events, [:event_id])
  end
end
