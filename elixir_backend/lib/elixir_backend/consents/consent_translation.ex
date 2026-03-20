defmodule ElixirBackend.Consents.ConsentTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "consent_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :consent, ElixirBackend.Consents.Consent,
      type: :string,
      primary_key: true

    field :title, :string
    field :short_text, :string
    field :body, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:consent_id, :language_code, :title, :short_text, :body])
    |> validate_required([:consent_id, :language_code])
    |> foreign_key_constraint(:consent_id)
  end
end
