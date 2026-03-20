defmodule ElixirBackendWeb.Schema.ProjectTypes do
  @moduledoc "Absinthe types for projects: object, branding, inputs, pagination."
  use Absinthe.Schema.Notation

  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  # ── Branding types ──

  object :color_set do
    field :accent, non_null(:string)
    field :accent_contrast, non_null(:string)
    field :on_accent, non_null(:string)
    field :background_default, non_null(:string)
    field :background_raised, non_null(:string)
    field :background_indent, non_null(:string)
    field :text_default, non_null(:string)
    field :text_muted, non_null(:string)
    field :text_hint, non_null(:string)
    field :shadow_default, non_null(:string)
    field :shadow_blank, non_null(:string)
    field :border_default, non_null(:string)
  end

  object :colors do
    field :dark, non_null(:color_set)
    field :light, non_null(:color_set)
  end

  object :branding do
    field :logo, :string
    field :banner, :string
    field :colors, non_null(:colors)
    field :rounding, non_null(:integer)
  end

  # ── Project object ──

  object :project do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :start_date, non_null(:datetime)
    field :end_date, non_null(:datetime)

    field :archived_at, :boolean do
      resolve(fn project, _, _ -> {:ok, project.archived} end)
    end

    field :rules, :string
    field :info_message, :string
    field :info_message_start, :datetime
    field :info_message_end, :datetime

    field :events, non_null(list_of(non_null(:event))), resolve: dataloader(ElixirBackend.Repo)

    field :leaderboard, non_null(:leaderboard_connection) do
      arg(:entity_type, non_null(:leaderboard_entity_type))
      arg(:filter, :leaderboard_filter)
      arg(:first, :integer)
      arg(:after, :string)
      arg(:last, :integer)
      arg(:before, :string)

      resolve(fn project, args, _ ->
        filter = args[:filter] || %{}
        pagination = Map.take(args, [:first, :after, :last, :before])

        ElixirBackend.Scoring.get_project_leaderboard(
          project.id,
          args.entity_type,
          filter,
          pagination
        )
      end)
    end

    field :branding, non_null(:branding) do
      resolve(fn project, _, _ ->
        {:ok,
         %{
           logo: project.logo_url,
           banner: project.banner_url,
           rounding: project.rounding,
           colors: %{
             light: %{
               accent: project.color_light_accent,
               accent_contrast: project.color_light_accent_contrast,
               on_accent: project.color_light_on_accent,
               background_default: project.color_light_background_default,
               background_raised: project.color_light_background_raised,
               background_indent: project.color_light_background_indent,
               text_default: project.color_light_text_default,
               text_muted: project.color_light_text_muted,
               text_hint: project.color_light_text_hint,
               shadow_default: project.color_light_shadow_default,
               shadow_blank: project.color_light_shadow_blank,
               border_default: project.color_light_border_default
             },
             dark: %{
               accent: project.color_dark_accent,
               accent_contrast: project.color_dark_accent_contrast,
               on_accent: project.color_dark_on_accent,
               background_default: project.color_dark_background_default,
               background_raised: project.color_dark_background_raised,
               background_indent: project.color_dark_background_indent,
               text_default: project.color_dark_text_default,
               text_muted: project.color_dark_text_muted,
               text_hint: project.color_dark_text_hint,
               shadow_default: project.color_dark_shadow_default,
               shadow_blank: project.color_dark_shadow_blank,
               border_default: project.color_dark_border_default
             }
           }
         }}
      end)
    end
  end

  # ── Pagination ──

  object :project_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:project)
  end

  object :project_connection do
    field :edges, non_null(list_of(non_null(:project_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  # ── Input types ──

  input_object :color_set_input do
    field :accent, non_null(:string)
    field :accent_contrast, non_null(:string)
    field :on_accent, non_null(:string)
    field :background_default, non_null(:string)
    field :background_raised, non_null(:string)
    field :background_indent, non_null(:string)
    field :text_default, non_null(:string)
    field :text_muted, non_null(:string)
    field :text_hint, non_null(:string)
    field :shadow_default, non_null(:string)
    field :shadow_blank, non_null(:string)
    field :border_default, non_null(:string)
  end

  input_object :colors_input do
    field :dark, non_null(:color_set_input)
    field :light, non_null(:color_set_input)
  end

  input_object :branding_input do
    field :logo, :string
    field :banner, :string
    field :colors, non_null(:colors_input)
    field :rounding, non_null(:integer)
  end

  input_object :create_project_input do
    field :name, non_null(:string)
    field :description, :string
    field :rules, :string
    field :info_message, :string
    field :info_message_start, :datetime
    field :info_message_end, :datetime
    field :start_date, non_null(:datetime)
    field :end_date, non_null(:datetime)
    field :branding, non_null(:branding_input)
  end

  input_object :update_project_input do
    field :name, :string
    field :description, :string
    field :rules, :string
    field :info_message, :string
    field :info_message_start, :datetime
    field :info_message_end, :datetime
    field :start_date, :datetime
    field :end_date, :datetime
    field :branding, :branding_input
  end

  input_object :project_filter do
    field :ids, list_of(non_null(:id))
    field :archived, :boolean
    field :start_date_after, :datetime
    field :start_date_before, :datetime
    field :end_date_after, :datetime
    field :end_date_before, :datetime
  end
end
