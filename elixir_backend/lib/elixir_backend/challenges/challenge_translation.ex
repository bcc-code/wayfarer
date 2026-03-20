defmodule ElixirBackend.Challenges.ChallengeTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "challenge_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :challenge, ElixirBackend.Challenges.Challenge,
      type: :string,
      primary_key: true

    field :name, :string
    field :description, :string
    field :button_text, :string
    field :notification_text, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [
      :challenge_id,
      :language_code,
      :name,
      :description,
      :button_text,
      :notification_text
    ])
    |> validate_required([:challenge_id, :language_code])
    |> foreign_key_constraint(:challenge_id)
  end
end
