defmodule ElixirBackendWeb.Schema.ChallengeQueries do
  @moduledoc "GraphQL query resolvers for challenges."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Challenges

  object :challenge_queries do
    field :challenge, non_null(:challenge) do
      arg(:id, non_null(:id))

      resolve(fn _parent, %{id: id}, %{context: context} ->
        Challenges.get_visible_challenge(id,
          user_id: context[:current_user_id],
          roles: context[:roles] || []
        )
      end)
    end

    field :challenges, non_null(:challenge_connection) do
      arg(:filter, :challenge_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn _parent, args, %{context: context} ->
        filter = Map.get(args, :filter, %{}) || %{}

        pagination_opts =
          args
          |> Map.take([:first, :after, :last, :before])
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        result =
          Challenges.list_visible_challenges(filter, pagination_opts,
            user_id: context[:current_user_id],
            roles: context[:roles] || []
          )

        {:ok, result}
      end)
    end
  end
end
