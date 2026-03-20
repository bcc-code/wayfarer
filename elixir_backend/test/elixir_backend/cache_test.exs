defmodule ElixirBackend.CacheTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Cache
  alias ElixirBackend.Challenges

  import ElixirBackend.TestHelpers

  defp create_test_challenge(project, attrs \\ %{}) do
    defaults = %{
      name: "Cache Test Challenge",
      challenge_type: "SIMPLE",
      button_text: "Complete",
      project_id: project.id,
      published_at: DateTime.utc_now() |> DateTime.add(-3600) |> DateTime.truncate(:second)
    }

    {:ok, challenge} = Challenges.create_challenge(Map.merge(defaults, attrs))
    challenge
  end

  # ── Challenge cache ──
  # Mirrors: challenge_by_id_test.go

  describe "challenge cache — key" do
    test "cache key contains challenge ID" do
      challenge_id = "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
      cache_key = Cache.challenge_key(challenge_id)

      assert cache_key != ""
      assert String.contains?(cache_key, challenge_id)
    end
  end

  describe "challenge cache — behavior" do
    setup do
      project = create_project()
      %{project: project}
    end

    test "caches full challenge struct with all fields", %{project: project} do
      event = create_event(project)
      end_time = DateTime.utc_now() |> DateTime.add(7 * 86_400) |> DateTime.truncate(:second)

      challenge =
        create_test_challenge(project, %{
          name: "Test Challenge",
          description: "Test description",
          image_url: "https://example.com/image.png",
          event_id: event.id,
          button_text: "Start Challenge",
          end_time: end_time
        })

      {:ok, cached} = Cache.get_challenge(challenge.id)

      assert cached.id == challenge.id
      assert cached.name == "Test Challenge"
      assert cached.description == "Test description"
      assert cached.image_url == "https://example.com/image.png"
      assert cached.button_text == "Start Challenge"
      assert cached.end_time == end_time
    end

    test "get_challenge returns cached value (cache hit)", %{project: project} do
      challenge = create_test_challenge(project)

      {:ok, _} = Challenges.get_challenge(challenge.id)

      # Delete from DB to prove cache is being used
      Repo.delete!(challenge)

      {:ok, cached} = Challenges.get_challenge(challenge.id)
      assert cached.id == challenge.id
    end

    test "cache miss returns :miss" do
      assert :miss = Cache.get_challenge("nonexistent")
    end
  end

  describe "challenge cache — expiry/invalidation" do
    setup do
      project = create_project()
      %{project: project}
    end

    test "delete_challenge removes from cache", %{project: project} do
      challenge = create_test_challenge(project)

      {:ok, _} = Challenges.get_challenge(challenge.id)
      assert {:ok, _} = Cache.get_challenge(challenge.id)

      {:ok, _} = Challenges.delete_challenge(challenge.id)
      assert :miss = Cache.get_challenge(challenge.id)
    end

    test "update_challenge updates the cache", %{project: project} do
      challenge = create_test_challenge(project)

      {:ok, _} = Challenges.get_challenge(challenge.id)
      {:ok, _} = Challenges.update_challenge(challenge.id, %{name: "Updated Name"})

      {:ok, cached} = Cache.get_challenge(challenge.id)
      assert cached.name == "Updated Name"
    end

    test "publish_challenge updates the cache", %{project: project} do
      challenge = create_test_challenge(project, %{published_at: nil})
      new_time = DateTime.utc_now() |> DateTime.truncate(:second)

      {:ok, _} = Challenges.publish_challenge(challenge.id, new_time)

      {:ok, cached} = Cache.get_challenge(challenge.id)
      assert cached.published_at == new_time
    end

    test "set_challenge_visibility updates the cache", %{project: project} do
      challenge = create_test_challenge(project)
      visible_at = DateTime.utc_now() |> DateTime.add(3600) |> DateTime.truncate(:second)

      {:ok, _} = Challenges.set_challenge_visibility(challenge.id, visible_at)

      {:ok, cached} = Cache.get_challenge(challenge.id)
      assert cached.visible_at == visible_at
    end

    test "assign_challenge_to_event updates the cache", %{project: project} do
      challenge = create_test_challenge(project)
      event = create_event(project)

      {:ok, _} = Challenges.assign_challenge_to_event(challenge.id, event.id)

      {:ok, cached} = Cache.get_challenge(challenge.id)
      assert cached.event_id == event.id
    end
  end

  describe "challenge cache — multiple challenges" do
    setup do
      project = create_project()
      %{project: project}
    end

    test "stores and retrieves multiple challenge types independently", %{project: project} do
      external =
        create_test_challenge(project, %{
          name: "Challenge 1",
          challenge_type: "EXTERNAL",
          button_text: "Start",
          url: "https://example.com/challenge1"
        })

      simple =
        create_test_challenge(project, %{
          name: "Challenge 2",
          challenge_type: "SIMPLE",
          button_text: "Begin",
          allow_self_completion: true
        })

      quiz =
        create_test_challenge(project, %{
          name: "Challenge 3",
          challenge_type: "QUIZ",
          button_text: "Go"
        })

      # Verify all challenges can be retrieved from cache
      {:ok, cached1} = Cache.get_challenge(external.id)
      assert cached1.id == external.id
      assert cached1.challenge_type == "EXTERNAL"

      {:ok, cached2} = Cache.get_challenge(simple.id)
      assert cached2.id == simple.id
      assert cached2.challenge_type == "SIMPLE"

      {:ok, cached3} = Cache.get_challenge(quiz.id)
      assert cached3.id == quiz.id
      assert cached3.challenge_type == "QUIZ"
    end
  end

  # ── Completion timestamp cache ──
  # Mirrors: user_challenge_completion_timestamp_test.go

  describe "completion cache — key" do
    test "cache key contains user ID and challenge ID" do
      user_id = "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
      challenge_id = "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
      cache_key = Cache.completion_key(user_id, challenge_id)

      assert cache_key != ""
      assert String.contains?(cache_key, user_id)
      assert String.contains?(cache_key, challenge_id)
      assert cache_key == "#{user_id}:#{challenge_id}"
    end
  end

  describe "completion cache — behavior" do
    setup do
      project = create_project()
      user = create_user()
      challenge = create_test_challenge(project)
      %{project: project, user: user, challenge: challenge}
    end

    test "complete_challenge caches the timestamp", %{user: user, challenge: challenge} do
      {:ok, completion} = Challenges.complete_challenge(user.id, challenge.id)

      {:ok, cached} = Cache.get_completion(user.id, challenge.id)
      assert cached == completion.completed_at
    end

    test "get_user_completed_at populates cache on miss", %{user: user, challenge: challenge} do
      {:ok, _} = Challenges.complete_challenge(user.id, challenge.id)

      # Clear cache to simulate cold start
      Cache.delete_completion(user.id, challenge.id)

      {:ok, timestamp} = Challenges.get_user_completed_at(user.id, challenge.id)
      assert timestamp != nil

      {:ok, cached} = Cache.get_completion(user.id, challenge.id)
      assert cached == timestamp
    end
  end

  describe "completion cache — nil caching" do
    setup do
      project = create_project()
      user = create_user()
      challenge = create_test_challenge(project)
      %{user: user, challenge: challenge}
    end

    test "caches nil for uncompleted challenge", %{user: user, challenge: challenge} do
      {:ok, nil} = Challenges.get_user_completed_at(user.id, challenge.id)

      # Should cache nil (as :cached_nil sentinel)
      {:ok, nil} = Cache.get_completion(user.id, challenge.id)
    end
  end

  describe "completion cache — invalidation" do
    setup do
      project = create_project()
      user = create_user()
      challenge = create_test_challenge(project)
      %{user: user, challenge: challenge}
    end

    test "uncomplete_challenge clears the cache", %{user: user, challenge: challenge} do
      {:ok, _} = Challenges.complete_challenge(user.id, challenge.id)
      assert {:ok, _} = Cache.get_completion(user.id, challenge.id)

      {:ok, true} = Challenges.uncomplete_challenge(user.id, challenge.id)
      assert :miss = Cache.get_completion(user.id, challenge.id)
    end
  end

  describe "completion cache — multiple challenges per user" do
    setup do
      project = create_project()
      user = create_user()
      %{project: project, user: user}
    end

    test "stores timestamps for multiple challenges including nil", %{
      project: project,
      user: user
    } do
      ch1 = create_test_challenge(project, %{name: "Ch 1"})
      ch2 = create_test_challenge(project, %{name: "Ch 2"})
      ch3 = create_test_challenge(project, %{name: "Ch 3"})

      # Complete ch1 and ch2, leave ch3 uncompleted
      {:ok, comp1} = Challenges.complete_challenge(user.id, ch1.id)
      {:ok, comp2} = Challenges.complete_challenge(user.id, ch2.id)
      {:ok, nil} = Challenges.get_user_completed_at(user.id, ch3.id)

      # Verify all timestamps can be retrieved from cache
      {:ok, cached1} = Cache.get_completion(user.id, ch1.id)
      assert cached1 == comp1.completed_at

      {:ok, cached2} = Cache.get_completion(user.id, ch2.id)
      assert cached2 == comp2.completed_at

      {:ok, cached3} = Cache.get_completion(user.id, ch3.id)
      assert cached3 == nil
    end
  end

  describe "completion cache — user isolation" do
    setup do
      project = create_project()
      challenge = create_test_challenge(project)
      %{challenge: challenge}
    end

    test "different users have separate cache entries", %{challenge: challenge} do
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})

      {:ok, comp1} = Challenges.complete_challenge(user1.id, challenge.id)
      {:ok, comp2} = Challenges.complete_challenge(user2.id, challenge.id)

      {:ok, cached1} = Cache.get_completion(user1.id, challenge.id)
      assert cached1 == comp1.completed_at

      {:ok, cached2} = Cache.get_completion(user2.id, challenge.id)
      assert cached2 == comp2.completed_at
    end

    test "invalidating one user does not affect another", %{challenge: challenge} do
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})

      {:ok, _} = Challenges.complete_challenge(user1.id, challenge.id)
      {:ok, _} = Challenges.complete_challenge(user2.id, challenge.id)

      # Invalidate user1
      Cache.delete_completion(user1.id, challenge.id)

      assert :miss = Cache.get_completion(user1.id, challenge.id)
      assert {:ok, _} = Cache.get_completion(user2.id, challenge.id)
    end
  end

  # ── Enrollment timestamp cache ──
  # Mirrors: user_challenge_enrollment_timestamp_test.go

  describe "enrollment cache — key" do
    test "cache key contains user ID and challenge ID" do
      user_id = "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
      challenge_id = "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
      cache_key = Cache.enrollment_key(user_id, challenge_id)

      assert cache_key != ""
      assert String.contains?(cache_key, user_id)
      assert String.contains?(cache_key, challenge_id)
      assert cache_key == "#{user_id}:#{challenge_id}"
    end
  end

  describe "enrollment cache — behavior" do
    setup do
      project = create_project()
      user = create_user()

      challenge =
        create_test_challenge(project, %{
          published_at: DateTime.utc_now() |> DateTime.add(-3600) |> DateTime.truncate(:second)
        })

      %{project: project, user: user, challenge: challenge}
    end

    test "enroll_in_challenge caches the timestamp", %{user: user, challenge: challenge} do
      {:ok, _} = Challenges.enroll_in_challenge(user.id, challenge.id)

      {:ok, cached} = Cache.get_enrollment(user.id, challenge.id)
      assert cached != nil
    end

    test "get_user_enrolled_at populates cache on miss", %{user: user, challenge: challenge} do
      {:ok, _} = Challenges.enroll_in_challenge(user.id, challenge.id)

      # Clear cache to simulate cold start
      Cache.delete_enrollment(user.id, challenge.id)

      {:ok, timestamp} = Challenges.get_user_enrolled_at(user.id, challenge.id)
      assert timestamp != nil

      {:ok, cached} = Cache.get_enrollment(user.id, challenge.id)
      assert cached == timestamp
    end
  end

  describe "enrollment cache — nil caching" do
    setup do
      project = create_project()
      user = create_user()
      challenge = create_test_challenge(project)
      %{user: user, challenge: challenge}
    end

    test "caches nil for unenrolled user", %{user: user, challenge: challenge} do
      {:ok, nil} = Challenges.get_user_enrolled_at(user.id, challenge.id)

      {:ok, nil} = Cache.get_enrollment(user.id, challenge.id)
    end
  end

  describe "enrollment cache — invalidation" do
    setup do
      project = create_project()
      user = create_user()

      challenge =
        create_test_challenge(project, %{
          published_at: DateTime.utc_now() |> DateTime.add(-3600) |> DateTime.truncate(:second)
        })

      %{user: user, challenge: challenge}
    end

    test "unenroll_from_challenge clears the cache", %{user: user, challenge: challenge} do
      {:ok, _} = Challenges.enroll_in_challenge(user.id, challenge.id)
      assert {:ok, _} = Cache.get_enrollment(user.id, challenge.id)

      {:ok, true} = Challenges.unenroll_from_challenge(user.id, challenge.id)
      assert :miss = Cache.get_enrollment(user.id, challenge.id)
    end
  end

  describe "enrollment cache — multiple challenges per user" do
    setup do
      project = create_project()
      user = create_user()
      %{project: project, user: user}
    end

    test "stores timestamps for multiple challenges including nil", %{
      project: project,
      user: user
    } do
      ch1 = create_test_challenge(project, %{name: "Ch 1"})
      ch2 = create_test_challenge(project, %{name: "Ch 2"})
      ch3 = create_test_challenge(project, %{name: "Ch 3"})

      # Enroll in ch1 and ch2, leave ch3 unenrolled
      {:ok, _} = Challenges.enroll_in_challenge(user.id, ch1.id)
      {:ok, _} = Challenges.enroll_in_challenge(user.id, ch2.id)
      {:ok, nil} = Challenges.get_user_enrolled_at(user.id, ch3.id)

      # Verify all timestamps can be retrieved from cache
      {:ok, cached1} = Cache.get_enrollment(user.id, ch1.id)
      assert cached1 != nil

      {:ok, cached2} = Cache.get_enrollment(user.id, ch2.id)
      assert cached2 != nil

      {:ok, cached3} = Cache.get_enrollment(user.id, ch3.id)
      assert cached3 == nil
    end
  end

  describe "enrollment cache — user isolation" do
    setup do
      project = create_project()

      challenge =
        create_test_challenge(project, %{
          published_at: DateTime.utc_now() |> DateTime.add(-3600) |> DateTime.truncate(:second)
        })

      %{challenge: challenge}
    end

    test "different users have separate cache entries", %{challenge: challenge} do
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})

      {:ok, _} = Challenges.enroll_in_challenge(user1.id, challenge.id)
      {:ok, _} = Challenges.enroll_in_challenge(user2.id, challenge.id)

      {:ok, cached1} = Cache.get_enrollment(user1.id, challenge.id)
      assert cached1 != nil

      {:ok, cached2} = Cache.get_enrollment(user2.id, challenge.id)
      assert cached2 != nil
    end

    test "invalidating one user does not affect another", %{challenge: challenge} do
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})

      {:ok, _} = Challenges.enroll_in_challenge(user1.id, challenge.id)
      {:ok, _} = Challenges.enroll_in_challenge(user2.id, challenge.id)

      # Invalidate user1
      Cache.delete_enrollment(user1.id, challenge.id)

      assert :miss = Cache.get_enrollment(user1.id, challenge.id)
      assert {:ok, _} = Cache.get_enrollment(user2.id, challenge.id)
    end
  end

  # ── User cache ──

  describe "user cache — key" do
    test "cache key contains user ID" do
      user_id = "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
      cache_key = Cache.user_key(user_id)

      assert cache_key != ""
      assert String.contains?(cache_key, user_id)
    end
  end

  describe "user cache — behavior" do
    test "put and get user" do
      user = create_user(%{name: "Cache User"})
      Cache.put_user(user)

      assert {:ok, cached} = Cache.get_user(user.id)
      assert cached.id == user.id
      assert cached.name == "Cache User"
    end

    test "cache miss returns :miss" do
      assert :miss = Cache.get_user("nonexistent")
    end

    test "delete removes from cache" do
      user = create_user(%{name: "Delete User"})
      Cache.put_user(user)

      assert {:ok, _} = Cache.get_user(user.id)

      Cache.delete_user(user.id)
      assert :miss = Cache.get_user(user.id)
    end

    test "fetch_user populates cache on miss" do
      user = create_user(%{name: "Fetch User"})
      # Ensure not in cache
      Cache.delete_user(user.id)

      {:ok, found} =
        Cache.fetch_user(user.id, fn ->
          case Repo.get(ElixirBackend.Accounts.User, user.id) do
            nil -> {:error, :not_found}
            u -> {:ok, u}
          end
        end)

      assert found.id == user.id

      # Should now be in cache
      assert {:ok, cached} = Cache.get_user(user.id)
      assert cached.id == user.id
    end
  end

  describe "user cache — invalidation on mutations" do
    test "lock_user_church updates cache" do
      user = create_user()

      # Prime the cache
      {:ok, _} = ElixirBackend.Accounts.get_user(user.id)
      assert {:ok, cached} = Cache.get_user(user.id)
      assert cached.church_locked_until == nil

      {:ok, _} = ElixirBackend.Accounts.lock_user_church(user.id)

      {:ok, cached} = Cache.get_user(user.id)
      assert cached.church_locked_until != nil
    end

    test "unlock_user_church updates cache" do
      user = create_user()

      {:ok, _} = ElixirBackend.Accounts.lock_user_church(user.id)
      {:ok, cached} = Cache.get_user(user.id)
      assert cached.church_locked_until != nil

      {:ok, _} = ElixirBackend.Accounts.unlock_user_church(user.id)
      {:ok, cached} = Cache.get_user(user.id)
      assert cached.church_locked_until == nil
    end
  end

  # ── clear_all ──

  describe "clear_all" do
    test "clears all caches" do
      Cache.put_challenge(%{id: "CL1", name: "test"})
      Cache.put_completion("US1", "CL1", ~U[2026-01-01 00:00:00Z])
      Cache.put_enrollment("US1", "CL1", ~U[2026-01-01 00:00:00Z])
      Cache.put_user(%{id: "US1", name: "test"})

      Cache.clear_all()

      assert :miss = Cache.get_challenge("CL1")
      assert :miss = Cache.get_completion("US1", "CL1")
      assert :miss = Cache.get_enrollment("US1", "CL1")
      assert :miss = Cache.get_user("US1")
    end
  end
end
