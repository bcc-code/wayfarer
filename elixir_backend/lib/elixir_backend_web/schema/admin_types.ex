defmodule ElixirBackendWeb.Schema.AdminTypes do
  use Absinthe.Schema.Notation
  @moduledoc false

  object :admin_dashboard_stats do
    field :total_users, non_null(:integer)
    field :total_projects, non_null(:integer)
    field :total_challenges, non_null(:integer)
    field :total_points_awarded, non_null(:integer)
    field :new_users_last_7_days, non_null(:integer)
    field :active_projects_count, non_null(:integer)
  end

  object :age_group_stats do
    field :age_group, non_null(:string)
    field :user_count, non_null(:integer)
    field :average_score, non_null(:float)
  end

  object :user_score do
    field :user_id, non_null(:id)
    field :name, non_null(:string)
    field :total_score, non_null(:integer)
  end

  object :church_admin_statistics do
    field :church_id, non_null(:id)
    field :church_name, non_null(:string)
    field :project_id, non_null(:id)
    field :project_name, non_null(:string)
    field :age_groups, non_null(list_of(non_null(:age_group_stats)))
    field :total_users_in_teams, non_null(:integer)
    field :user_scores, non_null(list_of(non_null(:user_score)))
    field :last_updated_at, non_null(:datetime)
  end
end
