defmodule ElixirBackend.Repo.Migrations.CreateChurches do
  use Ecto.Migration

  def change do
    create table(:churches, primary_key: false) do
      add :id, :string, size: 28, primary_key: true
      add :name, :string, null: false
      add :country, :string, null: false
      add :category, :string, size: 2, null: false

      timestamps(type: :utc_datetime)
    end
  end
end
