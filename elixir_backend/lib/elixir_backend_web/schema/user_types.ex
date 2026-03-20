defmodule ElixirBackendWeb.Schema.UserTypes do
  @moduledoc "Absinthe types for users and churches."
  use Absinthe.Schema.Notation

  import Absinthe.Resolution.Helpers, only: [dataloader: 1]

  alias ElixirBackend.Accounts

  # ── Enums ──

  enum :gender_enum do
    value(:male, as: "MALE")
    value(:female, as: "FEMALE")
    value(:unknown, as: "UNKNOWN")
  end

  # ── Objects ──

  object :user do
    field :id, non_null(:id)
    field :members_id, non_null(:string)
    field :person_uuid, :string
    field :gender, non_null(:gender_enum)
    field :church_id, :id
    field :church, :church, resolve: dataloader(ElixirBackend.Repo)
    field :church_locked_until, :datetime
    field :birthdate, :date

    field :age, :integer do
      resolve(fn user, _args, _res ->
        {:ok, Accounts.calculate_age(user.birthdate)}
      end)
    end

    field :email, non_null(:string)
    field :name, non_null(:string)
    field :avatar_url, :string
    field :display_name, :string
    field :language, :string
    field :created_at, non_null(:datetime), resolve: fn user, _, _ -> {:ok, user.inserted_at} end
  end

  object :church do
    field :id, non_null(:id)
    field :name, non_null(:string)
    field :country, non_null(:string)
    field :category, non_null(:string)
  end

  # ── Pagination ──

  object :user_edge do
    field :cursor, non_null(:string)
    field :node, non_null(:user)
  end

  object :user_connection do
    field :edges, non_null(list_of(non_null(:user_edge)))
    field :page_info, non_null(:page_info)
    field :total_count, non_null(:integer)
  end

  # ── Inputs ──

  input_object :user_filter do
    field :query, :string
    field :church_id, :id
    field :gender, :gender_enum
    field :min_age, :integer
    field :max_age, :integer
    field :project_id, :id
    field :event_id, :id
    field :ids, list_of(non_null(:id))
  end
end
