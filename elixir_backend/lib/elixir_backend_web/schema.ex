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
  import_types(ElixirBackendWeb.Schema.PushNotificationTypes)
  import_types(ElixirBackendWeb.Schema.WebhookTypes)
  import_types(ElixirBackendWeb.Schema.ConsentTypes)
  import_types(ElixirBackendWeb.Schema.FeedbackTypes)
  import_types(ElixirBackendWeb.Schema.FileUploadTypes)
  import_types(ElixirBackendWeb.Schema.AdminTypes)
  import_types(ElixirBackendWeb.Schema.TranslationTypes)

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
  import_types(ElixirBackendWeb.Schema.PushNotificationQueries)
  import_types(ElixirBackendWeb.Schema.WebhookQueries)
  import_types(ElixirBackendWeb.Schema.ConsentQueries)
  import_types(ElixirBackendWeb.Schema.FeedbackQueries)
  import_types(ElixirBackendWeb.Schema.FileUploadQueries)
  import_types(ElixirBackendWeb.Schema.SettingsQueries)
  import_types(ElixirBackendWeb.Schema.AdminQueries)

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
  import_types(ElixirBackendWeb.Schema.PushNotificationMutations)
  import_types(ElixirBackendWeb.Schema.WebhookMutations)
  import_types(ElixirBackendWeb.Schema.ConsentMutations)
  import_types(ElixirBackendWeb.Schema.FeedbackMutations)
  import_types(ElixirBackendWeb.Schema.AdminMutations)
  import_types(ElixirBackendWeb.Schema.TranslationMutations)

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
    import_fields(:push_notification_queries)
    import_fields(:webhook_queries)
    import_fields(:consent_queries)
    import_fields(:feedback_queries)
    import_fields(:file_upload_queries)
    import_fields(:settings_queries)
    import_fields(:admin_queries)
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
    import_fields(:push_notification_mutations)
    import_fields(:webhook_mutations)
    import_fields(:consent_mutations)
    import_fields(:feedback_mutations)
    import_fields(:admin_mutations)
    import_fields(:translation_mutations)
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
