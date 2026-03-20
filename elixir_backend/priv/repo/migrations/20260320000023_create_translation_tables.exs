defmodule ElixirBackend.Repo.Migrations.CreateTranslationTables do
  use Ecto.Migration

  def change do
    create table(:project_translations, primary_key: false) do
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :language_code, :string, null: false
      add :name, :string, size: 255
      add :description, :text
      add :rules, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:project_translations, [:project_id, :language_code])

    create table(:event_translations, primary_key: false) do
      add :event_id, references(:events, type: :string, on_delete: :delete_all), null: false
      add :language_code, :string, null: false
      add :name, :string, size: 255
      add :description, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:event_translations, [:event_id, :language_code])

    create table(:streak_translations, primary_key: false) do
      add :streak_id, references(:streaks, type: :string, on_delete: :delete_all), null: false
      add :language_code, :string, null: false
      add :name, :string, size: 255
      add :description, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:streak_translations, [:streak_id, :language_code])

    create table(:challenge_translations, primary_key: false) do
      add :challenge_id, references(:challenges, type: :string, on_delete: :delete_all),
        null: false

      add :language_code, :string, null: false
      add :name, :string, size: 255
      add :description, :text
      add :button_text, :string, size: 100
      add :notification_text, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:challenge_translations, [:challenge_id, :language_code])

    create table(:achievement_translations, primary_key: false) do
      add :achievement_id, references(:achievements, type: :string, on_delete: :delete_all),
        null: false

      add :language_code, :string, null: false
      add :name, :string, size: 255
      add :description_pending, :text
      add :description_completed, :text
      add :notification_text, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:achievement_translations, [:achievement_id, :language_code])

    create table(:consent_translations, primary_key: false) do
      add :consent_id, references(:consents, type: :string, on_delete: :delete_all), null: false
      add :language_code, :string, null: false
      add :title, :string, size: 255
      add :short_text, :text
      add :body, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:consent_translations, [:consent_id, :language_code])

    create table(:quiz_translations, primary_key: false) do
      add :quiz_id, references(:quizzes, type: :string, on_delete: :delete_all), null: false
      add :language_code, :string, null: false
      add :name, :string, size: 255
      add :description, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:quiz_translations, [:quiz_id, :language_code])

    create table(:quiz_question_translations, primary_key: false) do
      add :question_id, references(:quiz_questions, type: :string, on_delete: :delete_all),
        null: false

      add :language_code, :string, null: false
      add :question_text, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:quiz_question_translations, [:question_id, :language_code])

    create table(:quiz_answer_translations, primary_key: false) do
      add :answer_id, references(:quiz_predefined_answers, type: :string, on_delete: :delete_all),
        null: false

      add :language_code, :string, null: false
      add :answer_text, :text

      timestamps(type: :utc_datetime)
    end

    create unique_index(:quiz_answer_translations, [:answer_id, :language_code])
  end
end
