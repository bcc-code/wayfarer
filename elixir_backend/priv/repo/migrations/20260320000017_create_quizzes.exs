defmodule ElixirBackend.Repo.Migrations.CreateQuizzes do
  use Ecto.Migration

  def change do
    create table(:quizzes, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :challenge_id, :string, size: 28
      add :name, :string, null: false
      add :description, :text, null: false, default: ""
      add :image_url, :string, size: 500
      add :timeout_seconds, :integer
      add :randomize_questions, :boolean, null: false, default: false
      add :reveal_correct_answers, :boolean, null: false, default: true
      add :allow_retakes, :boolean, null: false, default: false
      add :completion_points, :integer, null: false, default: 0
      add :end_time, :utc_datetime

      timestamps(type: :utc_datetime)
    end

    create index(:quizzes, [:project_id])

    create table(:quiz_questions, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :quiz_id, references(:quizzes, type: :string, on_delete: :delete_all), null: false
      add :question_type, :string, size: 50, null: false
      add :question_text, :text, null: false
      add :question_order, :integer, null: false
      add :allow_multiple_selection, :boolean, default: false
      add :min_value, :decimal
      add :max_value, :decimal
      add :step_value, :decimal
      add :points, :integer
      add :timeout_seconds, :integer
      add :betting_enabled, :boolean, null: false, default: false
      add :betting_min_percentage, :float
      add :betting_max_percentage, :float
      add :betting_min_absolute, :integer
      add :betting_max_absolute, :integer

      timestamps(type: :utc_datetime)
    end

    create index(:quiz_questions, [:quiz_id])
    create unique_index(:quiz_questions, [:quiz_id, :question_order])

    create table(:quiz_predefined_answers, primary_key: false) do
      add :id, :string, size: 28, primary_key: true

      add :question_id, references(:quiz_questions, type: :string, on_delete: :delete_all),
        null: false

      add :answer_text, :text, null: false
      add :is_correct, :boolean, null: false, default: false
      add :answer_order, :integer, null: false

      timestamps(type: :utc_datetime)
    end

    create index(:quiz_predefined_answers, [:question_id])

    create table(:quiz_ordering_items, primary_key: false) do
      add :id, :string, size: 28, primary_key: true

      add :question_id, references(:quiz_questions, type: :string, on_delete: :delete_all),
        null: false

      add :item_text, :text, null: false
      add :correct_order, :integer, null: false

      timestamps(type: :utc_datetime)
    end

    create index(:quiz_ordering_items, [:question_id])

    create table(:quiz_submissions, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :quiz_id, references(:quizzes, type: :string, on_delete: :delete_all), null: false
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :session_id, :string, size: 28
      add :started_at, :utc_datetime, null: false, default: fragment("now()")
      add :completed_at, :utc_datetime
      add :expires_at, :utc_datetime
      add :auto_submitted, :boolean, null: false, default: false
      add :question_order, :jsonb, null: false, default: "[]"
      add :score, :integer
      add :max_score, :integer
      add :points_awarded, :integer
    end

    create index(:quiz_submissions, [:quiz_id])
    create index(:quiz_submissions, [:user_id])

    create table(:quiz_responses, primary_key: false) do
      add :id, :string, size: 28, primary_key: true

      add :submission_id,
          references(:quiz_submissions, type: :string, on_delete: :delete_all),
          null: false

      add :question_id,
          references(:quiz_questions, type: :string, on_delete: :delete_all),
          null: false

      add :selected_answer_ids, :jsonb
      add :text_response, :text
      add :number_response, :decimal
      add :json_response, :jsonb
      add :submitted_order, :jsonb
      add :is_correct, :boolean
      add :points_earned, :integer
      add :answered_at, :utc_datetime, default: fragment("now()")
      add :time_spent_seconds, :integer
      add :bet_amount, :integer
    end

    create unique_index(:quiz_responses, [:submission_id, :question_id])
    create index(:quiz_responses, [:submission_id])
    create index(:quiz_responses, [:question_id])

    # Quiz sessions
    create table(:quiz_sessions, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :quiz_id, references(:quizzes, type: :string, on_delete: :delete_all), null: false
      add :name, :string
      add :state, :string, size: 50, null: false, default: "DRAFT"
      add :open_at, :utc_datetime
      add :lock_at, :utc_datetime
      add :finish_at, :utc_datetime
      add :created_by, :string, size: 28, null: false

      timestamps(type: :utc_datetime)
    end

    create index(:quiz_sessions, [:quiz_id])

    create table(:quiz_session_access, primary_key: false) do
      add :id, :string, size: 28, primary_key: true

      add :session_id, references(:quiz_sessions, type: :string, on_delete: :delete_all),
        null: false

      add :user_id, references(:users, type: :string, on_delete: :delete_all)
      add :team_id, references(:teams, type: :string, on_delete: :delete_all)
      add :church_id, references(:churches, type: :string, on_delete: :delete_all)

      timestamps(type: :utc_datetime)
    end

    create index(:quiz_session_access, [:session_id])
    create index(:quiz_session_access, [:user_id])
  end
end
