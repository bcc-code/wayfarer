defmodule ElixirBackend.StreaksTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Streaks

  import ElixirBackend.TestHelpers

  defp create_streak_fixture(project, attrs \\ %{}) do
    default_attrs = %{
      name: "Daily Streak",
      description: "Read every day",
      project_id: project.id,
      relevant_days: [
        %{start: ~D[2026-03-01], end: ~D[2026-03-31]}
      ]
    }

    {:ok, streak} = Streaks.create_streak(Map.merge(default_attrs, attrs))
    streak
  end

  describe "create_streak/1" do
    test "creates a streak with relevant days" do
      project = create_project()

      assert {:ok, streak} =
               Streaks.create_streak(%{
                 name: "Daily Streak",
                 description: "Read every day",
                 project_id: project.id,
                 relevant_days: [
                   %{start: ~D[2026-03-01], end: ~D[2026-03-15]},
                   %{start: ~D[2026-03-20], end: ~D[2026-03-31]}
                 ]
               })

      assert String.starts_with?(streak.id, "SK")
      assert streak.name == "Daily Streak"
      assert streak.description == "Read every day"

      days = Streaks.get_relevant_days(streak.id)
      assert length(days) == 2
    end

    test "creates a streak with empty relevant days" do
      project = create_project()

      assert {:ok, streak} =
               Streaks.create_streak(%{
                 name: "Empty Streak",
                 description: "",
                 project_id: project.id,
                 relevant_days: []
               })

      days = Streaks.get_relevant_days(streak.id)
      assert Enum.empty?(days)
    end
  end

  describe "get_streak/1" do
    test "returns streak by id" do
      project = create_project()
      streak = create_streak_fixture(project)

      assert {:ok, found} = Streaks.get_streak(streak.id)
      assert found.id == streak.id
    end

    test "returns error for non-existent id" do
      assert {:error, :not_found} = Streaks.get_streak("SK00000000000000000000000000")
    end
  end

  describe "list_streaks/2" do
    test "lists all streaks" do
      project = create_project()
      create_streak_fixture(project, %{name: "Streak 1"})
      create_streak_fixture(project, %{name: "Streak 2"})

      {:ok, connection} = Streaks.list_streaks()
      assert connection.total_count >= 2
    end

    test "filters by project_id" do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      create_streak_fixture(p1, %{name: "S1"})
      create_streak_fixture(p2, %{name: "S2"})

      {:ok, conn1} = Streaks.list_streaks(%{project_id: p1.id})
      {:ok, conn2} = Streaks.list_streaks(%{project_id: p2.id})

      assert conn1.total_count == 1
      assert conn2.total_count == 1
    end

    test "filters by ids" do
      project = create_project()
      s1 = create_streak_fixture(project, %{name: "S1"})
      _s2 = create_streak_fixture(project, %{name: "S2"})

      {:ok, connection} = Streaks.list_streaks(%{ids: [s1.id]})
      assert connection.total_count == 1
    end
  end

  describe "update_streak/2" do
    test "updates streak name and description" do
      project = create_project()
      streak = create_streak_fixture(project)

      assert {:ok, updated} = Streaks.update_streak(streak.id, %{name: "Updated"})
      assert updated.name == "Updated"
    end

    test "replaces relevant days" do
      project = create_project()
      streak = create_streak_fixture(project)

      new_days = [%{start: ~D[2026-04-01], end: ~D[2026-04-30]}]
      assert {:ok, _updated} = Streaks.update_streak(streak.id, %{relevant_days: new_days})

      days = Streaks.get_relevant_days(streak.id)
      assert length(days) == 1
      assert hd(days).start_date == ~D[2026-04-01]
    end
  end

  describe "delete_streak/1" do
    test "deletes a streak" do
      project = create_project()
      streak = create_streak_fixture(project)

      assert {:ok, _} = Streaks.delete_streak(streak.id)
      assert {:error, :not_found} = Streaks.get_streak(streak.id)
    end

    test "returns error for non-existent streak" do
      assert {:error, :not_found} = Streaks.delete_streak("SK00000000000000000000000000")
    end
  end

  describe "record_activity/3" do
    test "records activity for a user on a date" do
      project = create_project()
      streak = create_streak_fixture(project)
      user = create_user()

      assert {:ok, _} = Streaks.record_activity(streak.id, user.id, ~D[2026-03-10])
    end

    test "is idempotent for same date" do
      project = create_project()
      streak = create_streak_fixture(project)
      user = create_user()

      assert {:ok, _} = Streaks.record_activity(streak.id, user.id, ~D[2026-03-10])
      assert {:ok, _} = Streaks.record_activity(streak.id, user.id, ~D[2026-03-10])
    end
  end

  describe "get_streak_status/2" do
    test "returns 0 for no activity" do
      project = create_project()

      streak =
        create_streak_fixture(project, %{
          relevant_days: [%{start: ~D[2026-03-01], end: ~D[2026-03-31]}]
        })

      user = create_user()

      assert Streaks.get_streak_status(streak.id, user.id) == 0
    end

    test "counts consecutive days with activity" do
      project = create_project()
      today = Date.utc_today()
      yesterday = Date.add(today, -1)

      streak =
        create_streak_fixture(project, %{
          relevant_days: [%{start: Date.add(today, -30), end: Date.add(today, 30)}]
        })

      user = create_user()

      Streaks.record_activity(streak.id, user.id, yesterday)
      Streaks.record_activity(streak.id, user.id, today)

      assert Streaks.get_streak_status(streak.id, user.id) == 2
    end
  end

  describe "get_listened_days/3" do
    test "returns days with activity flags" do
      project = create_project()

      streak =
        create_streak_fixture(project, %{
          relevant_days: [%{start: ~D[2026-03-01], end: ~D[2026-03-05]}]
        })

      user = create_user()

      Streaks.record_activity(streak.id, user.id, ~D[2026-03-01])
      Streaks.record_activity(streak.id, user.id, ~D[2026-03-03])

      days = Streaks.get_listened_days(streak.id, user.id, 5)
      assert length(days) == 5

      active_dates = days |> Enum.filter(& &1.active) |> Enum.map(& &1.date)
      assert ~D[2026-03-01] in active_dates
      assert ~D[2026-03-03] in active_dates
    end

    test "limits to last N days" do
      project = create_project()

      streak =
        create_streak_fixture(project, %{
          relevant_days: [%{start: ~D[2026-03-01], end: ~D[2026-03-10]}]
        })

      user = create_user()

      days = Streaks.get_listened_days(streak.id, user.id, 3)
      assert length(days) == 3
    end
  end
end
