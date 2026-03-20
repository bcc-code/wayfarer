defmodule ElixirBackendWeb.Schema.StreakMutations do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Streaks

  object :streak_mutations do
    field :create_streak, non_null(:streak) do
      arg(:input, non_null(:create_streak_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{input: input}, _ ->
        attrs = %{
          name: input.name,
          description: input.description,
          project_id: input.project_id,
          relevant_days: input.relevant_days
        }

        Streaks.create_streak(attrs)
      end)
    end

    field :update_streak, non_null(:streak) do
      arg(:id, non_null(:id))
      arg(:input, non_null(:update_streak_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id, input: input}, _ ->
        attrs =
          input
          |> Enum.reject(fn {_k, v} -> is_nil(v) end)
          |> Map.new()

        Streaks.update_streak(id, attrs)
      end)
    end

    field :delete_streak, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id}, _ ->
        case Streaks.delete_streak(id) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end
  end
end
