defmodule ElixirBackend.Repo.Migrations.CreateProjects do
  use Ecto.Migration

  def change do
    create table(:projects, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :name, :string, null: false
      add :start_date, :utc_datetime, null: false
      add :end_date, :utc_datetime, null: false

      timestamps(type: :utc_datetime)
    end
  end
end
