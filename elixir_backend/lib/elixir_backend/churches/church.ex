defmodule ElixirBackend.Churches.Church do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  @categories ~w(S L XL)

  schema "churches" do
    field :name, :string
    field :country, :string
    field :category, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(church, attrs) do
    church
    |> cast(attrs, [:id, :name, :country, :category])
    |> validate_required([:id, :name, :country, :category])
    |> validate_inclusion(:category, @categories)
  end

  def update_changeset(church, attrs) do
    church
    |> cast(attrs, [:name, :country, :category])
    |> validate_inclusion(:category, @categories)
  end
end
