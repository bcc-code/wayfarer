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
  import_types(ElixirBackendWeb.Schema.RoleTypes)
  import_types(ElixirBackendWeb.Schema.UserTypes)
  import_types(ElixirBackendWeb.Schema.StreakTypes)
  import_types(ElixirBackendWeb.Schema.ExternalContentTypes)
  import_types(ElixirBackendWeb.Schema.AchievementTypes)
  import_types(ElixirBackendWeb.Schema.QuizTypes)
  import_types(ElixirBackendWeb.Schema.ScoringTypes)
  import_types(ElixirBackendWeb.Schema.BulkJobTypes)

  # Queries
  import_types(ElixirBackendWeb.Schema.ChurchQueries)
  import_types(ElixirBackendWeb.Schema.ProjectQueries)
  import_types(ElixirBackendWeb.Schema.EventQueries)
  import_types(ElixirBackendWeb.Schema.ChallengeQueries)
  import_types(ElixirBackendWeb.Schema.TeamQueries)
  import_types(ElixirBackendWeb.Schema.RoleQueries)
  import_types(ElixirBackendWeb.Schema.UserQueries)
  import_types(ElixirBackendWeb.Schema.StreakQueries)
  import_types(ElixirBackendWeb.Schema.ExternalContentQueries)
  import_types(ElixirBackendWeb.Schema.AchievementQueries)
  import_types(ElixirBackendWeb.Schema.QuizQueries)
  import_types(ElixirBackendWeb.Schema.ScoringQueries)
  import_types(ElixirBackendWeb.Schema.BulkJobQueries)

  # Mutations
  import_types(ElixirBackendWeb.Schema.ChurchMutations)
  import_types(ElixirBackendWeb.Schema.ProjectMutations)
  import_types(ElixirBackendWeb.Schema.EventMutations)
  import_types(ElixirBackendWeb.Schema.ChallengeMutations)
  import_types(ElixirBackendWeb.Schema.TeamMutations)
  import_types(ElixirBackendWeb.Schema.RoleMutations)
  import_types(ElixirBackendWeb.Schema.UserMutations)
  import_types(ElixirBackendWeb.Schema.StreakMutations)
  import_types(ElixirBackendWeb.Schema.AchievementMutations)
  import_types(ElixirBackendWeb.Schema.QuizMutations)
  import_types(ElixirBackendWeb.Schema.ScoringMutations)

  query do
    import_fields(:church_queries)
    import_fields(:project_queries)
    import_fields(:event_queries)
    import_fields(:challenge_queries)
    import_fields(:team_queries)
    import_fields(:role_queries)
    import_fields(:user_queries)
    import_fields(:streak_queries)
    import_fields(:external_content_queries)
    import_fields(:achievement_queries)
    import_fields(:quiz_queries)
    import_fields(:scoring_queries)
    import_fields(:bulk_job_queries)
  end

  mutation do
    import_fields(:church_mutations)
    import_fields(:project_mutations)
    import_fields(:event_mutations)
    import_fields(:challenge_mutations)
    import_fields(:team_mutations)
    import_fields(:role_mutations)
    import_fields(:user_mutations)
    import_fields(:streak_mutations)
    import_fields(:achievement_mutations)
    import_fields(:quiz_mutations)
    import_fields(:scoring_mutations)
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
