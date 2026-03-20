defmodule ElixirBackendWeb.Schema.ConsentTypes do
  use Absinthe.Schema.Notation

  alias ElixirBackend.Translations

  enum :consent_action do
    value(:accepted, as: "ACCEPTED")
    value(:rejected, as: "REJECTED")
  end

  object :consent do
    field :id, non_null(:id)
    field :key, non_null(:string)
    field :version, non_null(:integer)
    field :title, non_null(:string)
    field :short_text, :string
    field :body, :string
    field :url, :string
    field :published_at, :datetime
    field :managed_by, :string

    field :translation_status, non_null(list_of(non_null(:translation_field_status))) do
      resolve(fn consent, _, _ ->
        {:ok, Translations.translation_status(:consent, consent.id)}
      end)
    end
  end

  object :user_consent do
    field :id, non_null(:id)
    field :consent_key, non_null(:string)
    field :action, non_null(:consent_action)
    field :occurred_at, non_null(:datetime)
  end

  object :user_consent_history_entry do
    field :id, non_null(:id)
    field :consent_key, non_null(:string)
    field :action, non_null(:consent_action)
    field :occurred_at, non_null(:datetime)
    field :source, :string
  end

  input_object :create_consent_input do
    field :key, non_null(:string)
    field :version, :integer
    field :title, non_null(:string)
    field :short_text, :string
    field :body, :string
    field :url, :string
    field :published_at, :datetime
    field :managed_by, :string
  end

  input_object :update_consent_input do
    field :title, :string
    field :short_text, :string
    field :body, :string
    field :url, :string
    field :published_at, :datetime
    field :managed_by, :string
  end
end
