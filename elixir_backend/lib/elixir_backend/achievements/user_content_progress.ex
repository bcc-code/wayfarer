defmodule ElixirBackend.Achievements.UserContentProgress do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "user_content_progress" do
    field :completed_at, :utc_datetime

    belongs_to :user, ElixirBackend.Accounts.User, type: :string
    belongs_to :achievement, ElixirBackend.Achievements.Achievement, type: :string
    belongs_to :external_content, ElixirBackend.ExternalContent.Content, type: :string
  end

  def changeset(ucp, attrs) do
    ucp
    |> cast(attrs, [:user_id, :achievement_id, :external_content_id, :completed_at])
    |> validate_required([:user_id, :achievement_id, :external_content_id])
    |> unique_constraint([:user_id, :achievement_id, :external_content_id])
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:achievement_id)
    |> foreign_key_constraint(:external_content_id)
  end
end
