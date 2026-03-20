defmodule ElixirBackend.Achievements.AchievementTranslation do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "achievement_translations" do
    field :language_code, :string, primary_key: true

    belongs_to :achievement, ElixirBackend.Achievements.Achievement,
      type: :string,
      primary_key: true

    field :name, :string
    field :description_pending, :string
    field :description_completed, :string
    field :notification_text, :string

    timestamps(type: :utc_datetime)
  end

  def changeset(translation, attrs) do
    translation
    |> cast(attrs, [
      :achievement_id,
      :language_code,
      :name,
      :description_pending,
      :description_completed,
      :notification_text
    ])
    |> validate_required([:achievement_id, :language_code])
    |> foreign_key_constraint(:achievement_id)
  end
end
