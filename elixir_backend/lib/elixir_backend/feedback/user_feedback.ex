defmodule ElixirBackend.Feedback.UserFeedback do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "user_feedback" do
    field :message, :string
    field :can_contact_me, :boolean, default: false
    field :user_agent, :string
    field :platform, :string
    field :screen_width, :integer
    field :screen_height, :integer
    field :app_version, :string
    field :locale, :string
    field :project_id, :string
    field :timezone, :string
    field :context_url, :string
    field :tags, {:array, :string}, default: []
    field :handled_at, :utc_datetime
    field :created_at, :utc_datetime

    belongs_to :user, ElixirBackend.Accounts.User, type: :string
  end

  def changeset(fb, attrs) do
    fb
    |> cast(attrs, [
      :id,
      :user_id,
      :message,
      :can_contact_me,
      :user_agent,
      :platform,
      :screen_width,
      :screen_height,
      :app_version,
      :locale,
      :project_id,
      :timezone,
      :context_url,
      :tags,
      :handled_at,
      :created_at
    ])
    |> validate_required([:id, :message])
    |> foreign_key_constraint(:user_id)
  end
end
