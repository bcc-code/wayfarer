defmodule ElixirBackend.Achievements.Achievement do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @valid_types ~w(SIMPLE CONTENT STREAK QUIZ)

  schema "achievements" do
    field :achievement_type, :string
    field :name, :string
    field :description_pending, :string
    field :description_completed, :string
    field :notification_text, :string
    field :image_pending, :string
    field :image_completed, :string
    field :points, :integer, default: 0
    field :hidden, :boolean, default: false
    field :awardable_from, :utc_datetime
    field :sort_order, :integer, default: 0

    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    belongs_to :event, ElixirBackend.Events.Event, type: :string

    field :challenge_id, :string

    timestamps(type: :utc_datetime)
  end

  @required_fields [:id, :achievement_type, :name, :project_id]
  @optional_fields [
    :description_pending,
    :description_completed,
    :notification_text,
    :image_pending,
    :image_completed,
    :points,
    :hidden,
    :awardable_from,
    :sort_order,
    :event_id,
    :challenge_id
  ]

  def changeset(achievement, attrs) do
    achievement
    |> cast(attrs, @required_fields ++ @optional_fields)
    |> validate_required(@required_fields)
    |> validate_inclusion(:achievement_type, @valid_types)
    |> foreign_key_constraint(:project_id)
    |> foreign_key_constraint(:event_id)
  end

  def update_changeset(achievement, attrs) do
    achievement
    |> cast(attrs, @optional_fields ++ [:name])
  end
end
