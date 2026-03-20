defmodule ElixirBackend.Webhooks.Webhook do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @valid_event_types ~w(CHALLENGE_COMPLETED ACHIEVEMENT_UNLOCKED SCORE_UPDATED USER_JOINED CUSTOM)

  schema "webhooks" do
    field :name, :string
    field :url, :string
    field :event_type, :string
    field :include_user, :boolean, default: false
    field :include_challenge, :boolean, default: false
    field :include_achievement, :boolean, default: false
    field :active, :boolean, default: true
    field :secret, :string

    belongs_to :project, ElixirBackend.Projects.Project, type: :string

    timestamps(type: :utc_datetime)
  end

  @required_fields [:id, :project_id, :name, :url, :event_type]
  @optional_fields [:include_user, :include_challenge, :include_achievement, :active, :secret]

  def changeset(webhook, attrs) do
    webhook
    |> cast(attrs, @required_fields ++ @optional_fields)
    |> validate_required(@required_fields)
    |> validate_inclusion(:event_type, @valid_event_types)
    |> foreign_key_constraint(:project_id)
  end

  def update_changeset(webhook, attrs) do
    webhook
    |> cast(attrs, @optional_fields ++ [:name, :url, :event_type])
    |> validate_inclusion(:event_type, @valid_event_types)
  end
end
