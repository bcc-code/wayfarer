defmodule ElixirBackend.Repo.Migrations.ExpandEvents do
  use Ecto.Migration

  def change do
    alter table(:events) do
      add :description, :text, null: false, default: ""
      add :start_date, :utc_datetime, null: false, default: fragment("now()")
      add :end_date, :utc_datetime, null: false, default: fragment("now()")
    end
  end
end
