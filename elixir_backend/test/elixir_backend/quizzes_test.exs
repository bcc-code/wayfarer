defmodule ElixirBackend.QuizzesTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.Quizzes
  import ElixirBackend.TestHelpers

  setup do
    project = create_project()
    user = create_user()
    quiz = create_quiz(project)
    %{project: project, user: user, quiz: quiz}
  end

  # ── Quiz CRUD ──

  describe "create_quiz/1" do
    test "creates a quiz with valid attrs", %{project: project} do
      {:ok, quiz} =
        Quizzes.create_quiz(%{
          name: "New Quiz",
          description: "Desc",
          project_id: project.id,
          completion_points: 5
        })

      assert quiz.name == "New Quiz"
      assert quiz.completion_points == 5
      assert String.starts_with?(quiz.id, "QZ")
    end
  end

  describe "get_quiz/1" do
    test "returns quiz by id", %{quiz: quiz} do
      assert {:ok, found} = Quizzes.get_quiz(quiz.id)
      assert found.id == quiz.id
    end

    test "returns error for missing quiz" do
      assert {:error, :not_found} = Quizzes.get_quiz("QZ00000000000000000000000000")
    end
  end

  describe "update_quiz/2" do
    test "updates quiz fields", %{quiz: quiz} do
      {:ok, updated} = Quizzes.update_quiz(quiz.id, %{name: "Updated"})
      assert updated.name == "Updated"
    end
  end

  describe "delete_quiz/1" do
    test "deletes a quiz", %{quiz: quiz} do
      assert {:ok, _} = Quizzes.delete_quiz(quiz.id)
      assert {:error, :not_found} = Quizzes.get_quiz(quiz.id)
    end
  end

  describe "list_quizzes/2" do
    test "returns paginated quizzes", %{project: project} do
      {:ok, conn} = Quizzes.list_quizzes(%{project_id: project.id}, %{first: 10})
      assert conn.total_count >= 1
    end

    test "filters by project_id", %{project: project} do
      other_project = create_project(%{name: "Other"})
      create_quiz(other_project, %{name: "Other Quiz"})

      {:ok, conn} = Quizzes.list_quizzes(%{project_id: project.id}, %{first: 10})
      assert Enum.all?(conn.edges, fn e -> e.node.project_id == project.id end)
    end
  end

  # ── Questions ──

  describe "add_question/2" do
    test "adds a free text question", %{quiz: quiz} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "What is 1+1?",
          question_order: 1
        })

      assert q.question_type == "FREE_TEXT"
      assert String.starts_with?(q.id, "QQ")
    end

    test "adds a predefined question with answers", %{quiz: quiz} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "PREDEFINED",
          question_text: "Pick one",
          question_order: 1,
          predefined_answers: [
            %{answer_text: "A", is_correct: true, answer_order: 1},
            %{answer_text: "B", is_correct: false, answer_order: 2}
          ]
        })

      answers = Quizzes.get_predefined_answers(q.id)
      assert length(answers) == 2
      assert Enum.any?(answers, fn a -> a.is_correct end)
    end

    test "adds an ordering question with items", %{quiz: quiz} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "ORDERING",
          question_text: "Order these",
          question_order: 1,
          ordering_items: [
            %{item_text: "First", correct_order: 1},
            %{item_text: "Second", correct_order: 2}
          ]
        })

      items = Quizzes.get_ordering_items(q.id)
      assert length(items) == 2
    end
  end

  describe "update_question/2" do
    test "updates question text", %{quiz: quiz} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Old",
          question_order: 1
        })

      {:ok, updated} = Quizzes.update_question(q.id, %{question_text: "New"})
      assert updated.question_text == "New"
    end

    test "replaces predefined answers on update", %{quiz: quiz} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "PREDEFINED",
          question_text: "Pick",
          question_order: 1,
          predefined_answers: [%{answer_text: "Old", is_correct: true, answer_order: 1}]
        })

      {:ok, _} =
        Quizzes.update_question(q.id, %{
          predefined_answers: [
            %{answer_text: "New1", is_correct: false, answer_order: 1},
            %{answer_text: "New2", is_correct: true, answer_order: 2}
          ]
        })

      answers = Quizzes.get_predefined_answers(q.id)
      assert length(answers) == 2
      assert Enum.any?(answers, fn a -> a.answer_text == "New1" end)
    end
  end

  describe "delete_question/1" do
    test "deletes a question", %{quiz: quiz} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Delete me",
          question_order: 1
        })

      assert {:ok, _} = Quizzes.delete_question(q.id)
      assert {:error, :not_found} = Quizzes.get_question(q.id)
    end
  end

  describe "reorder_questions/2" do
    test "reorders questions", %{quiz: quiz} do
      {:ok, q1} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q1",
          question_order: 1
        })

      {:ok, q2} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q2",
          question_order: 2
        })

      {:ok, reordered} = Quizzes.reorder_questions(quiz.id, [q2.id, q1.id])
      assert Enum.at(reordered, 0).id == q2.id
      assert Enum.at(reordered, 1).id == q1.id
    end
  end

  # ── Submissions ──

  describe "create_submission/3" do
    test "creates a submission with question order", %{quiz: quiz, user: user} do
      {:ok, q1} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q1",
          question_order: 1
        })

      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)
      assert sub.quiz_id == quiz.id
      assert sub.user_id == user.id
      assert sub.question_order == [q1.id]
      assert sub.started_at != nil
      assert sub.completed_at == nil
    end

    test "randomizes question order when enabled", %{project: project, user: user} do
      quiz = create_quiz(project, %{randomize_questions: true})

      # Add several questions to increase probability of different order
      for i <- 1..10 do
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q#{i}",
          question_order: i
        })
      end

      questions = Quizzes.get_questions(quiz.id)
      original_order = Enum.map(questions, & &1.id)

      # Create multiple submissions and check at least one differs
      orders =
        for _ <- 1..5 do
          {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)
          sub.question_order
        end

      # With 10 items and 5 tries, extremely unlikely all match the original
      assert Enum.any?(orders, fn order -> order != original_order end)
    end

    test "sets expires_at when quiz has timeout", %{project: project, user: user} do
      quiz = create_quiz(project, %{timeout_seconds: 300})

      {:ok, _q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q",
          question_order: 1
        })

      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)
      assert sub.expires_at != nil
    end
  end

  describe "submit_answer/2" do
    test "submits a text answer", %{quiz: quiz, user: user} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q",
          question_order: 1
        })

      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)

      {:ok, resp} =
        Quizzes.submit_answer(sub.id, %{
          question_id: q.id,
          text_response: "my answer"
        })

      assert resp.text_response == "my answer"
      assert resp.submission_id == sub.id
    end

    test "upserts answer on same question", %{quiz: quiz, user: user} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "FREE_TEXT",
          question_text: "Q",
          question_order: 1
        })

      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)

      {:ok, _} = Quizzes.submit_answer(sub.id, %{question_id: q.id, text_response: "first"})
      {:ok, _} = Quizzes.submit_answer(sub.id, %{question_id: q.id, text_response: "second"})

      responses = Quizzes.get_responses(sub.id)
      assert length(responses) == 1
      assert hd(responses).text_response == "second"
    end
  end

  describe "finalize_submission/1" do
    test "scores and finalizes submission", %{quiz: quiz, user: user} do
      {:ok, q} =
        Quizzes.add_question(quiz.id, %{
          question_type: "PREDEFINED",
          question_text: "Pick",
          question_order: 1,
          predefined_answers: [
            %{answer_text: "Correct", is_correct: true, answer_order: 1},
            %{answer_text: "Wrong", is_correct: false, answer_order: 2}
          ]
        })

      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)
      answers = Quizzes.get_predefined_answers(q.id)
      correct = Enum.find(answers, & &1.is_correct)

      {:ok, _} =
        Quizzes.submit_answer(sub.id, %{
          question_id: q.id,
          selected_answer_ids: [correct.id],
          is_correct: true
        })

      {:ok, finalized} = Quizzes.finalize_submission(sub.id)
      assert finalized.completed_at != nil
      assert finalized.score == 1
      assert finalized.max_score == 1
      assert finalized.points_awarded == quiz.completion_points
    end
  end

  describe "get_user_submissions/2" do
    test "returns user submissions for a quiz", %{quiz: quiz, user: user} do
      {:ok, _} = Quizzes.create_submission(quiz.id, user.id)
      subs = Quizzes.get_user_submissions(quiz.id, user.id)
      assert length(subs) == 1
    end
  end

  describe "get_user_active_submission/2" do
    test "returns active (non-completed) submission", %{quiz: quiz, user: user} do
      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)
      active = Quizzes.get_user_active_submission(quiz.id, user.id)
      assert active.id == sub.id
    end

    test "returns nil after finalization", %{quiz: quiz, user: user} do
      {:ok, sub} = Quizzes.create_submission(quiz.id, user.id)
      {:ok, _} = Quizzes.finalize_submission(sub.id)
      assert Quizzes.get_user_active_submission(quiz.id, user.id) == nil
    end
  end

  # ── Sessions ──

  describe "create_session/2" do
    test "creates a session in DRAFT state", %{quiz: quiz} do
      {:ok, session} =
        Quizzes.create_session(quiz.id, %{name: "Session 1", created_by: "admin-user"})

      assert session.state == "DRAFT"
      assert session.name == "Session 1"
      assert String.starts_with?(session.id, "QN")
    end
  end

  describe "transition_session_state/2" do
    test "transitions DRAFT -> OPEN -> LOCKED -> FINISHED", %{quiz: quiz} do
      {:ok, session} =
        Quizzes.create_session(quiz.id, %{name: "S", created_by: "admin"})

      {:ok, s2} = Quizzes.transition_session_state(session.id, "OPEN")
      assert s2.state == "OPEN"

      {:ok, s3} = Quizzes.transition_session_state(s2.id, "LOCKED")
      assert s3.state == "LOCKED"

      {:ok, s4} = Quizzes.transition_session_state(s3.id, "FINISHED")
      assert s4.state == "FINISHED"
    end

    test "allows LOCKED -> OPEN transition", %{quiz: quiz} do
      {:ok, session} =
        Quizzes.create_session(quiz.id, %{name: "S", created_by: "admin"})

      {:ok, s2} = Quizzes.transition_session_state(session.id, "OPEN")
      {:ok, s3} = Quizzes.transition_session_state(s2.id, "LOCKED")
      {:ok, s4} = Quizzes.transition_session_state(s3.id, "OPEN")
      assert s4.state == "OPEN"
    end

    test "rejects invalid transitions", %{quiz: quiz} do
      {:ok, session} =
        Quizzes.create_session(quiz.id, %{name: "S", created_by: "admin"})

      assert {:error, _} = Quizzes.transition_session_state(session.id, "FINISHED")
    end
  end

  describe "delete_session/1" do
    test "deletes session in DRAFT state", %{quiz: quiz} do
      {:ok, session} =
        Quizzes.create_session(quiz.id, %{name: "S", created_by: "admin"})

      assert {:ok, _} = Quizzes.delete_session(session.id)
    end

    test "rejects deleting non-DRAFT session", %{quiz: quiz} do
      {:ok, session} =
        Quizzes.create_session(quiz.id, %{name: "S", created_by: "admin"})

      {:ok, _} = Quizzes.transition_session_state(session.id, "OPEN")
      assert {:error, _} = Quizzes.delete_session(session.id)
    end
  end
end
