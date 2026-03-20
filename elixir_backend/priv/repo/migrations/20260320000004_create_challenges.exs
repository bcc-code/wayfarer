defmodule ElixirBackend.Repo.Migrations.CreateChallenges do
  use Ecto.Migration

  def change do
    create table(:challenges, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :project_id, references(:projects, type: :string, on_delete: :delete_all), null: false
      add :event_id, references(:events, type: :string, on_delete: :nilify_all)
      add :challenge_type, :string, size: 50, null: false, default: "SIMPLE"
      add :name, :string, null: false
      add :description, :text, null: false, default: ""
      add :image_url, :string, size: 500
      add :url, :string, size: 500
      add :button_text, :string, size: 100
      add :notification_text, :text, null: false, default: ""
      add :published_at, :utc_datetime
      add :visible_at, :utc_datetime
      add :started_at, :utc_datetime
      add :end_time, :utc_datetime
      add :allow_self_completion, :boolean, null: false, default: true
      add :requires_team_membership, :boolean, null: false, default: false
      add :requires_super_team_membership, :boolean, null: false, default: false
      add :plugin_challenge_id, :string, size: 100
      add :plugin_data, :map

      timestamps(type: :utc_datetime)
    end

    create index(:challenges, [:project_id])
    create index(:challenges, [:event_id])
    create index(:challenges, [:published_at])
    create index(:challenges, [:visible_at])
    create index(:challenges, [:challenge_type])

    create unique_index(:challenges, [:plugin_challenge_id],
             where: "plugin_challenge_id IS NOT NULL"
           )
  end
end
