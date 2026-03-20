defmodule ElixirBackendWeb.Schema.TranslationTypes do
  use Absinthe.Schema.Notation
  @moduledoc false

  object :translation_field_status do
    field :language_code, non_null(:string)
    field :fields, non_null(list_of(non_null(:string)))
  end

  # ── Translation Input Types ──

  input_object :project_translation_input do
    field :project_id, non_null(:id)
    field :language_code, non_null(:string)
    field :name, :string
    field :description, :string
    field :rules, :string
  end

  input_object :event_translation_input do
    field :event_id, non_null(:id)
    field :language_code, non_null(:string)
    field :name, :string
    field :description, :string
  end

  input_object :streak_translation_input do
    field :streak_id, non_null(:id)
    field :language_code, non_null(:string)
    field :name, :string
    field :description, :string
  end

  input_object :challenge_translation_input do
    field :challenge_id, non_null(:id)
    field :language_code, non_null(:string)
    field :name, :string
    field :description, :string
    field :button_text, :string
    field :notification_text, :string
  end

  input_object :achievement_translation_input do
    field :achievement_id, non_null(:id)
    field :language_code, non_null(:string)
    field :name, :string
    field :description_pending, :string
    field :description_completed, :string
    field :notification_text, :string
  end

  input_object :consent_translation_input do
    field :consent_id, non_null(:id)
    field :language_code, non_null(:string)
    field :title, :string
    field :short_text, :string
    field :body, :string
  end

  input_object :quiz_translation_input do
    field :quiz_id, non_null(:id)
    field :language_code, non_null(:string)
    field :name, :string
    field :description, :string
  end

  input_object :quiz_question_translation_input do
    field :question_id, non_null(:id)
    field :language_code, non_null(:string)
    field :question_text, :string
  end

  input_object :quiz_predefined_answer_translation_input do
    field :answer_id, non_null(:id)
    field :language_code, non_null(:string)
    field :answer_text, :string
  end
end
