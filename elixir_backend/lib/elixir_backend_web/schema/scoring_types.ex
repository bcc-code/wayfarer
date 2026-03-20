defmodule ElixirBackendWeb.Schema.ScoringTypes do
  use Absinthe.Schema.Notation
  @moduledoc false
  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  # ── Enums ──

  enum :score_source_type do
    value(:achievement, as: "ACHIEVEMENT")
    value(:challenge, as: "CHALLENGE")
    value(:event, as: "EVENT")
    value(:manual, as: "MANUAL")
    value(:bet, as: "BET")
  end

  enum :leaderboard_entity_type do
    value(:persons, as: "PERSONS")
    value(:teams, as: "TEAMS")
    value(:superteams, as: "SUPERTEAMS")
    value(:churches, as: "CHURCHES")
  end

  enum :team_score_distribution_mode do
    value(:split, as: "SPLIT")
    value(:each, as: "EACH")
  end

  # ── Score Journal ──

  object :score_journal do
    field :id, non_null(:id)
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)
    field :user, non_null(:user), resolve: dataloader(ElixirBackend.Repo)
    field :event, :event, resolve: dataloader(ElixirBackend.Repo)
    field :points, non_null(:integer)
    field :source_type, non_null(:score_source_type)
    field :reason, :string
    field :created_at, non_null(:datetime)
  end

  # ── Leaderboard ──

  object :leaderboard_entry do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :score, non_null(:integer)
    field :rank, :integer
    field :tags, non_null(list_of(non_null(:string)))
    field :image, :string
    field :last_score_at, :datetime
  end

  object :leaderboard_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:leaderboard_entry)
  end

  object :leaderboard_connection do
    field :edges, non_null(list_of(non_null(:leaderboard_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
    field :me, :leaderboard_entry
  end

  # ── Input Types ──

  input_object :create_score_adjustment_input do
    field :project_id, non_null(:id)
    field :user_id, non_null(:id)
    field :event_id, :id
    field :challenge_id, :id
    field :points, non_null(:integer)
    field :reason, :string
  end

  input_object :create_team_score_adjustment_input do
    field :team_id, non_null(:id)
    field :project_id, non_null(:id)
    field :event_id, :id
    field :points, non_null(:integer)
    field :distribution_mode, non_null(:team_score_distribution_mode)
    field :reason, :string
  end

  input_object :score_journal_filter do
    field :project_id, :id
    field :user_id, :id
    field :event_id, :id
    field :challenge_id, :id
    field :source_type, :score_source_type
    field :ids, list_of(non_null(:id))
  end

  input_object :leaderboard_filter do
    field :min_score, :integer
    field :max_score, :integer
    field :church_id, :id
    field :country, :string
    field :gender, :string
    field :team_id, :id
    field :super_team_id, :id
  end

  input_object :age_range_input do
    field :min, non_null(:integer)
    field :max, non_null(:integer)
  end

  # ── Pagination ──

  object :score_journal_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:score_journal)
  end

  object :score_journal_connection do
    field :edges, non_null(list_of(non_null(:score_journal_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end
end
