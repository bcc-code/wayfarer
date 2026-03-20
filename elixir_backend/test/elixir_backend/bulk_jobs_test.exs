defmodule ElixirBackend.BulkJobsTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.BulkJobs
  import ElixirBackend.TestHelpers

  setup do
    project = create_project()
    user = create_user()
    %{project: project, user: user}
  end

  describe "create_job/1" do
    test "creates a bulk job", %{project: project, user: user} do
      {:ok, job} =
        BulkJobs.create_job(%{
          operation_type: "BULK_SCORE_ADJUSTMENT",
          created_by: user.id,
          project_id: project.id,
          total_count: 100
        })

      assert job.status == "PENDING"
      assert job.operation_type == "BULK_SCORE_ADJUSTMENT"
      assert String.starts_with?(job.id, "BJ")
    end
  end

  describe "get_job/1" do
    test "returns job by id", %{user: user} do
      {:ok, job} =
        BulkJobs.create_job(%{
          operation_type: "TEST",
          created_by: user.id
        })

      assert {:ok, found} = BulkJobs.get_job(job.id)
      assert found.id == job.id
    end

    test "returns error for missing job" do
      assert {:error, :not_found} = BulkJobs.get_job("BJ00000000000000000000000000")
    end
  end

  describe "mark_processing/1" do
    test "transitions to PROCESSING", %{user: user} do
      {:ok, job} = BulkJobs.create_job(%{operation_type: "TEST", created_by: user.id})
      {:ok, updated} = BulkJobs.mark_processing(job.id)
      assert updated.status == "PROCESSING"
      assert updated.started_at != nil
    end
  end

  describe "mark_completed/1" do
    test "transitions to COMPLETED", %{user: user} do
      {:ok, job} = BulkJobs.create_job(%{operation_type: "TEST", created_by: user.id})
      {:ok, updated} = BulkJobs.mark_completed(job.id)
      assert updated.status == "COMPLETED"
      assert updated.completed_at != nil
    end
  end

  describe "mark_failed/2" do
    test "transitions to FAILED with error", %{user: user} do
      {:ok, job} = BulkJobs.create_job(%{operation_type: "TEST", created_by: user.id})
      {:ok, updated} = BulkJobs.mark_failed(job.id, "Something went wrong")
      assert updated.status == "FAILED"
      assert updated.error_message == "Something went wrong"
    end
  end

  describe "update_progress/2" do
    test "updates processed/success/failure counts", %{user: user} do
      {:ok, job} =
        BulkJobs.create_job(%{operation_type: "TEST", created_by: user.id, total_count: 10})

      :ok =
        BulkJobs.update_progress(job.id, %{
          processed_count: 5,
          success_count: 4,
          failure_count: 1
        })

      {:ok, updated} = BulkJobs.get_job(job.id)
      assert updated.processed_count == 5
      assert updated.success_count == 4
      assert updated.failure_count == 1
    end
  end

  describe "my_jobs/2" do
    test "returns user's jobs", %{user: user} do
      {:ok, _} = BulkJobs.create_job(%{operation_type: "TEST1", created_by: user.id})
      {:ok, _} = BulkJobs.create_job(%{operation_type: "TEST2", created_by: user.id})

      jobs = BulkJobs.my_jobs(user.id)
      assert length(jobs) == 2
    end
  end

  describe "list_jobs/2" do
    test "returns paginated jobs with filter", %{project: project, user: user} do
      {:ok, _} =
        BulkJobs.create_job(%{
          operation_type: "TEST",
          created_by: user.id,
          project_id: project.id
        })

      {:ok, conn} = BulkJobs.list_jobs(%{project_id: project.id}, %{first: 10})
      assert conn.total_count >= 1
    end

    test "filters by status", %{user: user} do
      {:ok, job} = BulkJobs.create_job(%{operation_type: "TEST", created_by: user.id})
      BulkJobs.mark_completed(job.id)

      {:ok, conn} = BulkJobs.list_jobs(%{status: "COMPLETED"}, %{first: 10})
      assert conn.total_count >= 1
    end
  end
end
