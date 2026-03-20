defmodule ElixirBackendWeb.Router do
  use ElixirBackendWeb, :router

  pipeline :api do
    plug :accepts, ["json"]
    plug ElixirBackendWeb.Plugs.Auth
    plug ElixirBackendWeb.Plugs.AbsintheContext
  end

  scope "/api" do
    pipe_through :api

    forward "/graphql", Absinthe.Plug, schema: ElixirBackendWeb.Schema
  end

  # GraphiQL in development
  if Application.compile_env(:elixir_backend, :dev_routes) do
    import Phoenix.LiveDashboard.Router

    scope "/api" do
      pipe_through :api

      forward "/graphiql", Absinthe.Plug.GraphiQL,
        schema: ElixirBackendWeb.Schema,
        interface: :playground
    end

    scope "/dev" do
      pipe_through [:fetch_session, :protect_from_forgery]

      live_dashboard "/dashboard", metrics: ElixirBackendWeb.Telemetry
      forward "/mailbox", Plug.Swoosh.MailboxPreview
    end
  end
end
