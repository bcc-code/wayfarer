defmodule ElixirBackend.Accounts.User do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}

  @genders ~w(MALE FEMALE UNKNOWN)

  schema "users" do
    field :name, :string
    field :members_id, :string
    field :person_uuid, Ecto.UUID
    field :email, :string
    field :gender, :string, default: "UNKNOWN"
    field :birthdate, :date
    field :church_locked_until, :utc_datetime
    field :avatar_url, :string
    field :display_name, :string
    field :language, :string, default: "en"

    belongs_to :church, ElixirBackend.Churches.Church, type: :string

    timestamps(type: :utc_datetime)
  end

  def changeset(user, attrs) do
    user
    |> cast(attrs, [:id, :name])
    |> validate_required([:id, :name])
  end

  def create_changeset(user, attrs) do
    user
    |> cast(attrs, [
      :id,
      :members_id,
      :person_uuid,
      :email,
      :name,
      :gender,
      :birthdate,
      :church_id,
      :church_locked_until,
      :avatar_url,
      :display_name,
      :language
    ])
    |> validate_required([:id, :members_id, :email, :name, :gender, :church_id])
    |> validate_inclusion(:gender, @genders)
    |> unique_constraint(:members_id)
    |> foreign_key_constraint(:church_id)
  end

  def update_changeset(user, attrs) do
    user
    |> cast(attrs, [
      :email,
      :name,
      :gender,
      :birthdate,
      :church_id,
      :avatar_url,
      :display_name,
      :language
    ])
    |> validate_inclusion(:gender, @genders)
    |> foreign_key_constraint(:church_id)
  end

  def church_lock_changeset(user, attrs) do
    user
    |> cast(attrs, [:church_locked_until])
  end
end
