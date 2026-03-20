defmodule ElixirBackend.PushNotifications.PushSubscription do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  schema "push_subscriptions" do
    field :endpoint, :string
    field :p256dh_key, :string
    field :auth_key, :string
    field :user_agent, :string

    belongs_to :user, ElixirBackend.Accounts.User, type: :string

    timestamps(type: :utc_datetime)
  end

  def changeset(sub, attrs) do
    sub
    |> cast(attrs, [:id, :user_id, :endpoint, :p256dh_key, :auth_key, :user_agent])
    |> validate_required([:id, :user_id, :endpoint, :p256dh_key, :auth_key])
    |> foreign_key_constraint(:user_id)
    |> unique_constraint(:endpoint)
  end
end
