defmodule ElixirBackendWeb.Schema do
  @moduledoc "Root Absinthe schema for the Elixir backend GraphQL API."
  use Absinthe.Schema

  import_types(ElixirBackendWeb.Schema.Scalars)

  # Domain types
  import_types(ElixirBackendWeb.Schema.ChurchTypes)
  import_types(ElixirBackendWeb.Schema.ProjectTypes)
  import_types(ElixirBackendWeb.Schema.EventTypes)
  import_types(ElixirBackendWeb.Schema.ChallengeTypes)
  import_types(ElixirBackendWeb.Schema.TeamTypes)
  import_types(ElixirBackendWeb.Schema.UserTypes)

  # Queries
  import_types(ElixirBackendWeb.Schema.ChurchQueries)
  import_types(ElixirBackendWeb.Schema.ProjectQueries)
  import_types(ElixirBackendWeb.Schema.EventQueries)
  import_types(ElixirBackendWeb.Schema.ChallengeQueries)
  import_types(ElixirBackendWeb.Schema.TeamQueries)
  import_types(ElixirBackendWeb.Schema.UserQueries)

  # Mutations
  import_types(ElixirBackendWeb.Schema.ChurchMutations)
  import_types(ElixirBackendWeb.Schema.ProjectMutations)
  import_types(ElixirBackendWeb.Schema.EventMutations)
  import_types(ElixirBackendWeb.Schema.ChallengeMutations)
  import_types(ElixirBackendWeb.Schema.TeamMutations)
  import_types(ElixirBackendWeb.Schema.UserMutations)

  query do
    import_fields(:church_queries)
    import_fields(:project_queries)
    import_fields(:event_queries)
    import_fields(:challenge_queries)
    import_fields(:team_queries)
    import_fields(:user_queries)
  end

  mutation do
    import_fields(:church_mutations)
    import_fields(:project_mutations)
    import_fields(:event_mutations)
    import_fields(:challenge_mutations)
    import_fields(:team_mutations)
    import_fields(:user_mutations)
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
