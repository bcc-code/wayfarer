defmodule ElixirBackend.Repo.Migrations.CreateExternalContent do
  use Ecto.Migration

  def change do
    create table(:external_content, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :plan_id, :text, null: false
      add :task_id, :text, null: false
      add :content_id, :text
      add :content_type, :text, null: false
      add :published_at, :utc_datetime
      add :synced_at, :utc_datetime, null: false, default: fragment("now()")
      add :source, :text, null: false, default: "ssf"
      add :url, :text
      add :complete_by, :utc_datetime

      timestamps(type: :utc_datetime)
    end

    create unique_index(:external_content, [:plan_id, :task_id])
    create index(:external_content, [:plan_id])
    create index(:external_content, [:task_id])
    create index(:external_content, [:content_type])
    create index(:external_content, [:source])

    create table(:external_content_translations, primary_key: false) do
      add :external_content_id,
          references(:external_content, type: :string, on_delete: :delete_all),
          null: false,
          primary_key: true

      add :language_code, :string, size: 10, null: false, primary_key: true
      add :title, :text

      timestamps(type: :utc_datetime)
    end
  end
end
