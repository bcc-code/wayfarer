defmodule ElixirBackend.Challenges.Challenge do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  @challenge_types ~w(SIMPLE QUIZ EXTERNAL PLUGIN)

  schema "challenges" do
    field :challenge_type, :string, default: "SIMPLE"
    field :name, :string
    field :description, :string, default: ""
    field :image_url, :string
    field :url, :string
    field :button_text, :string
    field :notification_text, :string, default: ""
    field :published_at, :utc_datetime
    field :visible_at, :utc_datetime
    field :started_at, :utc_datetime
    field :end_time, :utc_datetime
    field :allow_self_completion, :boolean, default: true
    field :requires_team_membership, :boolean, default: false
    field :requires_super_team_membership, :boolean, default: false
    field :plugin_challenge_id, :string
    field :plugin_data, :map

    belongs_to :project, ElixirBackend.Projects.Project, type: :string
    belongs_to :event, ElixirBackend.Events.Event, type: :string

    # Virtual fields for per-user resolution
    field :user_completed_at, :utc_datetime, virtual: true
    field :user_enrolled_at, :utc_datetime, virtual: true

    timestamps(type: :utc_datetime)
  end

  def create_changeset(challenge, attrs) do
    challenge
    |> cast(attrs, [
      :id,
      :project_id,
      :event_id,
      :challenge_type,
      :name,
      :description,
      :image_url,
      :url,
      :button_text,
      :notification_text,
      :published_at,
      :visible_at,
      :started_at,
      :end_time,
      :allow_self_completion,
      :requires_team_membership,
      :requires_super_team_membership,
      :plugin_challenge_id,
      :plugin_data
    ])
    |> validate_required([:id, :project_id, :challenge_type, :name])
    |> validate_inclusion(:challenge_type, @challenge_types)
    |> validate_type_specific_create()
    |> foreign_key_constraint(:project_id)
    |> foreign_key_constraint(:event_id)
  end

  def update_changeset(challenge, attrs) do
    challenge
    |> cast(attrs, [
      :name,
      :description,
      :image_url,
      :url,
      :button_text,
      :notification_text,
      :published_at,
      :visible_at,
      :started_at,
      :end_time,
      :event_id,
      :allow_self_completion,
      :requires_team_membership,
      :requires_super_team_membership,
      :plugin_challenge_id
    ])
    |> validate_type_specific_update(challenge.challenge_type)
    |> foreign_key_constraint(:event_id)
  end

  defp validate_type_specific_create(changeset) do
    case get_field(changeset, :challenge_type) do
      "SIMPLE" ->
        changeset
        |> validate_required([:button_text])
        |> validate_no_fields([:url, :plugin_challenge_id])

      "QUIZ" ->
        changeset
        |> validate_required([:button_text])
        |> validate_no_fields([:url, :plugin_challenge_id, :allow_self_completion])

      "EXTERNAL" ->
        changeset
        |> validate_required([:button_text, :url])
        |> validate_no_fields([:plugin_challenge_id, :allow_self_completion])

      "PLUGIN" ->
        changeset
        |> validate_required([:plugin_challenge_id])
        |> validate_no_fields([:url, :allow_self_completion])

      _ ->
        changeset
    end
  end

  defp validate_type_specific_update(changeset, challenge_type) do
    case challenge_type do
      "SIMPLE" ->
        validate_no_fields(changeset, [:url, :plugin_challenge_id])

      "QUIZ" ->
        validate_no_fields(changeset, [:url, :plugin_challenge_id, :allow_self_completion])

      "EXTERNAL" ->
        validate_no_fields(changeset, [:plugin_challenge_id, :allow_self_completion])

      "PLUGIN" ->
        validate_no_fields(changeset, [:url, :allow_self_completion])

      _ ->
        changeset
    end
  end

  defp validate_no_fields(changeset, fields) do
    params = changeset.params || %{}

    Enum.reduce(fields, changeset, fn field, cs ->
      str_key = Atom.to_string(field)

      if Map.has_key?(params, field) || Map.has_key?(params, str_key) do
        add_error(cs, field, "is not allowed for this challenge type")
      else
        cs
      end
    end)
  end
end
