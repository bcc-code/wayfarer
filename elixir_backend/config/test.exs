import Config

# Configure your database
#
# The MIX_TEST_PARTITION environment variable can be used
# to provide built-in test partitioning in CI environment.
# Run `mix help test` for more information.
config :elixir_backend, ElixirBackend.Repo,
  username: "postgres",
  password: "postgres",
  hostname: "localhost",
  database: "elixir_backend_test#{System.get_env("MIX_TEST_PARTITION")}",
  pool: Ecto.Adapters.SQL.Sandbox,
  pool_size: System.schedulers_online() * 2

# We don't run a server during test. If one is required,
# you can enable the server option below.
config :elixir_backend, ElixirBackendWeb.Endpoint,
  http: [ip: {127, 0, 0, 1}, port: 4002],
  secret_key_base: "15HfodpZpj2S+CFg/BZLR9+Bz2mqkgPMhIz4Hx7bfVpei+1SBCeKfYgtR8fOJazB",
  server: false

# In test we don't send emails
config :elixir_backend, ElixirBackend.Mailer, adapter: Swoosh.Adapters.Test

# Disable swoosh api client as it is only required for production adapters
config :swoosh, :api_client, false

# Print only warnings and errors during test
config :logger, level: :warning

# Disable Oban in tests (use Oban.Testing)
config :elixir_backend, Oban, testing: :inline

# Initialize plugs at runtime for faster test compilation
config :phoenix, :plug_init_mode, :runtime
