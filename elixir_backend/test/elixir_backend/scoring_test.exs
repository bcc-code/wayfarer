defmodule ElixirBackend.ScoringTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.Scoring
  import ElixirBackend.TestHelpers

  setup do
    project = create_project()
    user = create_user()
    %{project: project, user: user}
  end

  # ── Score Journal CRUD ──

  describe "create_entry/1" do
    test "creates a score journal entry", %{project: project, user: user} do
      {:ok, entry} =
        Scoring.create_entry(%{
          project_id: project.id,
          user_id: user.id,
          points: 100,
          source_type: "MANUAL",
          reason: "Test adjustment"
        })

      assert entry.points == 100
      assert entry.source_type == "MANUAL"
      assert String.starts_with?(entry.id, "SJ")
    end

    test "validates source_type", %{project: project, user: user} do
      {:error, changeset} =
        Scoring.create_entry(%{
          project_id: project.id,
          user_id: user.id,
          points: 50,
          source_type: "INVALID"
        })

      assert errors_on(changeset).source_type != nil
    end
  end

  describe "create_adjustment/1" do
    test "creates a MANUAL entry", %{project: project, user: user} do
      {:ok, entry} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 50,
          reason: "Bonus"
        })

      assert entry.source_type == "MANUAL"
      assert entry.points == 50
    end

    test "supports negative adjustments", %{project: project, user: user} do
      {:ok, entry} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: -25,
          reason: "Penalty"
        })

      assert entry.points == -25
    end
  end

  describe "get_entry/1" do
    test "returns entry by id", %{project: project, user: user} do
      {:ok, created} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 10
        })

      assert {:ok, found} = Scoring.get_entry(created.id)
      assert found.id == created.id
    end

    test "returns error for missing entry" do
      assert {:error, :not_found} = Scoring.get_entry("SJ00000000000000000000000000")
    end
  end

  describe "delete_entry/1" do
    test "deletes an entry", %{project: project, user: user} do
      {:ok, entry} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 10
        })

      assert {:ok, _} = Scoring.delete_entry(entry.id)
      assert {:error, :not_found} = Scoring.get_entry(entry.id)
    end
  end

  describe "list_entries/2" do
    test "returns paginated entries", %{project: project, user: user} do
      {:ok, _} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 10
        })

      {:ok, conn} = Scoring.list_entries(%{project_id: project.id}, %{first: 10})
      assert conn.total_count >= 1
    end

    test "filters by user_id", %{project: project, user: user} do
      other_user = create_user()

      {:ok, _} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 10
        })

      {:ok, _} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: other_user.id,
          points: 20
        })

      {:ok, conn} = Scoring.list_entries(%{user_id: user.id}, %{first: 10})
      assert conn.total_count == 1
    end

    test "filters by source_type", %{project: project, user: user} do
      {:ok, _} =
        Scoring.create_entry(%{
          project_id: project.id,
          user_id: user.id,
          points: 10,
          source_type: "ACHIEVEMENT",
          source_id: "AC00000000000000000000000000"
        })

      {:ok, _} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 5
        })

      {:ok, conn} = Scoring.list_entries(%{source_type: "ACHIEVEMENT"}, %{first: 10})
      assert conn.total_count == 1
    end
  end

  # ── Leaderboard Updates ──

  describe "update_leaderboards/1" do
    test "creates person leaderboard entry", %{project: project, user: user} do
      {:ok, entry} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 100
        })

      Scoring.update_leaderboards(entry)

      {:ok, lb} = Scoring.get_project_leaderboard(project.id, "PERSONS", %{}, %{first: 10})
      assert lb.total_count == 1
      person = hd(lb.edges).node
      assert person.score == 100
      assert person.name == user.name
    end

    test "accumulates scores across multiple entries", %{project: project, user: user} do
      {:ok, e1} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 50
        })

      Scoring.update_leaderboards(e1)

      {:ok, e2} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 30
        })

      Scoring.update_leaderboards(e2)

      {:ok, lb} = Scoring.get_project_leaderboard(project.id, "PERSONS", %{}, %{first: 10})
      person = hd(lb.edges).node
      assert person.score == 80
    end

    test "updates church leaderboard", %{project: project, user: user} do
      {:ok, entry} =
        Scoring.create_adjustment(%{
          project_id: project.id,
          user_id: user.id,
          points: 75
        })

      Scoring.update_leaderboards(entry)

      {:ok, lb} = Scoring.get_project_leaderboard(project.id, "CHURCHES", %{}, %{first: 10})
      assert lb.total_count == 1
      assert hd(lb.edges).node.score == 75
    end
  end

  # ── Team Score Adjustments ──

  describe "create_team_adjustment/1" do
    test "distributes points with EACH mode", %{project: project} do
      team =
        ElixirBackend.Repo.insert!(%ElixirBackend.Teams.Team{
          id: ElixirBackend.ULID.new_team_id(),
          name: "Test Team",
          project_id: project.id,
          join_code: "TESTJOIN"
        })

      user1 = create_user()
      user2 = create_user()

      ElixirBackend.Repo.insert!(%ElixirBackend.Teams.TeamMember{
        team_id: team.id,
        user_id: user1.id,
        joined_at: DateTime.utc_now() |> DateTime.truncate(:second)
      })

      ElixirBackend.Repo.insert!(%ElixirBackend.Teams.TeamMember{
        team_id: team.id,
        user_id: user2.id,
        joined_at: DateTime.utc_now() |> DateTime.truncate(:second)
      })

      {:ok, entries} =
        Scoring.create_team_adjustment(%{
          team_id: team.id,
          project_id: project.id,
          points: 100,
          distribution_mode: "EACH",
          reason: "Team bonus"
        })

      assert length(entries) == 2
      assert Enum.all?(entries, fn e -> e.points == 100 end)
    end

    test "distributes points with SPLIT mode", %{project: project} do
      team =
        ElixirBackend.Repo.insert!(%ElixirBackend.Teams.Team{
          id: ElixirBackend.ULID.new_team_id(),
          name: "Split Team",
          project_id: project.id,
          join_code: "SPLITJN"
        })

      user1 = create_user()
      user2 = create_user()

      ElixirBackend.Repo.insert!(%ElixirBackend.Teams.TeamMember{
        team_id: team.id,
        user_id: user1.id,
        joined_at: DateTime.utc_now() |> DateTime.truncate(:second)
      })

      ElixirBackend.Repo.insert!(%ElixirBackend.Teams.TeamMember{
        team_id: team.id,
        user_id: user2.id,
        joined_at: DateTime.utc_now() |> DateTime.truncate(:second)
      })

      {:ok, entries} =
        Scoring.create_team_adjustment(%{
          team_id: team.id,
          project_id: project.id,
          points: 100,
          distribution_mode: "SPLIT",
          reason: "Split bonus"
        })

      assert length(entries) == 2
      assert Enum.all?(entries, fn e -> e.points == 50 end)
    end
  end
end
