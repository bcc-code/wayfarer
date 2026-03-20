defmodule ElixirBackend.Challenges.Completion do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key false

  schema "user_challenge_completions" do
    field :user_id, :string, primary_key: true
    field :challenge_id, :string, primary_key: true
    field :completed_at, :utc_datetime
  end

  def changeset(completion, attrs) do
    completion
    |> cast(attrs, [:user_id, :challenge_id, :completed_at])
    |> validate_required([:user_id, :challenge_id])
    |> unique_constraint([:user_id, :challenge_id], name: :user_challenge_completions_pkey)
    |> foreign_key_constraint(:user_id)
    |> foreign_key_constraint(:challenge_id)
  end
end
