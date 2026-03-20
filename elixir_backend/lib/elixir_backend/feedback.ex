defmodule ElixirBackend.Feedback do
  @moduledoc "Context for user feedback management."

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Feedback.UserFeedback

  def get_feedback(id) do
    case Repo.get(UserFeedback, id) do
      nil -> {:error, :not_found}
      fb -> {:ok, fb}
    end
  end

  def submit_feedback(attrs) do
    id = ULID.new_feedback_id()
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    %UserFeedback{}
    |> UserFeedback.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:created_at, now)
    )
    |> Repo.insert()
  end

  def list_feedback(filter \\ %{}, pagination_opts \\ %{}) do
    query = from(f in UserFeedback)
    query = apply_filter(query, filter)
    total_count = Repo.aggregate(query, :count)

    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    items =
      query
      |> order_by([f], desc: f.created_at)
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

  def delete_feedback(id) do
    with {:ok, fb} <- get_feedback(id) do
      Repo.delete(fb)
    end
  end

  def mark_handled(id) do
    with {:ok, fb} <- get_feedback(id) do
      now = DateTime.utc_now() |> DateTime.truncate(:second)

      fb
      |> Ecto.Changeset.change(handled_at: now)
      |> Repo.update()
    end
  end

  def update_tags(id, tags) do
    with {:ok, fb} <- get_feedback(id) do
      fb
      |> Ecto.Changeset.change(tags: tags)
      |> Repo.update()
    end
  end

  def get_tags do
    from(f in UserFeedback,
      select: f.tags,
      where: not is_nil(f.tags)
    )
    |> Repo.all()
    |> List.flatten()
    |> Enum.uniq()
    |> Enum.sort()
  end

  def get_platforms do
    from(f in UserFeedback,
      select: f.platform,
      where: not is_nil(f.platform),
      distinct: true
    )
    |> Repo.all()
    |> Enum.sort()
  end

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:user_id, uid}, q when is_binary(uid) ->
        where(q, [f], f.user_id == ^uid)

      {:platform, p}, q when is_binary(p) ->
        where(q, [f], f.platform == ^p)

      {:handled, true}, q ->
        where(q, [f], not is_nil(f.handled_at))

      {:handled, false}, q ->
        where(q, [f], is_nil(f.handled_at))

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
