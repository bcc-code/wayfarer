defmodule ElixirBackend.Quizzes do
  @moduledoc """
  Context module for quiz management, questions, submissions, and sessions.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Cache
  alias ElixirBackend.Pagination

  alias ElixirBackend.Quizzes.{
    Quiz,
    QuizQuestion,
    QuizPredefinedAnswer,
    QuizOrderingItem,
    QuizSubmission,
    QuizResponse,
    QuizSession
  }

  # ── Quiz Read ──

  def get_quiz(id) do
    Cache.fetch(Cache.quiz_key(id), fn ->
      case Repo.get(Quiz, id) do
        nil -> {:error, :not_found}
        quiz -> {:ok, quiz}
      end
    end)
  end

  def get_quiz!(id), do: Repo.get!(Quiz, id)

  def list_quizzes(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(q in Quiz)

    query = apply_quiz_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  def get_questions(quiz_id) do
    Cache.fetch_raw(Cache.quiz_questions_key(quiz_id), fn ->
      from(q in QuizQuestion,
        where: q.quiz_id == ^quiz_id,
        order_by: [asc: q.question_order]
      )
      |> Repo.all()
    end)
  end

  def get_question(id) do
    case Repo.get(QuizQuestion, id) do
      nil -> {:error, :not_found}
      q -> {:ok, q}
    end
  end

  def get_predefined_answer(id) do
    case Repo.get(QuizPredefinedAnswer, id) do
      nil -> {:error, :not_found}
      a -> {:ok, a}
    end
  end

  def get_predefined_answers(question_id) do
    Cache.fetch_raw(Cache.quiz_answers_key(question_id), fn ->
      from(a in QuizPredefinedAnswer,
        where: a.question_id == ^question_id,
        order_by: [asc: a.answer_order]
      )
      |> Repo.all()
    end)
  end

  def get_ordering_items(question_id) do
    from(oi in QuizOrderingItem,
      where: oi.question_id == ^question_id,
      order_by: [asc: oi.correct_order]
    )
    |> Repo.all()
  end

  # ── Quiz Write ──

  def create_quiz(attrs) do
    id = ULID.new_quiz_id()

    %Quiz{}
    |> Quiz.changeset(Map.put(attrs, :id, id))
    |> Repo.insert()
  end

  def update_quiz(id, attrs) do
    with {:ok, quiz} <- get_quiz(id) do
      result =
        quiz
        |> Quiz.update_changeset(attrs)
        |> Repo.update()

      with {:ok, updated} <- result do
        Cache.invalidate_quiz(id)
        {:ok, updated}
      end
    end
  end

  def delete_quiz(id) do
    with {:ok, quiz} <- get_quiz(id) do
      result = Repo.delete(quiz)

      with {:ok, deleted} <- result do
        Cache.invalidate_quiz(id)
        {:ok, deleted}
      end
    end
  end

  # ── Question Write ──

  def add_question(quiz_id, attrs) do
    id = ULID.new_quiz_question_id()
    predefined_answers = attrs[:predefined_answers] || []
    ordering_items = attrs[:ordering_items] || []

    result =
      Repo.transaction(fn ->
        question =
          %QuizQuestion{}
          |> QuizQuestion.changeset(
            attrs
            |> Map.put(:id, id)
            |> Map.put(:quiz_id, quiz_id)
            |> Map.drop([:predefined_answers, :ordering_items])
          )
          |> Repo.insert!()

        insert_predefined_answers(question.id, predefined_answers)
        insert_ordering_items(question.id, ordering_items)

        question
      end)

    with {:ok, _} <- result, do: Cache.del(Cache.quiz_questions_key(quiz_id))
    result
  end

  def update_question(id, attrs) do
    with {:ok, question} <- get_question(id) do
      result = do_update_question(id, question, attrs)

      with {:ok, _} <- result do
        Cache.del(Cache.quiz_questions_key(question.quiz_id))
        Cache.invalidate_quiz_answers(id)
      end

      result
    end
  end

  defp do_update_question(id, question, attrs) do
    predefined_answers = attrs[:predefined_answers]
    ordering_items = attrs[:ordering_items]

    Repo.transaction(fn ->
      updated =
        question
        |> QuizQuestion.update_changeset(Map.drop(attrs, [:predefined_answers, :ordering_items]))
        |> Repo.update!()

      if predefined_answers do
        from(a in QuizPredefinedAnswer, where: a.question_id == ^id) |> Repo.delete_all()
        insert_predefined_answers(id, predefined_answers)
      end

      if ordering_items do
        from(oi in QuizOrderingItem, where: oi.question_id == ^id) |> Repo.delete_all()
        insert_ordering_items(id, ordering_items)
      end

      updated
    end)
  end

  def delete_question(id) do
    with {:ok, question} <- get_question(id) do
      result = Repo.delete(question)

      with {:ok, _} <- result do
        Cache.del(Cache.quiz_questions_key(question.quiz_id))
        Cache.invalidate_quiz_answers(id)
      end

      result
    end
  end

  def reorder_questions(quiz_id, question_ids) do
    result =
      Repo.transaction(fn ->
        # First pass: set to negative temp values to avoid unique constraint violations
        question_ids
        |> Enum.with_index(1)
        |> Enum.each(fn {qid, order} ->
          from(q in QuizQuestion, where: q.id == ^qid and q.quiz_id == ^quiz_id)
          |> Repo.update_all(set: [question_order: -order])
        end)

        # Second pass: set to final positive values
        question_ids
        |> Enum.with_index(1)
        |> Enum.each(fn {qid, order} ->
          from(q in QuizQuestion, where: q.id == ^qid and q.quiz_id == ^quiz_id)
          |> Repo.update_all(set: [question_order: order])
        end)

        # Query DB directly inside transaction to see uncommitted changes
        from(q in QuizQuestion,
          where: q.quiz_id == ^quiz_id,
          order_by: [asc: q.question_order]
        )
        |> Repo.all()
      end)

    with {:ok, _} <- result, do: Cache.del(Cache.quiz_questions_key(quiz_id))
    result
  end

  # ── Submission ──

  def get_submission(id) do
    case Repo.get(QuizSubmission, id) do
      nil -> {:error, :not_found}
      sub -> {:ok, sub}
    end
  end

  def list_submissions(quiz_id, opts \\ %{}) do
    base_query = from(s in QuizSubmission, where: s.quiz_id == ^quiz_id)

    query =
      case opts[:user_id] do
        nil -> base_query
        uid -> where(base_query, [s], s.user_id == ^uid)
      end

    total_count = Repo.aggregate(query, :count)

    limit = opts[:first] || opts[:last] || 25

    items =
      query
      |> order_by([s], desc: s.started_at)
      |> limit(^limit)
      |> Repo.all()

    edges = Enum.map(items, fn item -> %{cursor: item.id, node: item} end)

    {:ok,
     %{
       edges: edges,
       page_info: %{has_next_page: length(items) == limit, has_previous_page: false},
       total_count: total_count
     }}
  end

  def create_submission(quiz_id, user_id, opts \\ %{}) do
    with {:ok, quiz} <- get_quiz(quiz_id) do
      id = ULID.new_quiz_submission_id()
      questions = get_questions(quiz_id)
      now = DateTime.utc_now() |> DateTime.truncate(:second)

      question_ids = Enum.map(questions, & &1.id)

      question_order =
        if quiz.randomize_questions do
          Enum.shuffle(question_ids)
        else
          question_ids
        end

      expires_at =
        if quiz.timeout_seconds do
          DateTime.add(now, quiz.timeout_seconds)
        end

      %QuizSubmission{}
      |> QuizSubmission.changeset(%{
        id: id,
        quiz_id: quiz_id,
        user_id: user_id,
        session_id: opts[:session_id],
        started_at: now,
        expires_at: expires_at,
        question_order: question_order
      })
      |> Repo.insert()
    end
  end

  def submit_answer(submission_id, attrs) do
    id = ULID.new_quiz_response_id()
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    %QuizResponse{}
    |> QuizResponse.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:submission_id, submission_id)
      |> Map.put(:answered_at, now)
    )
    |> Repo.insert(
      on_conflict:
        {:replace,
         [
           :selected_answer_ids,
           :text_response,
           :number_response,
           :json_response,
           :submitted_order,
           :answered_at,
           :time_spent_seconds,
           :bet_amount
         ]},
      conflict_target: [:submission_id, :question_id]
    )
  end

  def finalize_submission(submission_id) do
    with {:ok, submission} <- get_submission(submission_id) do
      now = DateTime.utc_now() |> DateTime.truncate(:second)

      # Score the submission
      {score, max_score} = calculate_score(submission_id)

      # Get completion points from quiz
      quiz = get_quiz!(submission.quiz_id)

      submission
      |> Ecto.Changeset.change(
        completed_at: now,
        score: score,
        max_score: max_score,
        points_awarded: quiz.completion_points
      )
      |> Repo.update()
    end
  end

  def get_responses(submission_id) do
    from(r in QuizResponse,
      where: r.submission_id == ^submission_id,
      order_by: [asc: r.answered_at]
    )
    |> Repo.all()
  end

  def get_user_submissions(quiz_id, user_id) do
    from(s in QuizSubmission,
      where: s.quiz_id == ^quiz_id and s.user_id == ^user_id,
      order_by: [desc: s.started_at]
    )
    |> Repo.all()
  end

  def get_user_active_submission(quiz_id, user_id) do
    from(s in QuizSubmission,
      where:
        s.quiz_id == ^quiz_id and s.user_id == ^user_id and
          is_nil(s.completed_at),
      order_by: [desc: s.started_at],
      limit: 1
    )
    |> Repo.one()
  end

  # ── Sessions ──

  def get_session(id) do
    case Repo.get(QuizSession, id) do
      nil -> {:error, :not_found}
      session -> {:ok, session}
    end
  end

  def create_session(quiz_id, attrs) do
    id = ULID.new_quiz_session_id()

    %QuizSession{}
    |> QuizSession.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:quiz_id, quiz_id)
    )
    |> Repo.insert()
  end

  def update_session(id, attrs) do
    with {:ok, session} <- get_session(id) do
      session
      |> QuizSession.update_changeset(attrs)
      |> Repo.update()
    end
  end

  def delete_session(id) do
    with {:ok, session} <- get_session(id) do
      if session.state == "DRAFT" do
        Repo.delete(session)
      else
        {:error, "can only delete sessions in DRAFT state"}
      end
    end
  end

  def transition_session_state(id, new_state) do
    with {:ok, session} <- get_session(id) do
      if valid_transition?(session.state, new_state) do
        session
        |> Ecto.Changeset.change(state: new_state)
        |> Repo.update()
      else
        {:error, "invalid state transition from #{session.state} to #{new_state}"}
      end
    end
  end

  @doc """
  Transitions all quiz sessions whose scheduled time has passed.
  Handles DRAFT->OPEN, OPEN->LOCKED, and LOCKED->FINISHED in bulk.
  """
  def transition_due_sessions do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    transitions = [
      {"DRAFT", "OPEN", :open_at},
      {"OPEN", "LOCKED", :lock_at},
      {"LOCKED", "FINISHED", :finish_at}
    ]

    Enum.each(transitions, fn {from_state, to_state, time_field} ->
      from(s in QuizSession,
        where: s.state == ^from_state and field(s, ^time_field) <= ^now
      )
      |> Repo.update_all(set: [state: to_state, updated_at: now])
    end)
  end

  def get_quiz_sessions(quiz_id, state \\ nil) do
    query = from(s in QuizSession, where: s.quiz_id == ^quiz_id)

    query =
      if state do
        where(query, [s], s.state == ^state)
      else
        query
      end

    Repo.all(query)
  end

  # ── Private ──

  defp insert_predefined_answers(question_id, answers) do
    Enum.each(answers, fn answer ->
      %QuizPredefinedAnswer{}
      |> QuizPredefinedAnswer.changeset(
        answer
        |> Map.put(:id, ULID.new_quiz_answer_id())
        |> Map.put(:question_id, question_id)
      )
      |> Repo.insert!()
    end)
  end

  defp insert_ordering_items(question_id, items) do
    Enum.each(items, fn item ->
      %QuizOrderingItem{}
      |> QuizOrderingItem.changeset(
        item
        |> Map.put(:id, ULID.new_quiz_ordering_item_id())
        |> Map.put(:question_id, question_id)
      )
      |> Repo.insert!()
    end)
  end

  defp calculate_score(submission_id) do
    responses = get_responses(submission_id)

    score =
      Enum.count(responses, fn r ->
        r.is_correct == true
      end)

    gradable =
      Enum.count(responses, fn r ->
        r.is_correct != nil
      end)

    {score, gradable}
  end

  defp valid_transition?(from, to) do
    case {from, to} do
      {"DRAFT", "OPEN"} -> true
      {"OPEN", "LOCKED"} -> true
      {"LOCKED", "FINISHED"} -> true
      {"LOCKED", "OPEN"} -> true
      _ -> false
    end
  end

  defp apply_quiz_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, pid}, q when is_binary(pid) ->
        where(q, [qz], qz.project_id == ^pid)

      {:challenge_id, cid}, q when is_binary(cid) ->
        where(q, [qz], qz.challenge_id == ^cid)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [qz], qz.id in ^ids)

      _, q ->
        q
    end)
  end

  defp apply_quiz_filter(query, _), do: query
end
