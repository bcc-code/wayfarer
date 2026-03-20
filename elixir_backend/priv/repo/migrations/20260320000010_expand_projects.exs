defmodule ElixirBackend.Repo.Migrations.ExpandProjects do
  use Ecto.Migration

  def change do
    alter table(:projects) do
      add :description, :text, null: false, default: ""
      add :rules, :text
      add :info_message, :text
      add :info_message_start, :utc_datetime
      add :info_message_end, :utc_datetime
      add :archived, :boolean, default: false

      # Branding
      add :logo_url, :string
      add :banner_url, :string
      add :rounding, :integer, null: false, default: 0

      # Light mode colors
      add :color_light_accent, :string, null: false, default: "#000000"
      add :color_light_accent_contrast, :string, null: false, default: "#FFFFFF"
      add :color_light_on_accent, :string, null: false, default: "#FFFFFF"
      add :color_light_background_default, :string, null: false, default: "#FFFFFF"
      add :color_light_background_raised, :string, null: false, default: "#F5F5F5"
      add :color_light_background_indent, :string, null: false, default: "#E0E0E0"
      add :color_light_text_default, :string, null: false, default: "#000000"
      add :color_light_text_muted, :string, null: false, default: "#666666"
      add :color_light_text_hint, :string, null: false, default: "#999999"
      add :color_light_shadow_default, :string, null: false, default: "#00000020"
      add :color_light_shadow_blank, :string, null: false, default: "#00000000"
      add :color_light_border_default, :string, null: false, default: "#E0E0E0"

      # Dark mode colors
      add :color_dark_accent, :string, null: false, default: "#FFFFFF"
      add :color_dark_accent_contrast, :string, null: false, default: "#000000"
      add :color_dark_on_accent, :string, null: false, default: "#000000"
      add :color_dark_background_default, :string, null: false, default: "#121212"
      add :color_dark_background_raised, :string, null: false, default: "#1E1E1E"
      add :color_dark_background_indent, :string, null: false, default: "#2C2C2C"
      add :color_dark_text_default, :string, null: false, default: "#FFFFFF"
      add :color_dark_text_muted, :string, null: false, default: "#AAAAAA"
      add :color_dark_text_hint, :string, null: false, default: "#777777"
      add :color_dark_shadow_default, :string, null: false, default: "#00000040"
      add :color_dark_shadow_blank, :string, null: false, default: "#00000000"
      add :color_dark_border_default, :string, null: false, default: "#333333"
    end
  end
end
