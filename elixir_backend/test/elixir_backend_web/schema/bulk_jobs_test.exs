defmodule ElixirBackendWeb.Schema.BulkJobsTest do
  use ElixirBackendWeb.ConnCase, async: true

  import ElixirBackend.TestHelpers

  setup do
    user = create_user()
    %{user: user}
  end

  describe "bulkJob query" do
    test "returns job by id", %{conn: conn, user: user} do
      {:ok, job} =
        ElixirBackend.BulkJobs.create_job(%{
          operation_type: "TEST",
          created_by: user.id
        })

      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query($id: ID!) {
              bulkJob(id: $id) {
                id
                operationType
                status
                totalCount
                processedCount
              }
            }
          """,
          %{"id" => job.id}
        )

      data = json_response(resp, 200)["data"]["bulkJob"]
      assert data["id"] == job.id
      assert data["status"] == "PENDING"
    end
  end

  describe "myBulkJobs query" do
    test "returns user's jobs", %{conn: conn, user: user} do
      {:ok, _} =
        ElixirBackend.BulkJobs.create_job(%{
          operation_type: "TEST",
          created_by: user.id
        })

      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query {
              myBulkJobs(limit: 5) {
                id
                operationType
                status
              }
            }
          """
        )

      data = json_response(resp, 200)["data"]["myBulkJobs"]
      assert data != []
    end
  end

  describe "bulkJobs query (admin)" do
    test "admin lists all jobs", %{conn: conn, user: user} do
      {:ok, _} =
        ElixirBackend.BulkJobs.create_job(%{
          operation_type: "TEST",
          created_by: user.id
        })

      conn = auth_conn(conn, user.id, ["admin"])

      resp =
        graphql_query(
          conn,
          """
            query($first: Int) {
              bulkJobs(first: $first) {
                totalCount
                edges { node { id status } }
              }
            }
          """,
          %{"first" => 10}
        )

      data = json_response(resp, 200)["data"]["bulkJobs"]
      assert data["totalCount"] >= 1
    end

    test "non-admin cannot list jobs", %{conn: conn, user: user} do
      conn = auth_conn(conn, user.id, ["user"])

      resp =
        graphql_query(
          conn,
          """
            query {
              bulkJobs(first: 10) {
                totalCount
              }
            }
          """
        )

      errors = json_response(resp, 200)["errors"]
      assert errors != nil
    end
  end
end
