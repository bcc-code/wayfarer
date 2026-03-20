# This file is responsible for configuring your application
# and its dependencies with the aid of the Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.

# General application configuration
import Config

config :elixir_backend,
  ecto_repos: [ElixirBackend.Repo],
  generators: [timestamp_type: :utc_datetime],
  jwt_secret: "dev-secret"

# Configures the endpoint
config :elixir_backend, ElixirBackendWeb.Endpoint,
  url: [host: "localhost"],
  adapter: Bandit.PhoenixAdapter,
  render_errors: [
    formats: [json: ElixirBackendWeb.ErrorJSON],
    layout: false
  ],
  pubsub_server: ElixirBackend.PubSub,
  live_view: [signing_salt: "9Zk2EgmG"]

# Configures the mailer
#
# By default it uses the "Local" adapter which stores the emails
# locally. You can see the emails in your browser, at "/dev/mailbox".
#
# For production it's recommended to configure a different adapter
# at the `config/runtime.exs`.
config :elixir_backend, ElixirBackend.Mailer, adapter: Swoosh.Adapters.Local

# Configures Oban background jobs
config :elixir_backend, Oban,
  repo: ElixirBackend.Repo,
  queues: [default: 10, bulk: 5, maintenance: 2],
  plugins: [
    Oban.Plugins.Pruner,
    {Oban.Plugins.Cron,
     crontab: [
       # Quiz session state transitions — every minute
       {"* * * * *", ElixirBackend.Workers.QuizSessionTransitions},
       # Cleanup old bulk jobs — daily at 03:00
       {"0 3 * * *", ElixirBackend.Workers.BulkJobCleanup},
       # Sync incomplete user data — daily at 04:00
       {"0 4 * * *", ElixirBackend.Workers.UserDataSync}
     ]}
  ]

# Configures Elixir's Logger
config :logger, :default_formatter,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Use Jason for JSON parsing in Phoenix
config :phoenix, :json_library, Jason

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.
import_config "#{config_env()}.exs"
