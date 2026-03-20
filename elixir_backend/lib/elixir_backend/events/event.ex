defmodule ElixirBackend.Events.Event do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "events" do
    field :name, :string
    field :description, :string, default: ""
    field :start_date, :utc_datetime
    field :end_date, :utc_datetime
    belongs_to :project, ElixirBackend.Projects.Project, type: :string

    timestamps(type: :utc_datetime)
  end

  @required_fields [:id, :name, :project_id]
  @optional_fields [:description, :start_date, :end_date]

  def changeset(event, attrs) do
    event
    |> cast(attrs, @required_fields ++ @optional_fields)
    |> validate_required(@required_fields)
    |> foreign_key_constraint(:project_id)
  end

  def create_changeset(event, attrs) do
    changeset(event, attrs)
  end

  def update_changeset(event, attrs) do
    event
    |> cast(attrs, [:name, :description, :start_date, :end_date])
  end
end
