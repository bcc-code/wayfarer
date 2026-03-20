defmodule ElixirBackend.Events.Event do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "events" do
    field :name, :string
    belongs_to :project, ElixirBackend.Projects.Project, type: :string

    timestamps(type: :utc_datetime)
  end

  def changeset(event, attrs) do
    event
    |> cast(attrs, [:id, :name, :project_id])
    |> validate_required([:id, :name, :project_id])
    |> foreign_key_constraint(:project_id)
  end
end
