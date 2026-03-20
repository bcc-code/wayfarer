defmodule ElixirBackend.AchievementsTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Achievements
  alias ElixirBackend.ExternalContent, as: EC

  import ElixirBackend.TestHelpers

  defp create_simple_achievement(project, attrs \\ %{}) do
    defaults = %{
      name: "Test Achievement",
      description_pending: "Not done yet",
      description_completed: "Done!",
      notification_text: "You got it!",
      image_pending: "https://example.com/pending.png",
      image_completed: "https://example.com/completed.png",
      project_id: project.id,
      points: 10,
      hidden: false
    }

    {:ok, a} = Achievements.create_simple_achievement(Map.merge(defaults, attrs))
    a
  end

  defp create_external_content(attrs \\ %{}) do
    defaults = %{
      plan_id: "plan-#{System.unique_integer([:positive])}",
      task_id: "task-#{System.unique_integer([:positive])}",
      content_type: "ARTICLE",
      source: "ssf"
    }

    {:ok, c} = EC.upsert_content(Map.merge(defaults, attrs))
    c
  end

  describe "create_simple_achievement/1" do
    test "creates a simple achievement" do
      project = create_project()

      assert {:ok, a} =
               Achievements.create_simple_achievement(%{
                 name: "Simple",
                 description_pending: "Pending",
                 description_completed: "Done",
                 image_pending: "pending.png",
                 image_completed: "done.png",
                 project_id: project.id,
                 points: 50,
                 hidden: false
               })

      assert String.starts_with?(a.id, "AC")
      assert a.achievement_type == "SIMPLE"
      assert a.points == 50
    end
  end

  describe "create_content_achievement/1" do
    test "creates a content achievement with items" do
      project = create_project()
      ec1 = create_external_content()
      ec2 = create_external_content()

      assert {:ok, a} =
               Achievements.create_content_achievement(%{
                 name: "Content",
                 description_pending: "Read all",
                 description_completed: "All read",
                 image_pending: "p.png",
                 image_completed: "c.png",
                 project_id: project.id,
                 points: 20,
                 hidden: false,
                 items: [
                   %{external_content_id: ec1.id},
                   %{external_content_id: ec2.id}
                 ]
               })

      assert a.achievement_type == "CONTENT"
      items = Achievements.get_content_items(a.id)
      assert length(items) == 2
    end
  end

  describe "create_streak_achievement/1" do
    test "creates a streak achievement" do
      project = create_project()
      streak = create_streak(project)

      assert {:ok, a} =
               Achievements.create_streak_achievement(%{
                 name: "Streak",
                 description_pending: "Keep going",
                 description_completed: "Streak master",
                 image_pending: "p.png",
                 image_completed: "c.png",
                 project_id: project.id,
                 points: 30,
                 hidden: false,
                 streak_id: streak.id,
                 needed_streak: 7
               })

      assert a.achievement_type == "STREAK"
      sa = Achievements.get_streak_achievement_data(a.id)
      assert sa.needed_streak == 7
      assert sa.streak_id == streak.id
    end
  end

  describe "get_achievement/1" do
    test "returns achievement by id" do
      project = create_project()
      a = create_simple_achievement(project)
      assert {:ok, found} = Achievements.get_achievement(a.id)
      assert found.id == a.id
    end

    test "returns error for non-existent id" do
      assert {:error, :not_found} = Achievements.get_achievement("AC00000000000000000000000000")
    end
  end

  describe "list_achievements/2" do
    test "lists achievements with project filter" do
      p1 = create_project(%{name: "P1"})
      p2 = create_project(%{name: "P2"})
      create_simple_achievement(p1, %{name: "A1"})
      create_simple_achievement(p2, %{name: "A2"})

      {:ok, conn} = Achievements.list_achievements(%{project_id: p1.id})
      assert conn.total_count == 1
    end
  end

  describe "update_achievement/2" do
    test "updates name and points" do
      project = create_project()
      a = create_simple_achievement(project)

      assert {:ok, updated} =
               Achievements.update_achievement(a.id, %{name: "Updated", points: 99})

      assert updated.name == "Updated"
      assert updated.points == 99
    end
  end

  describe "delete_achievement/1" do
    test "deletes an achievement" do
      project = create_project()
      a = create_simple_achievement(project)

      assert {:ok, _} = Achievements.delete_achievement(a.id)
      assert {:error, :not_found} = Achievements.get_achievement(a.id)
    end
  end

  describe "award_achievement/2" do
    test "awards achievement to user" do
      project = create_project()
      user = create_user()
      a = create_simple_achievement(project)

      assert {:ok, _} = Achievements.award_achievement(user.id, a.id)
      assert Achievements.get_user_achieved_at(a.id, user.id) != nil
    end

    test "idempotent - awarding twice is safe" do
      project = create_project()
      user = create_user()
      a = create_simple_achievement(project)

      assert {:ok, _} = Achievements.award_achievement(user.id, a.id)
      assert {:ok, _} = Achievements.award_achievement(user.id, a.id)
    end

    test "rejects award when awardable_from is in the future" do
      project = create_project()
      user = create_user()
      future = DateTime.utc_now() |> DateTime.add(86400) |> DateTime.truncate(:second)
      a = create_simple_achievement(project, %{awardable_from: future})

      assert {:error, "achievement is not yet available for awarding"} =
               Achievements.award_achievement(user.id, a.id)
    end

    test "allows award when awardable_from is in the past" do
      project = create_project()
      user = create_user()
      past = DateTime.utc_now() |> DateTime.add(-86400) |> DateTime.truncate(:second)
      a = create_simple_achievement(project, %{awardable_from: past})

      assert {:ok, _} = Achievements.award_achievement(user.id, a.id)
    end
  end

  describe "revoke_achievement/2" do
    test "revokes an awarded achievement" do
      project = create_project()
      user = create_user()
      a = create_simple_achievement(project)

      Achievements.award_achievement(user.id, a.id)
      assert {:ok, true} = Achievements.revoke_achievement(user.id, a.id)
      assert Achievements.get_user_achieved_at(a.id, user.id) == nil
    end

    test "returns false for non-existent award" do
      user = create_user()

      assert {:ok, false} =
               Achievements.revoke_achievement(user.id, "AC00000000000000000000000000")
    end
  end

  describe "content progress" do
    test "marks content completed and auto-awards" do
      project = create_project()
      user = create_user()
      ec = create_external_content()

      {:ok, a} =
        Achievements.create_content_achievement(%{
          name: "Single Item",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false,
          items: [%{external_content_id: ec.id}]
        })

      {:ok, achievements} = Achievements.mark_content_completed(user.id, ec.id)
      assert length(achievements) == 1

      # Should be auto-awarded
      assert Achievements.get_user_achieved_at(a.id, user.id) != nil
    end

    test "partial completion does not auto-award" do
      project = create_project()
      user = create_user()
      ec1 = create_external_content()
      ec2 = create_external_content()

      {:ok, a} =
        Achievements.create_content_achievement(%{
          name: "Two Items",
          description_pending: "P",
          description_completed: "C",
          image_pending: "p.png",
          image_completed: "c.png",
          project_id: project.id,
          points: 10,
          hidden: false,
          items: [%{external_content_id: ec1.id}, %{external_content_id: ec2.id}]
        })

      Achievements.mark_content_completed(user.id, ec1.id)
      assert Achievements.get_user_achieved_at(a.id, user.id) == nil
      assert Achievements.get_completed_item_count(a.id, user.id) == 1
    end
  end

  describe "reorder_achievements/2" do
    test "reorders achievements in a project" do
      project = create_project()
      a1 = create_simple_achievement(project, %{name: "First"})
      a2 = create_simple_achievement(project, %{name: "Second"})

      {:ok, reordered} = Achievements.reorder_achievements(project.id, [a2.id, a1.id])
      assert length(reordered) == 2
      assert hd(reordered).id == a2.id
    end
  end
end
