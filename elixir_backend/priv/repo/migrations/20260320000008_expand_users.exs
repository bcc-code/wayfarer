defmodule ElixirBackend.Repo.Migrations.ExpandUsers do
  use Ecto.Migration

  def change do
    alter table(:users) do
      add :members_id, :string, size: 255, null: false, default: ""
      add :email, :string, size: 255, null: false, default: ""
      add :gender, :string, size: 10, null: false, default: "UNKNOWN"
      add :birthdate, :date
      add :church_id, references(:churches, type: :string, on_delete: :nothing)
      add :church_locked_until, :utc_datetime
      add :avatar_url, :string, size: 500
      add :display_name, :string, size: 255
      add :language, :string, size: 10, default: "en"
      add :person_uuid, :uuid
    end

    create unique_index(:users, [:members_id], where: "members_id != ''")
    create index(:users, [:church_id])
    create index(:users, [:gender])
    create index(:users, [:birthdate])
  end
end
