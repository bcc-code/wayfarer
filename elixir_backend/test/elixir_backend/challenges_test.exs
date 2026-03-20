defmodule ElixirBackend.ChallengesTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Challenges
  alias ElixirBackend.Challenges.Challenge
  alias ElixirBackend.Pagination
  import ElixirBackend.TestHelpers

  # ── Create Challenge ──

  describe "create_challenge/1" do
    test "creates a SIMPLE challenge with valid attrs" do
      project = create_project()

      assert {:ok, %Challenge{} = challenge} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "SIMPLE",
                 name: "Test Challenge",
                 description: "A test",
                 button_text: "Complete"
               })

      assert challenge.challenge_type == "SIMPLE"
      assert challenge.name == "Test Challenge"
      assert challenge.allow_self_completion == true
      assert challenge.requires_team_membership == false
      assert challenge.requires_super_team_membership == false
      assert String.starts_with?(challenge.id, "CL")
      assert byte_size(challenge.id) == 28
      assert challenge.published_at != nil
    end

    test "creates a SIMPLE challenge without event" do
      project = create_project()

      assert {:ok, challenge} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "SIMPLE",
                 name: "No Event Challenge",
                 button_text: "Do"
               })

      assert challenge.event_id == nil
    end

    test "creates a SIMPLE challenge with event" do
      project = create_project()
      event = create_event(project)

      assert {:ok, challenge} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 event_id: event.id,
                 challenge_type: "SIMPLE",
                 name: "Event Challenge",
                 button_text: "Do"
               })

      assert challenge.event_id == event.id
    end

    test "creates an EXTERNAL challenge with url" do
      project = create_project()

      assert {:ok, challenge} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "EXTERNAL",
                 name: "External Challenge",
                 button_text: "Open",
                 url: "https://example.com"
               })

      assert challenge.challenge_type == "EXTERNAL"
      assert challenge.url == "https://example.com"
    end

    test "creates a QUIZ challenge" do
      project = create_project()

      assert {:ok, challenge} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "QUIZ",
                 name: "Quiz Challenge",
                 button_text: "Start Quiz"
               })

      assert challenge.challenge_type == "QUIZ"
    end

    test "creates a PLUGIN challenge with plugin_challenge_id" do
      project = create_project()

      assert {:ok, challenge} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "PLUGIN",
                 name: "Plugin Challenge",
                 plugin_challenge_id: "plugin-123"
               })

      assert challenge.challenge_type == "PLUGIN"
      assert challenge.plugin_challenge_id == "plugin-123"
    end

    test "defaults published_at to now" do
      project = create_project()
      before = DateTime.utc_now() |> DateTime.add(-1)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Auto Published",
          button_text: "Go"
        })

      assert challenge.published_at != nil
      assert DateTime.compare(challenge.published_at, before) == :gt
    end

    test "allows explicit published_at" do
      project = create_project()
      future = ~U[2027-01-01 00:00:00Z]

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Future Published",
          button_text: "Go",
          published_at: future
        })

      assert challenge.published_at == future
    end

    # Type-specific validation

    test "fails when EXTERNAL challenge missing url" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "EXTERNAL",
                 name: "External Challenge",
                 button_text: "Open"
               })

      assert %{url: _} = errors_on(changeset)
    end

    test "fails when SIMPLE challenge has url" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "SIMPLE",
                 name: "Simple Challenge",
                 button_text: "Do it",
                 url: "https://nope.com"
               })

      assert %{url: _} = errors_on(changeset)
    end

    test "fails when PLUGIN challenge missing plugin_challenge_id" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "PLUGIN",
                 name: "Plugin Challenge"
               })

      assert %{plugin_challenge_id: _} = errors_on(changeset)
    end

    test "fails when QUIZ challenge has url" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "QUIZ",
                 name: "Quiz",
                 button_text: "Start",
                 url: "https://nope.com"
               })

      assert %{url: _} = errors_on(changeset)
    end

    test "fails when SIMPLE challenge has plugin_challenge_id" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "SIMPLE",
                 name: "Simple",
                 button_text: "Do",
                 plugin_challenge_id: "bad"
               })

      assert %{plugin_challenge_id: _} = errors_on(changeset)
    end

    test "fails when EXTERNAL challenge has allow_self_completion" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "EXTERNAL",
                 name: "External",
                 button_text: "Go",
                 url: "https://example.com",
                 allow_self_completion: true
               })

      assert %{allow_self_completion: _} = errors_on(changeset)
    end

    test "fails with missing required name" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "SIMPLE",
                 button_text: "Do"
               })

      assert %{name: _} = errors_on(changeset)
    end

    test "fails with invalid challenge type" do
      project = create_project()

      assert {:error, changeset} =
               Challenges.create_challenge(%{
                 project_id: project.id,
                 challenge_type: "INVALID",
                 name: "Bad",
                 button_text: "Do"
               })

      assert %{challenge_type: _} = errors_on(changeset)
    end
  end

  # ── Update Challenge ──

  describe "update_challenge/2" do
    test "updates challenge name" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, updated} = Challenges.update_challenge(challenge, %{name: "Updated Name"})
      assert updated.name == "Updated Name"
    end

    test "updates challenge description" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, updated} =
               Challenges.update_challenge(challenge, %{description: "New description"})

      assert updated.description == "New description"
    end

    test "rejects url on SIMPLE challenge update" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:error, changeset} =
               Challenges.update_challenge(challenge, %{url: "https://bad.com"})

      assert %{url: _} = errors_on(changeset)
    end

    test "rejects allow_self_completion on EXTERNAL challenge update" do
      project = create_project()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "EXTERNAL",
          name: "External",
          button_text: "Go",
          url: "https://example.com"
        })

      assert {:error, changeset} =
               Challenges.update_challenge(challenge, %{allow_self_completion: true})

      assert %{allow_self_completion: _} = errors_on(changeset)
    end

    test "allows updating url on EXTERNAL challenge" do
      project = create_project()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "EXTERNAL",
          name: "External",
          button_text: "Go",
          url: "https://example.com"
        })

      assert {:ok, updated} =
               Challenges.update_challenge(challenge, %{url: "https://new-url.com"})

      assert updated.url == "https://new-url.com"
    end

    test "update by id" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, updated} = Challenges.update_challenge(challenge.id, %{name: "By ID"})
      assert updated.name == "By ID"
    end

    test "update by id returns error for missing challenge" do
      assert {:error, :not_found} =
               Challenges.update_challenge("CL00000000000000000000000000", %{name: "X"})
    end
  end

  # ── Delete Challenge ──

  describe "delete_challenge/1" do
    test "deletes a challenge" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, _} = Challenges.delete_challenge(challenge)
      assert {:error, :not_found} = Challenges.get_challenge(challenge.id)
    end

    test "delete by id" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, _} = Challenges.delete_challenge(challenge.id)
      assert {:error, :not_found} = Challenges.get_challenge(challenge.id)
    end

    test "delete nonexistent returns error" do
      assert {:error, :not_found} = Challenges.delete_challenge("CL00000000000000000000000000")
    end
  end

  # ── Publish Challenge ──

  describe "publish_challenge/2" do
    test "sets published_at" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)
      publish_time = ~U[2026-06-15 12:00:00Z]

      assert {:ok, published} = Challenges.publish_challenge(challenge.id, publish_time)
      assert published.published_at == publish_time
    end

    test "publish nonexistent returns error" do
      assert {:error, :not_found} =
               Challenges.publish_challenge(
                 "CL00000000000000000000000000",
                 ~U[2026-01-01 00:00:00Z]
               )
    end
  end

  # ── Assign to Event ──

  describe "assign_challenge_to_event/2" do
    test "assigns challenge to event" do
      project = create_project()
      event = create_event(project)
      {:ok, challenge} = create_simple_challenge(project)

      assert challenge.event_id == nil

      assert {:ok, updated} = Challenges.assign_challenge_to_event(challenge.id, event.id)
      assert updated.event_id == event.id
    end

    test "reassigns challenge to different event" do
      project = create_project()
      event1 = create_event(project, %{name: "Event 1"})
      event2 = create_event(project, %{name: "Event 2"})

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          event_id: event1.id,
          challenge_type: "SIMPLE",
          name: "Reassign",
          button_text: "Do"
        })

      assert {:ok, updated} = Challenges.assign_challenge_to_event(challenge.id, event2.id)
      assert updated.event_id == event2.id
    end
  end

  # ── Set Visibility ──

  describe "set_challenge_visibility/3" do
    test "sets visible_at" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)
      visible_at = ~U[2026-06-01 00:00:00Z]

      assert {:ok, updated} = Challenges.set_challenge_visibility(challenge.id, visible_at)
      assert updated.visible_at == visible_at
      assert updated.started_at == nil
    end

    test "sets visible_at and started_at" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)
      visible_at = ~U[2026-06-01 00:00:00Z]
      started_at = ~U[2026-06-01 10:00:00Z]

      assert {:ok, updated} =
               Challenges.set_challenge_visibility(challenge.id, visible_at, started_at)

      assert updated.visible_at == visible_at
      assert updated.started_at == started_at
    end
  end

  # ── Set Requirements ──

  describe "set_challenge_requirements/2" do
    test "sets requires_team_membership" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, updated} =
               Challenges.set_challenge_requirements(challenge.id,
                 requires_team_membership: true
               )

      assert updated.requires_team_membership == true
    end

    test "sets requires_super_team_membership" do
      project = create_project()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, updated} =
               Challenges.set_challenge_requirements(challenge.id,
                 requires_super_team_membership: true
               )

      assert updated.requires_super_team_membership == true
    end
  end

  # ── List Challenges with Filters ──

  describe "list_challenges/2" do
    test "returns paginated challenges" do
      project = create_project()

      for i <- 1..5 do
        create_simple_challenge(project, %{name: "Challenge #{i}"})
      end

      result = Challenges.list_challenges(%{project_id: project.id}, %{first: 3})
      assert length(result.edges) == 3
      assert result.total_count == 5
      assert result.page_info.has_next_page == true
      assert result.page_info.has_previous_page == false
    end

    test "forward pagination with after cursor" do
      project = create_project()

      for i <- 1..5 do
        create_simple_challenge(project, %{name: "Challenge #{i}"})
      end

      # Get first page
      page1 = Challenges.list_challenges(%{project_id: project.id}, %{first: 2})
      assert length(page1.edges) == 2

      # Get second page using cursor
      end_cursor = page1.page_info.end_cursor

      page2 =
        Challenges.list_challenges(%{project_id: project.id}, %{first: 2, after: end_cursor})

      assert length(page2.edges) == 2
      # Ensure no overlap
      page1_ids = Enum.map(page1.edges, & &1.node.id)
      page2_ids = Enum.map(page2.edges, & &1.node.id)
      assert MapSet.disjoint?(MapSet.new(page1_ids), MapSet.new(page2_ids))
    end

    test "default pagination when no first/last specified" do
      project = create_project()

      for i <- 1..15 do
        create_simple_challenge(project, %{name: "Challenge #{i}"})
      end

      result = Challenges.list_challenges(%{project_id: project.id}, %{})
      # Default page size is 10
      assert length(result.edges) == 10
      assert result.total_count == 15
    end

    test "filters by project_id" do
      project1 = create_project(%{name: "Project 1"})
      project2 = create_project(%{name: "Project 2"})
      create_simple_challenge(project1, %{name: "P1 Challenge"})
      create_simple_challenge(project2, %{name: "P2 Challenge"})

      result = Challenges.list_challenges(%{project_id: project1.id}, %{first: 10})
      assert result.total_count == 1
      assert hd(result.edges).node.project_id == project1.id
    end

    test "filters by event_id" do
      project = create_project()
      event = create_event(project)

      Challenges.create_challenge(%{
        project_id: project.id,
        event_id: event.id,
        challenge_type: "SIMPLE",
        name: "Event Challenge",
        button_text: "Do"
      })

      create_simple_challenge(project, %{name: "No Event"})

      result = Challenges.list_challenges(%{event_id: event.id}, %{first: 10})
      assert result.total_count == 1
      assert hd(result.edges).node.event_id == event.id
    end

    test "filters by challenge_type" do
      project = create_project()
      create_simple_challenge(project)

      Challenges.create_challenge(%{
        project_id: project.id,
        challenge_type: "EXTERNAL",
        name: "External One",
        button_text: "Go",
        url: "https://example.com"
      })

      result = Challenges.list_challenges(%{challenge_type: "SIMPLE"}, %{first: 10})
      assert result.total_count == 1
      assert hd(result.edges).node.challenge_type == "SIMPLE"
    end

    test "filters by ids" do
      project = create_project()
      {:ok, c1} = create_simple_challenge(project, %{name: "C1"})
      {:ok, c2} = create_simple_challenge(project, %{name: "C2"})
      create_simple_challenge(project, %{name: "C3"})

      result = Challenges.list_challenges(%{ids: [c1.id, c2.id]}, %{first: 10})
      assert result.total_count == 2
      ids = Enum.map(result.edges, & &1.node.id)
      assert c1.id in ids
      assert c2.id in ids
    end

    test "filters by published_after" do
      project = create_project()
      create_simple_challenge(project, %{name: "Old", published_at: ~U[2025-01-01 00:00:00Z]})
      create_simple_challenge(project, %{name: "New", published_at: ~U[2026-06-01 00:00:00Z]})

      result =
        Challenges.list_challenges(%{published_after: ~U[2026-01-01 00:00:00Z]}, %{first: 10})

      assert result.total_count == 1
      assert hd(result.edges).node.name == "New"
    end

    test "filters by published_before" do
      project = create_project()
      create_simple_challenge(project, %{name: "Old", published_at: ~U[2025-01-01 00:00:00Z]})
      create_simple_challenge(project, %{name: "New", published_at: ~U[2026-06-01 00:00:00Z]})

      result =
        Challenges.list_challenges(%{published_before: ~U[2026-01-01 00:00:00Z]}, %{first: 10})

      assert result.total_count == 1
      assert hd(result.edges).node.name == "Old"
    end

    test "combined project_id and event_id filter" do
      project = create_project()
      event = create_event(project)

      Challenges.create_challenge(%{
        project_id: project.id,
        event_id: event.id,
        challenge_type: "SIMPLE",
        name: "Both",
        button_text: "Do"
      })

      create_simple_challenge(project, %{name: "Project Only"})

      result =
        Challenges.list_challenges(%{project_id: project.id, event_id: event.id}, %{first: 10})

      assert result.total_count == 1
      assert hd(result.edges).node.name == "Both"
    end

    test "empty filter returns all" do
      project = create_project()
      create_simple_challenge(project, %{name: "C1"})
      create_simple_challenge(project, %{name: "C2"})

      result = Challenges.list_challenges(%{}, %{first: 10})
      assert result.total_count == 2
    end

    test "nil filter returns all" do
      project = create_project()
      create_simple_challenge(project)

      result = Challenges.list_challenges(nil, %{first: 10})
      assert result.total_count == 1
    end
  end

  # ── Visibility ──

  describe "visibility" do
    test "get_visible_challenge returns challenge with past visible_at" do
      project = create_project()
      past = DateTime.add(DateTime.utc_now(), -3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Visible",
          button_text: "Do",
          visible_at: past
        })

      assert {:ok, _} = Challenges.get_visible_challenge(challenge.id)
    end

    test "get_visible_challenge hides challenge with future visible_at from unenrolled user" do
      project = create_project()
      user = create_user()
      future = DateTime.add(DateTime.utc_now(), 3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Future Visible",
          button_text: "Do",
          visible_at: future
        })

      assert {:error, :not_found} =
               Challenges.get_visible_challenge(challenge.id, user_id: user.id)
    end

    test "get_visible_challenge hides challenge with NULL visible_at from unenrolled user" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Null Visible",
          button_text: "Do",
          visible_at: nil,
          published_at: DateTime.add(DateTime.utc_now(), 3600)
        })

      assert {:error, :not_found} =
               Challenges.get_visible_challenge(challenge.id, user_id: user.id)
    end

    test "enrolled user CAN see challenge with future visible_at" do
      project = create_project()
      user = create_user()
      future = DateTime.add(DateTime.utc_now(), 3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Future Visible",
          button_text: "Do",
          visible_at: future
        })

      # Enroll user
      Challenges.enroll_in_challenge(user.id, challenge.id)

      assert {:ok, _} =
               Challenges.get_visible_challenge(challenge.id, user_id: user.id)
    end

    test "enrolled user CAN see challenge with NULL visible_at" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Null Visible",
          button_text: "Do",
          visible_at: nil
        })

      Challenges.enroll_in_challenge(user.id, challenge.id)

      assert {:ok, _} =
               Challenges.get_visible_challenge(challenge.id, user_id: user.id)
    end

    test "user CAN see challenge with past visible_at regardless of enrollment" do
      project = create_project()
      user = create_user()
      past = DateTime.add(DateTime.utc_now(), -3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Past Visible",
          button_text: "Do",
          visible_at: past
        })

      assert {:ok, _} =
               Challenges.get_visible_challenge(challenge.id, user_id: user.id)
    end

    test "admin can see ALL challenges regardless of visibility" do
      project = create_project()
      future = DateTime.add(DateTime.utc_now(), 3600)

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Admin Sees All",
          button_text: "Do",
          visible_at: future,
          published_at: future
        })

      assert {:ok, _} =
               Challenges.get_visible_challenge(challenge.id, roles: ["admin"])
    end

    test "list_visible_challenges excludes non-visible for unenrolled user" do
      project = create_project()
      future = DateTime.add(DateTime.utc_now(), 3600)
      past = DateTime.add(DateTime.utc_now(), -3600)

      create_simple_challenge(project, %{name: "Visible", visible_at: past})

      Challenges.create_challenge(%{
        project_id: project.id,
        challenge_type: "SIMPLE",
        name: "Hidden",
        button_text: "Do",
        visible_at: future
      })

      result =
        Challenges.list_visible_challenges(
          %{project_id: project.id},
          %{first: 10},
          roles: ["user"]
        )

      assert result.total_count == 1
      assert hd(result.edges).node.name == "Visible"
    end

    test "list_visible_challenges includes enrolled-but-not-visible challenges" do
      project = create_project()
      user = create_user()
      future = DateTime.add(DateTime.utc_now(), 3600)

      {:ok, hidden} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Hidden But Enrolled",
          button_text: "Do",
          visible_at: future
        })

      Challenges.enroll_in_challenge(user.id, hidden.id)

      result =
        Challenges.list_visible_challenges(
          %{project_id: project.id},
          %{first: 10},
          user_id: user.id,
          roles: ["user"]
        )

      assert result.total_count == 1
      assert hd(result.edges).node.name == "Hidden But Enrolled"
    end

    test "admin list_visible_challenges sees all" do
      project = create_project()
      future = DateTime.add(DateTime.utc_now(), 3600)
      past = DateTime.add(DateTime.utc_now(), -3600)

      create_simple_challenge(project, %{name: "Visible", visible_at: past})

      Challenges.create_challenge(%{
        project_id: project.id,
        challenge_type: "SIMPLE",
        name: "Hidden",
        button_text: "Do",
        visible_at: future
      })

      result =
        Challenges.list_visible_challenges(
          %{project_id: project.id},
          %{first: 10},
          roles: ["admin"]
        )

      assert result.total_count == 2
    end
  end

  # ── Completions ──

  describe "completions" do
    test "complete and uncomplete a challenge" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, _completion} = Challenges.complete_challenge(user.id, challenge.id)
      assert {:ok, %DateTime{}} = Challenges.get_user_completed_at(user.id, challenge.id)

      assert {:ok, true} = Challenges.uncomplete_challenge(user.id, challenge.id)
      assert {:ok, nil} = Challenges.get_user_completed_at(user.id, challenge.id)
    end

    test "complete is idempotent" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, _} = Challenges.complete_challenge(user.id, challenge.id)
      assert {:ok, _} = Challenges.complete_challenge(user.id, challenge.id)
    end

    test "admin_complete_challenge with explicit timestamp" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)
      ts = ~U[2026-06-15 12:00:00Z]

      assert {:ok, _} = Challenges.admin_complete_challenge(user.id, challenge.id, ts)
      assert {:ok, ^ts} = Challenges.get_user_completed_at(user.id, challenge.id)
    end

    test "different users have separate completions" do
      project = create_project()
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})
      {:ok, challenge} = create_simple_challenge(project)

      Challenges.complete_challenge(user1.id, challenge.id)

      assert {:ok, %DateTime{}} = Challenges.get_user_completed_at(user1.id, challenge.id)
      assert {:ok, nil} = Challenges.get_user_completed_at(user2.id, challenge.id)
    end

    test "uncomplete nonexistent returns false" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, false} = Challenges.uncomplete_challenge(user.id, challenge.id)
    end
  end

  # ── Enrollments ──

  describe "enrollments" do
    test "enroll and unenroll" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, _challenge} = Challenges.enroll_in_challenge(user.id, challenge.id)
      assert {:ok, %DateTime{}} = Challenges.get_user_enrolled_at(user.id, challenge.id)

      assert {:ok, true} = Challenges.unenroll_from_challenge(user.id, challenge.id)
      assert {:ok, nil} = Challenges.get_user_enrolled_at(user.id, challenge.id)
    end

    test "enrollment is idempotent" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, _} = Challenges.enroll_in_challenge(user.id, challenge.id)
      assert {:ok, _} = Challenges.enroll_in_challenge(user.id, challenge.id)
    end

    test "rejects enrollment for unpublished challenge" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Unpublished",
          button_text: "Do it",
          published_at: DateTime.add(DateTime.utc_now(), 3600)
        })

      assert {:error, "challenge is not yet published"} =
               Challenges.enroll_in_challenge(user.id, challenge.id)
    end

    test "rejects enrollment for ended challenge" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Ended",
          button_text: "Do it",
          published_at: DateTime.add(DateTime.utc_now(), -7200),
          end_time: DateTime.add(DateTime.utc_now(), -3600)
        })

      assert {:error, "challenge has ended"} =
               Challenges.enroll_in_challenge(user.id, challenge.id)
    end

    test "different users have separate enrollments" do
      project = create_project()
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})
      {:ok, challenge} = create_simple_challenge(project)

      Challenges.enroll_in_challenge(user1.id, challenge.id)

      assert {:ok, %DateTime{}} = Challenges.get_user_enrolled_at(user1.id, challenge.id)
      assert {:ok, nil} = Challenges.get_user_enrolled_at(user2.id, challenge.id)
    end

    test "unenroll nonexistent returns false" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, false} = Challenges.unenroll_from_challenge(user.id, challenge.id)
    end
  end

  # ── Self-Completion ──

  describe "self_complete_challenge/2" do
    test "allows self-completion for simple challenge" do
      project = create_project()
      user = create_user()
      {:ok, challenge} = create_simple_challenge(project)

      assert {:ok, %Challenge{}} = Challenges.self_complete_challenge(user.id, challenge.id)
    end

    test "rejects self-completion when not allowed" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "No Self",
          button_text: "Do it",
          allow_self_completion: false
        })

      assert {:error, "self-completion not allowed for this challenge"} =
               Challenges.self_complete_challenge(user.id, challenge.id)
    end

    test "rejects self-completion for external challenge" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "EXTERNAL",
          name: "External",
          button_text: "Go",
          url: "https://example.com"
        })

      assert {:error, "self-completion is only available for simple challenges"} =
               Challenges.self_complete_challenge(user.id, challenge.id)
    end

    test "rejects self-completion for quiz challenge" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "QUIZ",
          name: "Quiz",
          button_text: "Start"
        })

      assert {:error, "self-completion is only available for simple challenges"} =
               Challenges.self_complete_challenge(user.id, challenge.id)
    end

    test "rejects self-completion for unpublished challenge" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Future",
          button_text: "Do",
          allow_self_completion: true,
          published_at: DateTime.add(DateTime.utc_now(), 3600)
        })

      assert {:error, "challenge is not yet published"} =
               Challenges.self_complete_challenge(user.id, challenge.id)
    end

    test "rejects self-completion for ended challenge" do
      project = create_project()
      user = create_user()

      {:ok, challenge} =
        Challenges.create_challenge(%{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Ended",
          button_text: "Do",
          allow_self_completion: true,
          published_at: DateTime.add(DateTime.utc_now(), -7200),
          end_time: DateTime.add(DateTime.utc_now(), -3600)
        })

      assert {:error, "challenge has ended"} =
               Challenges.self_complete_challenge(user.id, challenge.id)
    end
  end

  # ── Cursor Pagination Encoding ──

  describe "cursor pagination encoding" do
    test "encode and decode round-trips" do
      challenge = %{published_at: ~U[2026-03-15 10:00:00Z], id: "CL01234567890123456789ABCDEF"}

      cursor = Pagination.encode_cursor(challenge)
      assert {:ok, decoded} = Pagination.decode_cursor(cursor)
      assert decoded.sort_key == ~U[2026-03-15 10:00:00Z]
      assert decoded.id == "CL01234567890123456789ABCDEF"
    end

    test "encode with nil published_at" do
      challenge = %{published_at: nil, id: "CL01234567890123456789ABCDEF"}
      cursor = Pagination.encode_cursor(challenge)
      assert {:ok, decoded} = Pagination.decode_cursor(cursor)
      assert decoded.sort_key == nil
      assert decoded.id == "CL01234567890123456789ABCDEF"
    end

    test "decode returns error for invalid cursor" do
      assert {:error, :invalid_cursor} = Pagination.decode_cursor("not-valid")
    end

    test "decode nil returns nil" do
      assert Pagination.decode_cursor(nil) == nil
    end
  end

  # ── Helpers ──

  defp create_simple_challenge(project, extra_attrs \\ %{}) do
    attrs =
      Map.merge(
        %{
          project_id: project.id,
          challenge_type: "SIMPLE",
          name: "Simple Challenge",
          button_text: "Complete"
        },
        extra_attrs
      )

    Challenges.create_challenge(attrs)
  end
end
