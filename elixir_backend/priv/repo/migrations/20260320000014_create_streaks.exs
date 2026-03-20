defmodule ElixirBackend.Repo.Migrations.CreateStreaks do
  use Ecto.Migration

  def change do
    create table(:streaks, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :name, :string, null: false
      add :description, :text, null: false, default: ""

      timestamps(type: :utc_datetime)
    end

    create index(:streaks, [:project_id])

    create table(:streak_relevant_days, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :streak_id, references(:streaks, type: :string, on_delete: :delete_all), null: false
      add :start_date, :date, null: false
      add :end_date, :date, null: false
    end

    create index(:streak_relevant_days, [:streak_id])

    create table(:user_streak_activity, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :streak_id, references(:streaks, type: :string, on_delete: :delete_all), null: false
      add :activity_date, :date, null: false
      add :created_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:user_streak_activity, [:user_id, :streak_id, :activity_date])
    create index(:user_streak_activity, [:user_id])
    create index(:user_streak_activity, [:streak_id])
  end
end
