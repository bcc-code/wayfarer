defmodule ElixirBackendWeb.Schema.TranslationMutations do
  use Absinthe.Schema.Notation
  @moduledoc false

  alias ElixirBackend.Translations
  alias ElixirBackendWeb.Schema.Middleware.RequireRole

  object :translation_mutations do
    # ── Project ──

    field :upsert_project_translation, non_null(:project) do
      arg(:input, non_null(:project_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:project, input) do
          ElixirBackend.Projects.get_project(input.project_id)
        end
      end)
    end

    field :delete_project_translation, non_null(:project) do
      arg(:project_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{project_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:project, id, lang)
        ElixirBackend.Projects.get_project(id)
      end)
    end

    # ── Event ──

    field :upsert_event_translation, non_null(:event) do
      arg(:input, non_null(:event_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:event, input) do
          ElixirBackend.Events.get_event(input.event_id)
        end
      end)
    end

    field :delete_event_translation, non_null(:event) do
      arg(:event_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{event_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:event, id, lang)
        ElixirBackend.Events.get_event(id)
      end)
    end

    # ── Streak ──

    field :upsert_streak_translation, non_null(:streak) do
      arg(:input, non_null(:streak_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:streak, input) do
          ElixirBackend.Streaks.get_streak(input.streak_id)
        end
      end)
    end

    field :delete_streak_translation, non_null(:streak) do
      arg(:streak_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{streak_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:streak, id, lang)
        ElixirBackend.Streaks.get_streak(id)
      end)
    end

    # ── Challenge ──

    field :upsert_challenge_translation, non_null(:challenge) do
      arg(:input, non_null(:challenge_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:challenge, input) do
          ElixirBackend.Challenges.get_challenge(input.challenge_id)
        end
      end)
    end

    field :delete_challenge_translation, non_null(:challenge) do
      arg(:challenge_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{challenge_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:challenge, id, lang)
        ElixirBackend.Challenges.get_challenge(id)
      end)
    end

    # ── Achievement ──

    field :upsert_achievement_translation, non_null(:achievement) do
      arg(:input, non_null(:achievement_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:achievement, input) do
          ElixirBackend.Achievements.get_achievement(input.achievement_id)
        end
      end)
    end

    field :delete_achievement_translation, non_null(:achievement) do
      arg(:achievement_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{achievement_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:achievement, id, lang)
        ElixirBackend.Achievements.get_achievement(id)
      end)
    end

    # ── Consent ──

    field :upsert_consent_translation, non_null(:consent) do
      arg(:input, non_null(:consent_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:consent, input) do
          ElixirBackend.Consents.get_consent(input.consent_id)
        end
      end)
    end

    field :delete_consent_translation, non_null(:consent) do
      arg(:consent_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{consent_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:consent, id, lang)
        ElixirBackend.Consents.get_consent(id)
      end)
    end

    # ── Quiz ──

    field :upsert_quiz_translation, non_null(:quiz) do
      arg(:input, non_null(:quiz_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:quiz, input) do
          ElixirBackend.Quizzes.get_quiz(input.quiz_id)
        end
      end)
    end

    field :delete_quiz_translation, non_null(:quiz) do
      arg(:quiz_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{quiz_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:quiz, id, lang)
        ElixirBackend.Quizzes.get_quiz(id)
      end)
    end

    # ── Quiz Question ──

    field :upsert_quiz_question_translation, non_null(:quiz_question) do
      arg(:input, non_null(:quiz_question_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:quiz_question, input) do
          ElixirBackend.Quizzes.get_question(input.question_id)
        end
      end)
    end

    field :delete_quiz_question_translation, non_null(:quiz_question) do
      arg(:question_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{question_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:quiz_question, id, lang)
        ElixirBackend.Quizzes.get_question(id)
      end)
    end

    # ── Quiz Predefined Answer ──

    field :upsert_quiz_answer_translation, non_null(:quiz_predefined_answer) do
      arg(:input, non_null(:quiz_predefined_answer_translation_input))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{input: input}, _ ->
        with {:ok, _translation} <- Translations.upsert_translation(:quiz_answer, input) do
          ElixirBackend.Quizzes.get_predefined_answer(input.answer_id)
        end
      end)
    end

    field :delete_quiz_answer_translation, non_null(:quiz_predefined_answer) do
      arg(:answer_id, non_null(:id))
      arg(:language_code, non_null(:string))
      middleware(RequireRole, roles: ["admin", "superadmin"])

      resolve(fn _parent, %{answer_id: id, language_code: lang}, _ ->
        Translations.delete_translation(:quiz_answer, id, lang)
        ElixirBackend.Quizzes.get_predefined_answer(id)
      end)
    end
  end
end
