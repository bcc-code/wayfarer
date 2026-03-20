defmodule ElixirBackendWeb.Schema.ScoringMutations do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Scoring

  object :scoring_mutations do
    field :create_score_adjustment, non_null(:score_journal) do
      arg(:input, non_null(:create_score_adjustment_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["m2m", "admin", "superadmin"]
      )

      resolve(fn _, %{input: input}, %{context: context} ->
        attrs =
          input
          |> Map.put(:source_type, "MANUAL")
          |> Map.put(:awarded_by, context[:current_user_id])

        with {:ok, entry} <- Scoring.create_entry(attrs) do
          Scoring.update_leaderboards(entry)
          {:ok, entry}
        end
      end)
    end

    field :create_team_score_adjustment, non_null(list_of(non_null(:score_journal))) do
      arg(:input, non_null(:create_team_score_adjustment_input))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["m2m", "admin", "superadmin"]
      )

      resolve(fn _, %{input: input}, _ ->
        Scoring.create_team_adjustment(input)
      end)
    end

    field :delete_score_journal_entry, non_null(:boolean) do
      arg(:id, non_null(:id))

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, %{id: id}, _ ->
        case Scoring.delete_entry(id) do
          {:ok, _} -> {:ok, true}
          {:error, reason} -> {:error, reason}
        end
      end)
    end
  end
end
