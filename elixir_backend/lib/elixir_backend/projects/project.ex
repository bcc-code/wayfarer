defmodule ElixirBackend.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "projects" do
    field :name, :string
    field :description, :string, default: ""
    field :start_date, :utc_datetime
    field :end_date, :utc_datetime
    field :rules, :string
    field :info_message, :string
    field :info_message_start, :utc_datetime
    field :info_message_end, :utc_datetime
    field :archived, :boolean, default: false

    # Branding
    field :logo_url, :string
    field :banner_url, :string
    field :rounding, :integer, default: 0

    # Light mode colors
    field :color_light_accent, :string, default: "#000000"
    field :color_light_accent_contrast, :string, default: "#FFFFFF"
    field :color_light_on_accent, :string, default: "#FFFFFF"
    field :color_light_background_default, :string, default: "#FFFFFF"
    field :color_light_background_raised, :string, default: "#F5F5F5"
    field :color_light_background_indent, :string, default: "#E0E0E0"
    field :color_light_text_default, :string, default: "#000000"
    field :color_light_text_muted, :string, default: "#666666"
    field :color_light_text_hint, :string, default: "#999999"
    field :color_light_shadow_default, :string, default: "#00000020"
    field :color_light_shadow_blank, :string, default: "#00000000"
    field :color_light_border_default, :string, default: "#E0E0E0"

    # Dark mode colors
    field :color_dark_accent, :string, default: "#FFFFFF"
    field :color_dark_accent_contrast, :string, default: "#000000"
    field :color_dark_on_accent, :string, default: "#000000"
    field :color_dark_background_default, :string, default: "#121212"
    field :color_dark_background_raised, :string, default: "#1E1E1E"
    field :color_dark_background_indent, :string, default: "#2C2C2C"
    field :color_dark_text_default, :string, default: "#FFFFFF"
    field :color_dark_text_muted, :string, default: "#AAAAAA"
    field :color_dark_text_hint, :string, default: "#777777"
    field :color_dark_shadow_default, :string, default: "#00000040"
    field :color_dark_shadow_blank, :string, default: "#00000000"
    field :color_dark_border_default, :string, default: "#333333"

    has_many :events, ElixirBackend.Events.Event

    timestamps(type: :utc_datetime)
  end

  @required_fields [:id, :name, :start_date, :end_date]
  @optional_fields [
    :description,
    :rules,
    :info_message,
    :info_message_start,
    :info_message_end,
    :archived,
    :logo_url,
    :banner_url,
    :rounding,
    :color_light_accent,
    :color_light_accent_contrast,
    :color_light_on_accent,
    :color_light_background_default,
    :color_light_background_raised,
    :color_light_background_indent,
    :color_light_text_default,
    :color_light_text_muted,
    :color_light_text_hint,
    :color_light_shadow_default,
    :color_light_shadow_blank,
    :color_light_border_default,
    :color_dark_accent,
    :color_dark_accent_contrast,
    :color_dark_on_accent,
    :color_dark_background_default,
    :color_dark_background_raised,
    :color_dark_background_indent,
    :color_dark_text_default,
    :color_dark_text_muted,
    :color_dark_text_hint,
    :color_dark_shadow_default,
    :color_dark_shadow_blank,
    :color_dark_border_default
  ]

  def changeset(project, attrs) do
    project
    |> cast(attrs, @required_fields ++ @optional_fields)
    |> validate_required(@required_fields)
  end

  def create_changeset(project, attrs) do
    changeset(project, attrs)
  end

  def update_changeset(project, attrs) do
    project
    |> cast(attrs, @optional_fields ++ [:name, :start_date, :end_date])
  end
end
