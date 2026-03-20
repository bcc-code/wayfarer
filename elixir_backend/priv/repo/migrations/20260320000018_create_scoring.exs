defmodule ElixirBackend.Repo.Migrations.CreateScoring do
  use Ecto.Migration

  def change do
    create table(:score_journal, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :event_id, references(:events, type: :string, on_delete: :nilify_all)
      add :challenge_id, :string, size: 28
      add :points, :integer, null: false
      add :source_type, :string, size: 50, null: false
      add :source_id, :string, size: 28
      add :reason, :text
      add :awarded_by, :string, size: 28
      add :created_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create index(:score_journal, [:project_id])
    create index(:score_journal, [:user_id])
    create index(:score_journal, [:event_id])
    create index(:score_journal, [:source_type])
    create index(:score_journal, [:created_at])

    # ── Leaderboard tables (denormalized aggregates) ──

    create table(:leaderboard_project_persons, primary_key: false) do
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_project_persons, [:project_id, :user_id])

    create table(:leaderboard_project_teams, primary_key: false) do
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :team_id, references(:teams, type: :string, on_delete: :delete_all), null: false
      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_project_teams, [:project_id, :team_id])

    create table(:leaderboard_project_superteams, primary_key: false) do
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false

      add :super_team_id, references(:super_teams, type: :string, on_delete: :delete_all),
        null: false

      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_project_superteams, [:project_id, :super_team_id])

    create table(:leaderboard_project_churches, primary_key: false) do
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :church_id, references(:churches, type: :string, on_delete: :delete_all), null: false
      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_project_churches, [:project_id, :church_id])

    create table(:leaderboard_event_persons, primary_key: false) do
      add :event_id, references(:events, type: :string, on_delete: :delete_all), null: false
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_event_persons, [:event_id, :user_id])

    create table(:leaderboard_event_teams, primary_key: false) do
      add :event_id, references(:events, type: :string, on_delete: :delete_all), null: false
      add :team_id, references(:teams, type: :string, on_delete: :delete_all), null: false
      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_event_teams, [:event_id, :team_id])

    create table(:leaderboard_event_superteams, primary_key: false) do
      add :event_id, references(:events, type: :string, on_delete: :delete_all), null: false

      add :super_team_id, references(:super_teams, type: :string, on_delete: :delete_all),
        null: false

      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_event_superteams, [:event_id, :super_team_id])

    create table(:leaderboard_event_churches, primary_key: false) do
      add :event_id, references(:events, type: :string, on_delete: :delete_all), null: false
      add :church_id, references(:churches, type: :string, on_delete: :delete_all), null: false
      add :score, :bigint, null: false, default: 0
      add :last_score_at, :utc_datetime
      add :updated_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:leaderboard_event_churches, [:event_id, :church_id])
  end
end
