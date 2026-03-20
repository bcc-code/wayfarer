defmodule ElixirBackend.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "projects" do
    field :name, :string
    field :start_date, :utc_datetime
    field :end_date, :utc_datetime

    timestamps(type: :utc_datetime)
  end

  def changeset(project, attrs) do
    project
    |> cast(attrs, [:id, :name, :start_date, :end_date])
    |> validate_required([:id, :name, :start_date, :end_date])
  end
end
