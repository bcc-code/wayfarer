defmodule ElixirBackend.Churches do
  @moduledoc """
  Context module for church business logic.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.Pagination
  alias ElixirBackend.Churches.Church

  # ── Read ──

  def get_church(id) do
    case Repo.get(Church, id) do
      nil -> {:error, :not_found}
      church -> {:ok, church}
    end
  end

  def get_church!(id), do: Repo.get!(Church, id)

  def list_churches(filter \\ %{}, pagination_opts \\ %{}) do
    base_query = from(c in Church)

    query = apply_filter(base_query, filter)
    total_count = Repo.aggregate(query, :count)

    pagination_opts = Map.put(pagination_opts, :sort_field, :created_at)

    items =
      query
      |> Pagination.paginate(pagination_opts)
      |> Repo.all()

    {:ok, Pagination.build_connection(items, pagination_opts, total_count)}
  end

  # ── Write ──

  def update_church(id, attrs) do
    with {:ok, church} <- get_church(id) do
      church
      |> Church.update_changeset(attrs)
      |> Repo.update()
    end
  end

  # ── Private helpers ──

  defp apply_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [c], c.id in ^ids)

      {:country, country}, q when is_binary(country) and country != "" ->
        where(q, [c], c.country == ^country)

      {:category, category}, q when is_binary(category) and category != "" ->
        where(q, [c], c.category == ^category)

      _, q ->
        q
    end)
  end

  defp apply_filter(query, _), do: query
end
