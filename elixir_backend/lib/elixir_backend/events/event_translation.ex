defmodule ElixirBackend.Events.EventTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "event_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :event, ElixirBackend.Events.Event,
      type: :string,
      primary_key: true

    field :name, :string
    field :description, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:event_id, :language_code, :name, :description])
    |> validate_required([:event_id, :language_code])
    |> foreign_key_constraint(:event_id)
  end
end
