defmodule ElixirBackend.Projects.ProjectTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "project_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :project, ElixirBackend.Projects.Project,
      type: :string,
      primary_key: true

    field :name, :string
    field :description, :string
    field :rules, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:project_id, :language_code, :name, :description, :rules])
    |> validate_required([:project_id, :language_code])
    |> foreign_key_constraint(:project_id)
  end
end
