defmodule ElixirBackend.Streaks.StreakTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "streak_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :streak, ElixirBackend.Streaks.Streak,
      type: :string,
      primary_key: true

    field :name, :string
    field :description, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [:streak_id, :language_code, :name, :description])
    |> validate_required([:streak_id, :language_code])
    |> foreign_key_constraint(:streak_id)
  end
end
