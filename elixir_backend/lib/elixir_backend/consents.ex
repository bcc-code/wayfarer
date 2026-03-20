defmodule ElixirBackend.Consents do
  @moduledoc "Context for consent management."

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Consents.{Consent, UserConsentHistory}

  def get_consent(id) do
    case Repo.get(Consent, id) do
      nil -> {:error, :not_found}
      c -> {:ok, c}
    end
  end

  def list_consents do
    from(c in Consent,
      where: not is_nil(c.published_at),
      order_by: [asc: c.key, desc: c.version]
    )
    |> Repo.all()
  end

  def create_consent(attrs) do
    id = ULID.new_consent_id()

    %Consent{}
    |> Consent.changeset(Map.put(attrs, :id, id))
    |> Repo.insert()
  end

  def update_consent(id, attrs) do
    with {:ok, consent} <- get_consent(id) do
      consent
      |> Consent.changeset(attrs)
      |> Repo.update()
    end
  end

  def accept_consent(user_id, consent_id) do
    record_consent_action(user_id, consent_id, "ACCEPTED")
  end

  def reject_consent(user_id, consent_id) do
    record_consent_action(user_id, consent_id, "REJECTED")
  end

  def get_user_history(user_id) do
    from(h in UserConsentHistory,
      where: h.user_id == ^user_id,
      order_by: [desc: h.occurred_at]
    )
    |> Repo.all()
  end

  def pending_consents(user_id) do
    all = list_consents()
    history = get_user_history(user_id)
    acted_keys = MapSet.new(history, fn h -> h.consent_key end)
    Enum.reject(all, fn c -> MapSet.member?(acted_keys, c.key) end)
  end

  defp record_consent_action(user_id, consent_id, action) do
    with {:ok, consent} <- get_consent(consent_id) do
      id = ULID.new_consent_history_id()
      now = DateTime.utc_now() |> DateTime.truncate(:second)

      %UserConsentHistory{}
      |> UserConsentHistory.changeset(%{
        id: id,
        user_id: user_id,
        consent_id: consent_id,
        consent_key: consent.key,
        action: action,
        occurred_at: now
      })
      |> Repo.insert()
    end
  end
end
