defmodule ElixirBackend.Achievements.ContentAchievementItem do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "content_achievement_items" do
    field :sort_order, :integer, default: 0

    belongs_to :achievement, ElixirBackend.Achievements.Achievement, type: :string
    belongs_to :external_content, ElixirBackend.ExternalContent.Content, type: :string
  end

  def changeset(item, attrs) do
    item
    |> cast(attrs, [:id, :achievement_id, :external_content_id, :sort_order])
    |> validate_required([:id, :achievement_id, :external_content_id])
    |> unique_constraint([:achievement_id, :external_content_id])
    |> foreign_key_constraint(:achievement_id)
    |> foreign_key_constraint(:external_content_id)
  end
end
