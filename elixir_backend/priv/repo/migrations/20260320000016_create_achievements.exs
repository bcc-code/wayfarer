defmodule ElixirBackend.Repo.Migrations.CreateAchievements do
  use Ecto.Migration

  def change do
    create table(:achievements, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :achievement_type, :string, size: 50, null: false
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :event_id, references(:events, type: :string, on_delete: :nilify_all)
      add :challenge_id, :string, size: 28
      add :name, :string, null: false
      add :description_pending, :text, null: false, default: ""
      add :description_completed, :text, null: false, default: ""
      add :notification_text, :string, null: false, default: ""
      add :image_pending, :string, size: 500, null: false, default: ""
      add :image_completed, :string, size: 500, null: false, default: ""
      add :points, :integer, null: false, default: 0
      add :hidden, :boolean, null: false, default: false
      add :awardable_from, :utc_datetime
      add :sort_order, :integer, null: false, default: 0

      timestamps(type: :utc_datetime)
    end

    create index(:achievements, [:project_id])
    create index(:achievements, [:event_id])
    create index(:achievements, [:achievement_type])

    # Content achievements (extends achievements where type = CONTENT)
    create table(:content_achievements, primary_key: false) do
      add :achievement_id,
          references(:achievements, type: :string, on_delete: :delete_all),
          primary_key: true
    end

    create table(:content_achievement_items, primary_key: false) do
      add :id, :string, size: 28, primary_key: true

      add :achievement_id,
          references(:content_achievements,
            column: :achievement_id,
            type: :string,
            on_delete: :delete_all
          ),
          null: false

      add :external_content_id,
          references(:external_content, type: :string, on_delete: :delete_all),
          null: false

      add :sort_order, :integer, null: false, default: 0
    end

    create unique_index(:content_achievement_items, [:achievement_id, :external_content_id])
    create index(:content_achievement_items, [:achievement_id])
    create index(:content_achievement_items, [:external_content_id])

    # Streak achievements (extends achievements where type = STREAK)
    create table(:streak_achievements, primary_key: false) do
      add :achievement_id,
          references(:achievements, type: :string, on_delete: :delete_all),
          primary_key: true

      add :streak_id, references(:streaks, type: :string, on_delete: :delete_all), null: false
      add :needed_streak, :integer, null: false
    end

    # Quiz achievements (extends achievements where type = QUIZ)
    create table(:quiz_achievements, primary_key: false) do
      add :achievement_id,
          references(:achievements, type: :string, on_delete: :delete_all),
          primary_key: true

      add :quiz_id, :string, size: 28
      add :min_score_percentage, :integer
      add :require_completion, :boolean, null: false, default: true
    end

    # User achievements (who has earned what)
    create table(:user_achievements, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false

      add :achievement_id, references(:achievements, type: :string, on_delete: :delete_all),
        null: false

      add :achieved_at, :utc_datetime, null: false, default: fragment("now()")
      add :celebrated_at, :utc_datetime
    end

    create unique_index(:user_achievements, [:user_id, :achievement_id])
    create index(:user_achievements, [:user_id])
    create index(:user_achievements, [:achievement_id])

    # Team achievements
    create table(:team_achievements, primary_key: false) do
      add :team_id, references(:teams, type: :string, on_delete: :delete_all), null: false

      add :achievement_id, references(:achievements, type: :string, on_delete: :delete_all),
        null: false

      add :achieved_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:team_achievements, [:team_id, :achievement_id])

    # Super team achievements
    create table(:super_team_achievements, primary_key: false) do
      add :super_team_id,
          references(:super_teams, type: :string, on_delete: :delete_all),
          null: false

      add :achievement_id, references(:achievements, type: :string, on_delete: :delete_all),
        null: false

      add :achieved_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:super_team_achievements, [:super_team_id, :achievement_id])

    # User content progress (individual content item completion)
    create table(:user_content_progress, primary_key: false) do
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false

      add :achievement_id, references(:achievements, type: :string, on_delete: :delete_all),
        null: false

      add :external_content_id,
          references(:external_content, type: :string, on_delete: :delete_all),
          null: false

      add :completed_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create unique_index(:user_content_progress, [
             :user_id,
             :achievement_id,
             :external_content_id
           ])

    create index(:user_content_progress, [:user_id])
    create index(:user_content_progress, [:achievement_id])
  end
end
