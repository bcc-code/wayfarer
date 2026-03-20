defmodule ElixirBackend.Achievements.UserAchievement do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "user_achievements" do
    field :achieved_at, :utc_datetime
    field :celebrated_at, :utc_datetime

    belongs_to :user, ElixirBackend.Accounts.User, type: :string
    belongs_to :achievement, ElixirBackend.Achievements.Achievement, type: :string
  end

  def changeset(ua, attrs) do
    ua
    |> cast(attrs, [:user_id, :achievement_id, :achieved_at, :celebrated_at])
    |> validate_required([:user_id, :achievement_id])
    |> unique_constraint([:user_id, :achievement_id])
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:achievement_id)
  end
end
