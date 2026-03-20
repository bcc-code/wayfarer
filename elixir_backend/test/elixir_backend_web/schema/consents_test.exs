defmodule ElixirBackendWeb.Schema.ConsentsTest do
  use ElixirBackendWeb.ConnCase, async: true

  alias ElixirBackend.Consents
  alias ElixirBackend.TestHelpers

  describe "consents query" do
    test "returns published consents", %{conn: conn} do
      {:ok, _} =
        Consents.create_consent(%{
          key: "gql-consent",
          title: "GQL Consent",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      query = """
      query {
        consents {
          id
          key
          title
        }
      }
      """

      resp =
        conn
        |> TestHelpers.graphql_query(query)
        |> json_response(200)

      consents = resp["data"]["consents"]
      assert is_list(consents)
      assert Enum.any?(consents, fn c -> c["key"] == "gql-consent" end)
    end
  end

  describe "acceptConsent mutation" do
    test "accepts a consent for current user", %{conn: conn} do
      user = TestHelpers.create_user()

      {:ok, consent} =
        Consents.create_consent(%{
          key: "accept-gql",
          title: "Accept GQL",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      mutation = """
      mutation($consentId: ID!) {
        acceptConsent(consentId: $consentId) {
          id
          action
          occurredAt
        }
      }
      """

      resp =
        conn
        |> TestHelpers.auth_conn(user.id, ["user"])
        |> TestHelpers.graphql_query(mutation, %{"consentId" => consent.id})
        |> json_response(200)

      entry = resp["data"]["acceptConsent"]
      assert entry["action"] == "ACCEPTED"
    end
  end

  describe "createConsent mutation" do
    test "admin can create a consent", %{conn: conn} do
      user = TestHelpers.create_user()

      mutation = """
      mutation($input: CreateConsentInput!) {
        createConsent(input: $input) {
          id
          key
          title
        }
      }
      """

      resp =
        conn
        |> TestHelpers.auth_conn(user.id, ["admin"])
        |> TestHelpers.graphql_query(mutation, %{
          "input" => %{
            "key" => "admin-created",
            "title" => "Admin Created Consent",
            "body" => "Some body text"
          }
        })
        |> json_response(200)

      consent = resp["data"]["createConsent"]
      assert consent["key"] == "admin-created"
    end
  end

  describe "pendingConsents query" do
    test "returns consents user hasn't acted on", %{conn: conn} do
      user = TestHelpers.create_user()

      {:ok, _} =
        Consents.create_consent(%{
          key: "pending-gql",
          title: "Pending GQL",
          published_at: ~U[2026-01-01 00:00:00Z]
        })

      query = """
      query {
        pendingConsents {
          id
          key
        }
      }
      """

      resp =
        conn
        |> TestHelpers.auth_conn(user.id, ["user"])
        |> TestHelpers.graphql_query(query)
        |> json_response(200)

      pending = resp["data"]["pendingConsents"]
      assert Enum.any?(pending, fn c -> c["key"] == "pending-gql" end)
    end
  end
end
