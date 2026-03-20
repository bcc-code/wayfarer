defmodule ElixirBackendWeb.Plugs.AbsintheContext do
  @moduledoc """
  Copies auth info from conn.assigns into Absinthe context.
  """
  @behaviour Plug

  @impl true
  def init(opts), do: opts

  @impl true
  def call(conn, _opts) do
    language = parse_accept_language(conn)

    context =
      %{language: language}
      |> maybe_put(:current_user_id, conn.assigns[:current_user_id])
      |> maybe_put(:roles, conn.assigns[:roles])

    Absinthe.Plug.put_options(conn, context: context)
  end

  defp parse_accept_language(conn) do
    case Plug.Conn.get_req_header(conn, "accept-language") do
      [header | _] when is_binary(header) and header != "" ->
        header
        |> String.split(",", parts: 2)
        |> List.first()
        |> String.split(";", parts: 2)
        |> List.first()
        |> String.split("-", parts: 2)
        |> List.first()
        |> String.trim()
        |> String.downcase()

      _ ->
        "no"
    end
  end

  defp maybe_put(map, _key, nil), do: map
  defp maybe_put(map, key, value), do: Map.put(map, key, value)
end
