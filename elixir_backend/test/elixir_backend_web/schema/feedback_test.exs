defmodule ElixirBackendWeb.Schema.FeedbackTest do
  use ElixirBackendWeb.ConnCase, async: true

  alias ElixirBackend.Feedback
  alias ElixirBackend.TestHelpers

  describe "submitFeedback mutation" do
    test "submits feedback", %{conn: conn} do
      user = TestHelpers.create_user()

      mutation = """
      mutation($input: SubmitFeedbackInput!) {
        submitFeedback(input: $input) {
          id
          message
          canContactMe
        }
      }
      """

      resp =
        conn
        |> TestHelpers.auth_conn(user.id, ["user"])
        |> TestHelpers.graphql_query(mutation, %{
          "input" => %{
            "message" => "This is great!",
            "canContactMe" => true
          }
        })
        |> json_response(200)

      fb = resp["data"]["submitFeedback"]
      assert fb["message"] == "This is great!"
      assert fb["canContactMe"] == true
    end
  end

  describe "feedback query (admin)" do
    test "lists feedback with admin role", %{conn: conn} do
      user = TestHelpers.create_user()

      Feedback.submit_feedback(%{
        user_id: user.id,
        message: "Admin visible feedback",
        can_contact_me: false
      })

      query = """
      query {
        feedback {
          edges {
            node {
              id
              message
            }
          }
          totalCount
        }
      }
      """

      resp =
        conn
        |> TestHelpers.auth_conn(user.id, ["admin"])
        |> TestHelpers.graphql_query(query)
        |> json_response(200)

      data = resp["data"]["feedback"]
      assert data["totalCount"] >= 1
    end
  end

  describe "markFeedbackHandled mutation" do
    test "marks feedback as handled", %{conn: conn} do
      user = TestHelpers.create_user()

      {:ok, fb} =
        Feedback.submit_feedback(%{
          user_id: user.id,
          message: "Handle this",
          can_contact_me: false
        })

      mutation = """
      mutation($feedbackId: ID!) {
        markFeedbackHandled(feedbackId: $feedbackId) {
          id
          handledAt
        }
      }
      """

      resp =
        conn
        |> TestHelpers.auth_conn(user.id, ["admin"])
        |> TestHelpers.graphql_query(mutation, %{"feedbackId" => fb.id})
        |> json_response(200)

      assert resp["data"]["markFeedbackHandled"]["handledAt"] != nil
    end
  end
end
