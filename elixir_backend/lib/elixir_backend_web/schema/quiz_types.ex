defmodule ElixirBackendWeb.Schema.QuizTypes do
  use Absinthe.Schema.Notation
  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  alias ElixirBackend.Quizzes
  alias ElixirBackend.Translations

  # ── Enums ──

  enum :quiz_question_type do
    value(:predefined, as: "PREDEFINED")
    value(:free_text, as: "FREE_TEXT")
    value(:number, as: "NUMBER")
    value(:json, as: "JSON")
    value(:ordering, as: "ORDERING")
  end

  enum :quiz_session_state do
    value(:draft, as: "DRAFT")
    value(:open, as: "OPEN")
    value(:locked, as: "LOCKED")
    value(:finished, as: "FINISHED")
  end

  # ── Quiz ──

  # Replace the placeholder :quiz from achievement_types
  # Note: the placeholder is already defined there, we reuse it by NOT redefining here

  # Since :quiz is already defined as a placeholder in achievement_types, we extend it there.
  # We need to remove the placeholder and define the full type here instead.

  object :quiz do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :image, :string, resolve: fn q, _, _ -> {:ok, q.image_url} end
    field :project, non_null(:project), resolve: dataloader(ElixirBackend.Repo)
    field :timeout_seconds, :integer
    field :randomize_questions, non_null(:boolean)
    field :reveal_correct_answers, non_null(:boolean)
    field :allow_retakes, non_null(:boolean)
    field :completion_points, non_null(:integer)
    field :end_time, :datetime

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn quiz, _, _ -> {:ok, Translations.translation_status(:quiz, quiz.id)} end)
    end

    field :questions, non_null(list_of(non_null(:quiz_question))) do
      resolve(fn quiz, _, _ ->
        {:ok, Quizzes.get_questions(quiz.id)}
      end)
    end

    field :user_submissions, non_null(list_of(non_null(:quiz_submission))) do
      resolve(fn quiz, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, []}
          user_id -> {:ok, Quizzes.get_user_submissions(quiz.id, user_id)}
        end
      end)
    end

    field :user_active_submission, :quiz_submission do
      resolve(fn quiz, _, %{context: context} ->
        case context[:current_user_id] do
          nil -> {:ok, nil}
          user_id -> {:ok, Quizzes.get_user_active_submission(quiz.id, user_id)}
        end
      end)
    end

    field :sessions, non_null(list_of(non_null(:quiz_session))) do
      arg(:state, :quiz_session_state)

      resolve(fn quiz, args, _ ->
        {:ok, Quizzes.get_quiz_sessions(quiz.id, args[:state])}
      end)
    end
  end

  # ── Quiz Question (interface) ──

  interface :quiz_question do
    field :id, non_null(:id)
    field :question_text, non_null(:string)
    field :question_order, non_null(:integer)
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, non_null(:boolean)
    field :translation_status, non_null(list_of(non_null(:translation_field_status)))

    resolve_type(fn
      %{question_type: "PREDEFINED"}, _ -> :predefined_question
      %{question_type: "FREE_TEXT"}, _ -> :free_text_question
      %{question_type: "NUMBER"}, _ -> :number_question
      %{question_type: "JSON"}, _ -> :json_question
      %{question_type: "ORDERING"}, _ -> :ordering_question
      _, _ -> :free_text_question
    end)
  end

  object :predefined_question do
    interface(:quiz_question)
    field :id, non_null(:id)
    field :question_text, non_null(:string)
    field :question_order, non_null(:integer)
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, non_null(:boolean)
    field :allow_multiple_selection, non_null(:boolean)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn q, _, _ -> {:ok, Translations.translation_status(:quiz_question, q.id)} end)
    end

    field :predefined_answers, non_null(list_of(non_null(:quiz_predefined_answer))) do
      resolve(fn q, _, _ -> {:ok, Quizzes.get_predefined_answers(q.id)} end)
    end
  end

  object :free_text_question do
    interface(:quiz_question)
    field :id, non_null(:id)
    field :question_text, non_null(:string)
    field :question_order, non_null(:integer)
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, non_null(:boolean)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn q, _, _ -> {:ok, Translations.translation_status(:quiz_question, q.id)} end)
    end
  end

  object :number_question do
    interface(:quiz_question)
    field :id, non_null(:id)
    field :question_text, non_null(:string)
    field :question_order, non_null(:integer)
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, non_null(:boolean)
    field :min_value, :float
    field :max_value, :float
    field :step_value, :float

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn q, _, _ -> {:ok, Translations.translation_status(:quiz_question, q.id)} end)
    end
  end

  object :json_question do
    interface(:quiz_question)
    field :id, non_null(:id)
    field :question_text, non_null(:string)
    field :question_order, non_null(:integer)
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, non_null(:boolean)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn q, _, _ -> {:ok, Translations.translation_status(:quiz_question, q.id)} end)
    end
  end

  object :ordering_question do
    interface(:quiz_question)
    field :id, non_null(:id)
    field :question_text, non_null(:string)
    field :question_order, non_null(:integer)
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, non_null(:boolean)

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn q, _, _ -> {:ok, Translations.translation_status(:quiz_question, q.id)} end)
    end

    field :ordering_items, non_null(list_of(non_null(:quiz_ordering_item))) do
      resolve(fn q, _, _ -> {:ok, Quizzes.get_ordering_items(q.id)} end)
    end
  end

  object :quiz_predefined_answer do
    field :id, non_null(:id)
    field :answer_text, non_null(:string)
    field :answer_order, non_null(:integer)
    field :is_correct, :boolean

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn a, _, _ -> {:ok, Translations.translation_status(:quiz_answer, a.id)} end)
    end
  end

  object :quiz_ordering_item do
    field :id, non_null(:id)
    field :item_text, non_null(:string)
    field :correct_order, :integer
  end

  # ── Submission ──

  object :quiz_submission do
    field :id, non_null(:id)
    field :started_at, non_null(:datetime)
    field :completed_at, :datetime
    field :expires_at, :datetime
    field :auto_submitted, non_null(:boolean)
    field :question_order, non_null(list_of(non_null(:id)))
    field :score, :integer
    field :max_score, :integer
    field :points_awarded, :integer

    field :responses, non_null(list_of(non_null(:quiz_response))) do
      resolve(fn sub, _, _ -> {:ok, Quizzes.get_responses(sub.id)} end)
    end

    field :score_percentage, :float do
      resolve(fn sub, _, _ ->
        case {sub.score, sub.max_score} do
          {s, m} when is_integer(s) and is_integer(m) and m > 0 ->
            {:ok, s / m * 100.0}

          _ ->
            {:ok, nil}
        end
      end)
    end
  end

  # ── Response (interface) ──

  interface :quiz_response do
    field :id, non_null(:id)
    field :answered_at, :datetime
    field :time_spent_seconds, :integer
    field :points_earned, :integer
    field :bet_amount, :integer

    resolve_type(fn response, _ ->
      # Determine type based on which response field is populated
      cond do
        response.selected_answer_ids != nil -> :predefined_response
        response.text_response != nil -> :free_text_response
        response.number_response != nil -> :number_response
        response.json_response != nil -> :json_response
        response.submitted_order != nil -> :ordering_response
        true -> :free_text_response
      end
    end)
  end

  object :predefined_response do
    interface(:quiz_response)
    field :id, non_null(:id)
    field :answered_at, :datetime
    field :time_spent_seconds, :integer
    field :points_earned, :integer
    field :bet_amount, :integer
    field :selected_answer_ids, non_null(list_of(non_null(:id)))
    field :is_correct, :boolean
  end

  object :free_text_response do
    interface(:quiz_response)
    field :id, non_null(:id)
    field :answered_at, :datetime
    field :time_spent_seconds, :integer
    field :points_earned, :integer
    field :bet_amount, :integer
    field :text_response, non_null(:string)
  end

  object :number_response do
    interface(:quiz_response)
    field :id, non_null(:id)
    field :answered_at, :datetime
    field :time_spent_seconds, :integer
    field :points_earned, :integer
    field :bet_amount, :integer
    field :number_response, non_null(:float)
  end

  object :json_response do
    interface(:quiz_response)
    field :id, non_null(:id)
    field :answered_at, :datetime
    field :time_spent_seconds, :integer
    field :points_earned, :integer
    field :bet_amount, :integer
    field :json_response, non_null(:json)
  end

  object :ordering_response do
    interface(:quiz_response)
    field :id, non_null(:id)
    field :answered_at, :datetime
    field :time_spent_seconds, :integer
    field :points_earned, :integer
    field :bet_amount, :integer
    field :submitted_order, non_null(list_of(non_null(:id)))
    field :is_correct, :boolean
  end

  # ── Session ──

  object :quiz_session do
    field :id, non_null(:id)
    field :name, :string
    field :state, non_null(:string)
    field :open_at, :datetime
    field :lock_at, :datetime
    field :finish_at, :datetime
    field :created_by, non_null(:string)
  end

  # ── Input Types ──

  input_object :create_quiz_input do
    field :name, non_null(:string)
    field :description, non_null(:string)
    field :image, :string
    field :project_id, non_null(:id)
    field :challenge_id, non_null(:id)
    field :timeout_seconds, :integer
    field :randomize_questions, non_null(:boolean)
    field :reveal_correct_answers, non_null(:boolean)
    field :allow_retakes, non_null(:boolean)
    field :completion_points, non_null(:integer)
    field :end_time, :datetime
  end

  input_object :update_quiz_input do
    field :name, :string
    field :description, :string
    field :image, :string
    field :timeout_seconds, :integer
    field :randomize_questions, :boolean
    field :reveal_correct_answers, :boolean
    field :allow_retakes, :boolean
    field :completion_points, :integer
    field :end_time, :datetime
  end

  input_object :create_quiz_question_input do
    field :question_type, non_null(:quiz_question_type)
    field :question_text, non_null(:string)
    field :question_order, non_null(:integer)
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, :boolean
    field :allow_multiple_selection, :boolean
    field :predefined_answers, list_of(non_null(:create_predefined_answer_input))
    field :min_value, :float
    field :max_value, :float
    field :step_value, :float
    field :ordering_items, list_of(non_null(:create_ordering_item_input))
  end

  input_object :create_predefined_answer_input do
    field :answer_text, non_null(:string)
    field :is_correct, non_null(:boolean)
    field :answer_order, non_null(:integer)
  end

  input_object :create_ordering_item_input do
    field :item_text, non_null(:string)
    field :correct_order, non_null(:integer)
  end

  input_object :update_quiz_question_input do
    field :question_text, :string
    field :question_order, :integer
    field :timeout_seconds, :integer
    field :points, :integer
    field :betting_enabled, :boolean
    field :allow_multiple_selection, :boolean
    field :predefined_answers, list_of(non_null(:create_predefined_answer_input))
    field :min_value, :float
    field :max_value, :float
    field :step_value, :float
    field :ordering_items, list_of(non_null(:create_ordering_item_input))
  end

  input_object :submit_quiz_answer_input do
    field :question_id, non_null(:id)
    field :selected_answer_ids, list_of(non_null(:id))
    field :text_response, :string
    field :number_response, :float
    field :json_response, :json
    field :submitted_order, list_of(non_null(:id))
    field :time_spent_seconds, :integer
    field :bet_amount, :integer
  end

  input_object :quiz_filter do
    field :project_id, :id
    field :challenge_id, :id
    field :ids, list_of(non_null(:id))
  end

  # ── Pagination ──

  object :quiz_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:quiz)
  end

  object :quiz_connection do
    field :edges, non_null(list_of(non_null(:quiz_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  object :quiz_submission_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:quiz_submission)
  end

  object :quiz_submission_connection do
    field :edges, non_null(list_of(non_null(:quiz_submission_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end
end
