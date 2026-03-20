defmodule ElixirBackendWeb.Schema.ChallengeMutations do
  @moduledoc "GraphQL mutation resolvers for challenges."
  use Absinthe.Schema.Notation

  alias ElixirBackend.Challenges
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :challenge_mutations do
    field :create_challenge, non_null(:challenge) do
      arg(:project_id, non_null(:id))
      arg(:event_id, :id)
      arg(:input, non_null(:create_challenge_input))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{project_id: project_id, input: input} = args, _resolution ->
        attrs =
          input
          |> Map.put(:project_id, project_id)
          |> Map.put(:event_id, args[:event_id])
          |> Map.put(:challenge_type, input.type)
          |> Map.put(:image_url, input[:image])
          |> Map.delete(:type)
          |> Map.delete(:image)

        Challenges.create_challenge(attrs)
      end)
    end

    field :update_challenge, non_null(:challenge) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_challenge_input))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id, input: input}, _resolution ->
        attrs =
          input
          |> maybe_rename(:image, :image_url)

        Challenges.update_challenge(id, attrs)
      end)
    end

    field :delete_challenge, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id}, _resolution ->
        case Challenges.delete_challenge(id) do
          {:ok, _} -> {:ok, true}
          {:error, _} -> {:error, "failed to delete challenge"}
        end
      end)
    end

    field :publish_challenge, non_null(:challenge) do
      arg(:id, non_null(:id))
      arg(:published_at, non_null(:datetime))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id, published_at: published_at}, _resolution ->
        Challenges.publish_challenge(id, published_at)
      end)
    end

    field :assign_challenge_to_event, non_null(:challenge) do
      arg(:challenge_id, non_null(:id))
      arg(:event_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{challenge_id: challenge_id, event_id: event_id}, _resolution ->
        Challenges.assign_challenge_to_event(challenge_id, event_id)
      end)
    end

    field :set_challenge_visibility, non_null(:challenge) do
      arg(:id, non_null(:id))
      arg(:visible_at, non_null(:datetime))
      arg(:started_at, :datetime)

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id, visible_at: visible_at} = args, _resolution ->
        Challenges.set_challenge_visibility(id, visible_at, args[:started_at])
      end)
    end

    field :set_challenge_requirements, non_null(:challenge) do
      arg(:id, non_null(:id))
      arg(:requires_team_membership, :boolean)
      arg(:requires_super_team_membership, :boolean)

      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{id: id} = args, _resolution ->
        Challenges.set_challenge_requirements(id,
          requires_team_membership: args[:requires_team_membership],
          requires_super_team_membership: args[:requires_super_team_membership]
        )
      end)
    end

    field :enroll_in_challenge, non_null(:challenge) do
      arg(:challenge_id, non_null(:id))

      resolve(fn _parent, %{challenge_id: challenge_id}, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            Challenges.enroll_in_challenge(user_id, challenge_id)

          _ ->
            {:error, "authentication required"}
        end
      end)
    end

    field :unenroll_from_challenge, non_null(:boolean) do
      arg(:challenge_id, non_null(:id))

      resolve(fn _parent, %{challenge_id: challenge_id}, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            case Challenges.unenroll_from_challenge(user_id, challenge_id) do
              {:ok, result} -> {:ok, result}
              error -> error
            end

          _ ->
            {:error, "authentication required"}
        end
      end)
    end

    field :complete_challenge, non_null(:challenge) do
      arg(:user_id, non_null(:id))
      arg(:challenge_id, non_null(:id))
      arg(:completed_at, :datetime)

      middleware(RequireRole, roles: ["admin", "superadmin", "m2m"])

      resolve(fn _parent, %{user_id: user_id, challenge_id: challenge_id} = args, _resolution ->
        Challenges.admin_complete_challenge(user_id, challenge_id, args[:completed_at])
        Challenges.get_challenge(challenge_id)
      end)
    end

    field :uncomplete_challenge, non_null(:boolean) do
      arg(:user_id, non_null(:id))
      arg(:challenge_id, non_null(:id))

      middleware(RequireRole, roles: ["admin", "superadmin", "m2m"])

      resolve(fn _parent, %{user_id: user_id, challenge_id: challenge_id}, _resolution ->
        case Challenges.uncomplete_challenge(user_id, challenge_id) do
          {:ok, result} -> {:ok, result}
          error -> error
        end
      end)
    end

    field :self_complete_challenge, non_null(:simple_challenge) do
      arg(:challenge_id, non_null(:id))

      resolve(fn _parent, %{challenge_id: challenge_id}, %{context: context} ->
        case context do
          %{current_user_id: user_id} ->
            Challenges.self_complete_challenge(user_id, challenge_id)

          _ ->
            {:error, "authentication required"}
        end
      end)
    end

    # Skipped for spike: bulk operations, async bulk operations, push notifications
  end

  defp maybe_rename(map, old_key, new_key) do
    case Map.pop(map, old_key) do
      {nil, map} -> map
      {value, map} -> Map.put(map, new_key, value)
    end
  end
end
