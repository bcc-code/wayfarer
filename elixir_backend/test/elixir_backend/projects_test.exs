defmodule ElixirBackend.ProjectsTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Projects

  import ElixirBackend.TestHelpers

  describe "get_project/1" do
    test "returns project by id" do
      project = create_project(%{name: "My Project"})
      assert {:ok, found} = Projects.get_project(project.id)
      assert found.id == project.id
      assert found.name == "My Project"
    end

    test "returns error for nonexistent project" do
      assert {:error, :not_found} = Projects.get_project("PR00000000000000000000000000")
    end
  end

  describe "list_projects/2" do
    test "returns paginated projects" do
      create_project(%{name: "Project A"})
      create_project(%{name: "Project B"})
      create_project(%{name: "Project C"})

      assert {:ok, result} = Projects.list_projects(%{}, %{first: 2})
      assert length(result.edges) == 2
      assert result.total_count >= 3
      assert result.page_info.has_next_page == true
    end

    test "filters by archived" do
      create_project(%{name: "Active"})
      p = create_project(%{name: "Archived"})
      Projects.archive_project(p.id)

      assert {:ok, result} = Projects.list_projects(%{archived: false}, %{first: 10})
      assert Enum.all?(result.edges, fn e -> e.node.archived == false end)
    end

    test "filters by ids" do
      p1 = create_project(%{name: "Project 1"})
      _p2 = create_project(%{name: "Project 2"})

      assert {:ok, result} = Projects.list_projects(%{ids: [p1.id]}, %{first: 10})
      assert result.total_count == 1
      assert hd(result.edges).node.id == p1.id
    end

    test "filters by date range" do
      create_project(%{
        name: "Early",
        start_date: ~U[2025-01-01 00:00:00Z],
        end_date: ~U[2025-06-30 23:59:59Z]
      })

      create_project(%{
        name: "Late",
        start_date: ~U[2026-07-01 00:00:00Z],
        end_date: ~U[2026-12-31 23:59:59Z]
      })

      assert {:ok, result} =
               Projects.list_projects(
                 %{start_date_after: ~U[2026-01-01 00:00:00Z]},
                 %{first: 10}
               )

      assert Enum.all?(result.edges, fn e -> e.node.name == "Late" end)
    end
  end

  describe "create_project/1" do
    test "creates a project with valid attrs" do
      assert {:ok, project} =
               Projects.create_project(%{
                 name: "New Project",
                 description: "Description",
                 start_date: ~U[2026-01-01 00:00:00Z],
                 end_date: ~U[2026-12-31 23:59:59Z]
               })

      assert String.starts_with?(project.id, "PR")
      assert project.name == "New Project"
      assert project.description == "Description"
    end

    test "fails without required fields" do
      assert {:error, _changeset} = Projects.create_project(%{name: "No Dates"})
    end
  end

  describe "update_project/2" do
    test "updates project fields" do
      project = create_project()
      assert {:ok, updated} = Projects.update_project(project.id, %{name: "Updated"})
      assert updated.name == "Updated"
    end

    test "returns error for nonexistent project" do
      assert {:error, :not_found} =
               Projects.update_project("PR00000000000000000000000000", %{name: "X"})
    end
  end

  describe "delete_project/1" do
    test "deletes a project" do
      project = create_project()
      assert {:ok, _} = Projects.delete_project(project.id)
      assert {:error, :not_found} = Projects.get_project(project.id)
    end

    test "returns error for nonexistent project" do
      assert {:error, :not_found} = Projects.delete_project("PR00000000000000000000000000")
    end
  end

  describe "archive_project/1" do
    test "archives a project" do
      project = create_project()
      assert {:ok, archived} = Projects.archive_project(project.id)
      assert archived.archived == true
    end
  end

  describe "my_projects/1" do
    test "returns projects user has joined" do
      user = create_user()
      p1 = create_project(%{name: "Joined"})
      _p2 = create_project(%{name: "Not Joined"})
      create_user_project(user, p1)

      assert {:ok, projects} = Projects.my_projects(user.id)
      assert length(projects) == 1
      assert hd(projects).id == p1.id
    end

    test "returns empty list for user with no projects" do
      user = create_user()
      assert {:ok, []} = Projects.my_projects(user.id)
    end
  end

  describe "join_project/2" do
    test "joins a user to a project" do
      user = create_user()
      project = create_project()

      assert {:ok, joined} = Projects.join_project(user.id, project.id)
      assert joined.id == project.id

      assert {:ok, [p]} = Projects.my_projects(user.id)
      assert p.id == project.id
    end

    test "idempotent - joining twice does not error" do
      user = create_user()
      project = create_project()

      assert {:ok, _} = Projects.join_project(user.id, project.id)
      assert {:ok, _} = Projects.join_project(user.id, project.id)

      assert {:ok, [_]} = Projects.my_projects(user.id)
    end
  end
end
