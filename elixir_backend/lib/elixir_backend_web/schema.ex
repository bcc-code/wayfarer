defmodule ElixirBackendWeb.Schema do
  @moduledoc "Root Absinthe schema for the Elixir backend GraphQL API."
  use Absinthe.Schema

  import_types(ElixirBackendWeb.Schema.Scalars)
  import_types(ElixirBackendWeb.Schema.ChallengeTypes)
  import_types(ElixirBackendWeb.Schema.ChallengeQueries)
  import_types(ElixirBackendWeb.Schema.ChallengeMutations)

  query do
    import_fields(:challenge_queries)
  end

  mutation do
    import_fields(:challenge_mutations)
  end

  # Dataloader plugin for batch-loading associations
  def context(ctx) do
    loader =
      Dataloader.new()
      |> Dataloader.add_source(ElixirBackend.Repo, Dataloader.Ecto.new(ElixirBackend.Repo))

    Map.put(ctx, :loader, loader)
  end

  def plugins do
    [Absinthe.Middleware.Dataloader | Absinthe.Plugin.defaults()]
  end
end
