defmodule ElixirBackend.Repo.Migrations.CreateTeams do
  use Ecto.Migration

  def change do
    create table(:super_teams, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :name, :string, null: false
      add :description, :text
      add :image_url, :string
      add :color, :string, size: 7

      timestamps(type: :utc_datetime)
    end

    create index(:super_teams, [:project_id])

    create table(:teams, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :name, :string, null: false
      add :description, :text
      add :join_code, :string, size: 50, null: false
      add :super_team_id, references(:super_teams, type: :string, on_delete: :nilify_all)
      add :leaderboard_excluded, :boolean, default: false

      timestamps(type: :utc_datetime)
    end

    create unique_index(:teams, [:join_code])
    create index(:teams, [:project_id])
    create index(:teams, [:super_team_id])

    create table(:team_members, primary_key: false) do
      add :team_id, references(:teams, type: :string, on_delete: :delete_all), null: false
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :is_team_lead, :boolean, default: false
      add :joined_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:team_members, [:team_id, :user_id])
    create index(:team_members, [:user_id])
  end
end
