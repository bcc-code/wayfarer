defmodule ElixirBackend.ExternalContent.Translation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "external_content_translations" do
    field :language_code, :string

    belongs_to :external_content, ElixirBackend.ExternalContent.Content, type: :string

    field :title, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:external_content_id, :language_code, :title])
    |> validate_required([:external_content_id, :language_code])
    |> foreign_key_constraint(:external_content_id)
  end
end
