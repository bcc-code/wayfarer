defmodule ElixirBackendWeb.Schema.ScoringQueries do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Scoring

  object :scoring_queries do
    field :score_journal, non_null(:score_journal_connection) do
      arg(:project_id, non_null(:id))
      arg(:user_id, non_null(:id))
      arg(:filter, :score_journal_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole,
        roles: ["user", "admin", "superadmin"]
      )

      resolve(fn _, args, _ ->
        filter =
          (args[:filter] || %{})
          |> Map.put(:project_id, args.project_id)
          |> Map.put(:user_id, args.user_id)

        pagination = Map.take(args, [:first, :after, :last, :before])
        Scoring.list_entries(filter, pagination)
      end)
    end

    field :admin_score_journal, non_null(:score_journal_connection) do
      arg(:filter, :score_journal_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      middleware(ElixirBackendWeb.Schema.Middleware.RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _, args, _ ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])
        Scoring.list_entries(filter, pagination)
      end)
    end
  end
end
