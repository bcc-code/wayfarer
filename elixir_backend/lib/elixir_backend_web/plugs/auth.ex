defmodule ElixirBackendWeb.Plugs.Auth do
  @moduledoc """
  Plug that parses JWT from Authorization header and extracts user claims.
  Uses the same HS256 secret as the Go backend.
  """
  import Plug.Conn
  @behaviour Plug

  @impl true
  def init(opts), do: opts

  @impl true
  def call(conn, _opts) do
    with ["Bearer " <> token] <- get_req_header(conn, "authorization"),
         {:ok, claims} <- verify_token(token) do
      conn
      |> assign(:current_user_id, claims["sub"] || claims["user_id"])
      |> assign(:roles, claims["roles"] || [])
    else
      _ -> conn
    end
  end

  defp verify_token(token) do
    secret = Application.get_env(:elixir_backend, :jwt_secret, "dev-secret")
    signer = Joken.Signer.create("HS256", secret)

    case Joken.verify(token, signer) do
      {:ok, claims} -> {:ok, claims}
      {:error, reason} -> {:error, reason}
    end
  end
end
