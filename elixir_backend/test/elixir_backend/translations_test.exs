defmodule ElixirBackend.TranslationsTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Translations

  import ElixirBackend.TestHelpers

  describe "get_translation/3" do
    test "returns translation for an entity" do
      project = create_project()

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "en",
          name: "English Name"
        })

      assert {:ok, translation} = Translations.get_translation(:project, project.id, "en")
      assert translation.name == "English Name"
      assert translation.language_code == "en"
    end

    test "returns error for missing translation" do
      project = create_project()
      assert {:error, :not_found} = Translations.get_translation(:project, project.id, "en")
    end
  end

  describe "upsert_translation/2" do
    test "inserts a new translation" do
      project = create_project()

      assert {:ok, translation} =
               Translations.upsert_translation(:project, %{
                 project_id: project.id,
                 language_code: "en",
                 name: "English Name",
                 description: "English Description"
               })

      assert translation.name == "English Name"
      assert translation.description == "English Description"
    end

    test "updates an existing translation" do
      project = create_project()

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "en",
          name: "Old Name"
        })

      {:ok, updated} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "en",
          name: "New Name"
        })

      assert updated.name == "New Name"
    end

    test "works for challenge translations with extra fields" do
      project = create_project()

      {:ok, challenge} =
        ElixirBackend.Challenges.create_challenge(%{
          name: "Test Challenge",
          challenge_type: "SIMPLE",
          project_id: project.id,
          button_text: "Fullfør"
        })

      assert {:ok, translation} =
               Translations.upsert_translation(:challenge, %{
                 challenge_id: challenge.id,
                 language_code: "en",
                 name: "English Challenge",
                 button_text: "Click me",
                 notification_text: "You did it"
               })

      assert translation.button_text == "Click me"
      assert translation.notification_text == "You did it"
    end

    test "works for achievement translations" do
      project = create_project()

      {:ok, achievement} =
        ElixirBackend.Achievements.create_simple_achievement(%{
          name: "Test Achievement",
          description_pending: "Do it",
          description_completed: "Done",
          image_pending: "pending.png",
          image_completed: "completed.png",
          points: 100,
          hidden: false,
          project_id: project.id
        })

      assert {:ok, translation} =
               Translations.upsert_translation(:achievement, %{
                 achievement_id: achievement.id,
                 language_code: "en",
                 name: "English Achievement",
                 description_pending: "Do it in English",
                 description_completed: "Done in English"
               })

      assert translation.description_pending == "Do it in English"
    end

    test "works for consent translations" do
      {:ok, consent} =
        ElixirBackend.Consents.create_consent(%{
          key: "privacy",
          title: "Personvern",
          short_text: "Kort tekst",
          body: "Full tekst"
        })

      assert {:ok, translation} =
               Translations.upsert_translation(:consent, %{
                 consent_id: consent.id,
                 language_code: "en",
                 title: "Privacy",
                 short_text: "Short text",
                 body: "Full text"
               })

      assert translation.title == "Privacy"
    end
  end

  describe "delete_translation/3" do
    test "deletes an existing translation" do
      project = create_project()

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "en",
          name: "English"
        })

      assert :ok = Translations.delete_translation(:project, project.id, "en")
      assert {:error, :not_found} = Translations.get_translation(:project, project.id, "en")
    end

    test "returns error for nonexistent translation" do
      project = create_project()

      assert {:error, :not_found} =
               Translations.delete_translation(:project, project.id, "en")
    end
  end

  describe "translation_status/2" do
    test "returns status for all language translations" do
      project = create_project()

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "en",
          name: "English Name"
        })

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "de",
          name: "German Name",
          description: "German Description"
        })

      status = Translations.translation_status(:project, project.id)
      assert length(status) == 2

      en_status = Enum.find(status, &(&1.language_code == "en"))
      assert en_status.fields == ["name"]

      de_status = Enum.find(status, &(&1.language_code == "de"))
      assert "name" in de_status.fields
      assert "description" in de_status.fields
    end

    test "returns empty list when no translations exist" do
      project = create_project()
      assert [] = Translations.translation_status(:project, project.id)
    end
  end

  describe "apply_translation/3" do
    test "overlays translated fields onto entity" do
      project = create_project(%{name: "Norwegian Name", description: "Norwegian Description"})

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "en",
          name: "English Name"
        })

      translated = Translations.apply_translation(project, :project, "en")
      assert translated.name == "English Name"
      # description was not translated, keeps original
      assert translated.description == "Norwegian Description"
    end

    test "skips translation for default language" do
      project = create_project(%{name: "Norwegian Name"})

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "no",
          name: "Should not be applied"
        })

      result = Translations.apply_translation(project, :project, "no")
      assert result.name == "Norwegian Name"
    end

    test "skips translation for nil language" do
      project = create_project(%{name: "Norwegian Name"})
      result = Translations.apply_translation(project, :project, nil)
      assert result.name == "Norwegian Name"
    end

    test "returns entity unchanged when no translation exists" do
      project = create_project(%{name: "Norwegian Name"})
      result = Translations.apply_translation(project, :project, "en")
      assert result.name == "Norwegian Name"
    end

    test "does not overlay empty string translations" do
      project = create_project(%{name: "Norwegian Name"})

      {:ok, _} =
        Translations.upsert_translation(:project, %{
          project_id: project.id,
          language_code: "en",
          name: ""
        })

      result = Translations.apply_translation(project, :project, "en")
      assert result.name == "Norwegian Name"
    end
  end

  describe "quiz question and answer translations" do
    test "works for quiz question translations" do
      project = create_project()
      quiz = create_quiz(project)

      {:ok, question} =
        ElixirBackend.Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Hva heter du?",
          question_order: 1
        })

      assert {:ok, translation} =
               Translations.upsert_translation(:quiz_question, %{
                 question_id: question.id,
                 language_code: "en",
                 question_text: "What is your name?"
               })

      assert translation.question_text == "What is your name?"
    end
  end
end
