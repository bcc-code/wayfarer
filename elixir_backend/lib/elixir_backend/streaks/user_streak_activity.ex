defmodule ElixirBackend.Streaks.UserStreakActivity do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "user_streak_activity" do
    field :activity_date, :date
    field :created_at, :utc_datetime

    belongs_to :user, ElixirBackend.Accounts.User, type: :string
    belongs_to :streak, ElixirBackend.Streaks.Streak, type: :string
  end

  def changeset(activity, attrs) do
    activity
    |> cast(attrs, [:user_id, :streak_id, :activity_date, :created_at])
    |> validate_required([:user_id, :streak_id, :activity_date])
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:streak_id)
    |> unique_constraint([:user_id, :streak_id, :activity_date])
  end
end
