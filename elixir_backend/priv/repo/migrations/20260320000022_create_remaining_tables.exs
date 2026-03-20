defmodule ElixirBackend.Repo.Migrations.CreateRemainingTables do
  use Ecto.Migration

  def change do
    # ── Consents ──
    create table(:consents, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :key, :string, size: 100, null: false
      add :version, :integer, null: false, default: 1
      add :title, :string, null: false
      add :short_text, :text
      add :body, :text
      add :url, :text
      add :published_at, :utc_datetime
      add :managed_by, :string, size: 50
      add :is_remote, :boolean, null: false, default: false

      timestamps(type: :utc_datetime)
    end

    create unique_index(:consents, [:key, :version])

    create table(:user_consent_history, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :user_id, references(:users, type: :string, on_delete: :delete_all), null: false
      add :consent_id, references(:consents, type: :string, on_delete: :delete_all), null: false
      add :consent_key, :string, size: 100, null: false
      add :action, :string, size: 20, null: false
      add :occurred_at, :utc_datetime, null: false, default: fragment("now()")
      add :source, :string, size: 50
      add :external_consent_id, :string
      add :external_timestamp, :utc_datetime
    end

    create index(:user_consent_history, [:user_id])
    create index(:user_consent_history, [:consent_id])

    # ── Feedback ──
    create table(:user_feedback, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :user_id, references(:users, type: :string, on_delete: :delete_all)
      add :message, :text, null: false
      add :can_contact_me, :boolean, null: false, default: false
      add :user_agent, :text
      add :platform, :string, size: 50
      add :screen_width, :integer
      add :screen_height, :integer
      add :app_version, :string, size: 50
      add :locale, :string, size: 20
      add :project_id, :string, size: 28
      add :timezone, :string, size: 50
      add :context_url, :text
      add :tags, {:array, :string}, default: []
      add :handled_at, :utc_datetime
      add :created_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create index(:user_feedback, [:user_id])
    create index(:user_feedback, [:created_at])

    # ── File Uploads ──
    create table(:file_uploads, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :filename, :string, null: false
      add :stored_filename, :string, null: false
      add :file_size, :bigint
      add :mime_type, :string, size: 100
      add :public_url, :text
      add :uploaded_by, :string, size: 28
      add :width, :integer
      add :height, :integer
      add :blurhash, :string
      add :created_at, :utc_datetime, null: false, default: fragment("now()")
    end

    create index(:file_uploads, [:uploaded_by])
  end
end
