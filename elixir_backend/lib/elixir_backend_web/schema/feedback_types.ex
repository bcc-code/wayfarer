defmodule ElixirBackendWeb.Schema.FeedbackTypes do
  use Absinthe.Schema.Notation

  object :user_feedback do
    field :id, non_null(:id)
    field :message, non_null(:string)
    field :can_contact_me, non_null(:boolean)
    field :user_agent, :string
    field :platform, :string
    field :screen_width, :integer
    field :screen_height, :integer
    field :app_version, :string
    field :locale, :string
    field :project_id, :id
    field :timezone, :string
    field :context_url, :string
    field :tags, non_null(list_of(non_null(:string)))
    field :handled_at, :datetime
    field :created_at, non_null(:datetime)
  end

  input_object :submit_feedback_input do
    field :message, non_null(:string)
    field :can_contact_me, :boolean
    field :project_id, :id
    field :tags, list_of(non_null(:string))
    field :user_agent, :string
    field :platform, :string
    field :screen_width, :integer
    field :screen_height, :integer
    field :app_version, :string
    field :locale, :string
    field :timezone, :string
    field :context_url, :string
  end

  input_object :feedback_filter do
    field :user_id, :id
    field :tags, list_of(non_null(:string))
    field :handled, :boolean
    field :platform, :string
  end

  object :feedback_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:user_feedback)
  end

  object :feedback_connection do
    field :edges, non_null(list_of(non_null(:feedback_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end
end
