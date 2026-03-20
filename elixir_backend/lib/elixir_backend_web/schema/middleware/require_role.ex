defmodule ElixirBackendWeb.Schema.Middleware.RequireRole do
  @moduledoc """
  Absinthe middleware that checks the current user has one of the required roles.
  Applied per-field on mutations.
  """
  @behaviour Absinthe.Middleware

  @impl true
  def call(%{context: %{roles: user_roles}} = resolution, roles: required_roles) do
    if Enum.any?(required_roles, &(&1 in user_roles)) do
      resolution
    else
      Absinthe.Resolution.put_result(resolution, {:error, "unauthorized: insufficient role"})
    end
  end

  def call(resolution, _config) do
    Absinthe.Resolution.put_result(resolution, {:error, "unauthorized: not authenticated"})
  end
end
