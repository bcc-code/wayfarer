defmodule ElixirBackendWeb.Schema.ChallengeResolvers do
  @moduledoc """
  Shared field resolvers for challenge types.
  """

  alias ElixirBackend.Challenges

  def resolve_user_completed_at(%{id: challenge_id}, _args, %{
        context: %{current_user_id: user_id}
      }) do
    Challenges.get_user_completed_at(user_id, challenge_id)
  end

  def resolve_user_completed_at(_parent, _args, _resolution) do
    {:ok, nil}
  end

  def resolve_user_enrolled_at(%{id: challenge_id}, _args, %{context: %{current_user_id: user_id}}) do
    Challenges.get_user_enrolled_at(user_id, challenge_id)
  end

  def resolve_user_enrolled_at(_parent, _args, _resolution) do
    {:ok, nil}
  end
end
