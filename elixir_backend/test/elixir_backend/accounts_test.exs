defmodule ElixirBackend.AccountsTest do
  use ElixirBackend.DataCase

  alias ElixirBackend.Accounts

  import ElixirBackend.TestHelpers

  # ── validateUserAccess equivalent ──
  # Go: TestValidateUserAccess — tests auth extraction + user loading

  describe "get_user — basic lookups (validateUserAccess equivalent)" do
    test "valid user found by ID" do
      user = create_user(%{name: "Valid User"})

      assert {:ok, found} = Accounts.get_user(user.id)
      assert found.id == user.id
      assert found.name == "Valid User"
    end

    test "user not found returns error" do
      assert {:error, :not_found} = Accounts.get_user("US00000000000000000000000000")
    end

    test "me returns user for valid user_id" do
      user = create_user()
      assert {:ok, found} = Accounts.me(user.id)
      assert found.id == user.id
    end

    test "me returns error for non-existent user_id" do
      assert {:error, :not_found} = Accounts.me("US00000000000000000000000000")
    end
  end

  # ── checkUserPermissions equivalent ──
  # Go: TestCheckUserPermissions — tests permission level determination
  # Note: church_admin, project_admin, team_lead are scoped out (no RoleService).
  # We test the JWT-role-based equivalent: has_elevated_role? behavior.

  describe "get_accessible_user — permission checking (checkUserPermissions equivalent)" do
    test "superadmin has full access to any user" do
      user = create_user()

      assert {:ok, found} =
               Accounts.get_accessible_user(user.id,
                 user_id: "US_OTHER_SA",
                 roles: ["superadmin"]
               )

      assert found.id == user.id
    end

    test "admin has full access to any user" do
      user = create_user()

      assert {:ok, found} =
               Accounts.get_accessible_user(user.id,
                 user_id: "US_OTHER_ADMIN",
                 roles: ["admin"]
               )

      assert found.id == user.id
    end

    test "m2m has full access to any user" do
      user = create_user()

      assert {:ok, found} =
               Accounts.get_accessible_user(user.id,
                 user_id: "US_M2M_SERVICE",
                 roles: ["m2m"]
               )

      assert found.id == user.id
    end

    test "user with multiple elevated roles has access" do
      user = create_user()

      assert {:ok, found} =
               Accounts.get_accessible_user(user.id,
                 user_id: "US_MULTI_ROLE",
                 roles: ["user", "admin", "m2m"]
               )

      assert found.id == user.id
    end

    test "regular user can access themselves" do
      user = create_user()

      assert {:ok, found} =
               Accounts.get_accessible_user(user.id,
                 user_id: user.id,
                 roles: ["user"]
               )

      assert found.id == user.id
    end

    test "regular user cannot access other users" do
      user = create_user()
      other = create_user(%{name: "Other User"})

      assert {:error, :not_found} =
               Accounts.get_accessible_user(other.id,
                 user_id: user.id,
                 roles: ["user"]
               )
    end

    test "user with no roles is denied" do
      user = create_user()
      other = create_user(%{name: "Other"})

      assert {:error, :not_found} =
               Accounts.get_accessible_user(other.id,
                 user_id: user.id,
                 roles: []
               )
    end

    test "unauthenticated (nil user_id) is denied" do
      user = create_user()

      assert {:error, :not_found} =
               Accounts.get_accessible_user(user.id, user_id: nil, roles: [])
    end

    test "accessing non-existent user returns not_found even for admin" do
      assert {:error, :not_found} =
               Accounts.get_accessible_user("US00000000000000000000000000",
                 user_id: "US_ADMIN",
                 roles: ["admin"]
               )
    end
  end

  # ── applyPermissionFilters equivalent ──
  # Go: TestApplyPermissionFilters — tests that roles restrict listing visibility
  # Elixir: admins see all, users denied entirely (no granular church/project/team scoping yet)

  describe "list_users — permission-based access (applyPermissionFilters equivalent)" do
    test "superadmin sees all users — no filter restriction" do
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})

      {:ok, result} = Accounts.list_users(%{}, %{}, roles: ["superadmin"])

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
      assert user2.id in ids
    end

    test "admin sees all users — no filter restriction" do
      user1 = create_user(%{name: "User 1"})
      user2 = create_user(%{name: "User 2"})

      {:ok, result} = Accounts.list_users(%{}, %{}, roles: ["admin"])

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
      assert user2.id in ids
    end

    test "m2m sees all users" do
      _user = create_user()

      {:ok, result} = Accounts.list_users(%{}, %{}, roles: ["m2m"])
      assert result.total_count >= 1
    end

    test "regular user is denied from listing" do
      _user = create_user()

      assert {:error, :unauthorized} =
               Accounts.list_users(%{}, %{}, user_id: "US_REGULAR", roles: ["user"])
    end

    test "user with no roles is denied from listing" do
      assert {:error, :unauthorized} =
               Accounts.list_users(%{}, %{}, user_id: "US_NOROLE", roles: [])
    end

    test "unauthenticated is denied from listing" do
      assert {:error, :unauthorized} =
               Accounts.list_users(%{}, %{}, roles: [])
    end

    test "nil filter is treated as empty — returns all for admin" do
      _user = create_user()

      {:ok, result} = Accounts.list_users(nil, %{}, roles: ["admin"])
      assert result.total_count >= 1
    end
  end

  # ── buildUserFilterParams / buildCountFilterParams equivalent ──
  # Go: TestBuildUserFilterParams (8 tests) + TestBuildCountFilterParams (10 tests)
  # Tests that each filter field correctly restricts results and count.

  describe "list_users — filter params (buildUserFilterParams equivalent)" do
    setup do
      church1 = create_church(%{name: "Church Alpha", country: "NO"})
      church2 = create_church(%{name: "Church Beta", country: "SE"})

      user1 =
        create_user(%{
          name: "Alice Smith",
          email: "alice@example.com",
          gender: "FEMALE",
          birthdate: ~D[2000-06-15],
          church_id: church1.id
        })

      user2 =
        create_user(%{
          name: "Bob Jones",
          email: "bob@example.com",
          gender: "MALE",
          birthdate: ~D[1990-03-20],
          church_id: church1.id
        })

      user3 =
        create_user(%{
          name: "Carol Davis",
          email: "carol@example.com",
          gender: "FEMALE",
          birthdate: ~D[1985-11-01],
          church_id: church2.id
        })

      admin_opts = [roles: ["admin"]]

      %{
        user1: user1,
        user2: user2,
        user3: user3,
        church1: church1,
        church2: church2,
        admin_opts: admin_opts
      }
    end

    test "all filters populated", %{
      user1: user1,
      church1: church1,
      admin_opts: opts
    } do
      project = create_project()
      event = create_event(project)
      create_user_project(user1, project)
      create_user_event(user1, event)

      {:ok, result} =
        Accounts.list_users(
          %{
            church_id: church1.id,
            gender: "FEMALE",
            min_age: 20,
            max_age: 30,
            project_id: project.id,
            event_id: event.id,
            ids: [user1.id],
            query: "Alice"
          },
          %{},
          opts
        )

      assert result.total_count == 1
      assert hd(result.edges).node.id == user1.id
    end

    test "only church filter", %{church1: church1, user1: user1, user2: user2, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{church_id: church1.id}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
      assert user2.id in ids
      assert result.total_count == 2
    end

    test "only gender filter — FEMALE", %{user1: user1, user3: user3, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{gender: "FEMALE"}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
      assert user3.id in ids
      assert Enum.all?(result.edges, fn e -> e.node.gender == "FEMALE" end)
    end

    test "only gender filter — MALE", %{user2: user2, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{gender: "MALE"}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user2.id in ids
      assert Enum.all?(result.edges, fn e -> e.node.gender == "MALE" end)
    end

    test "only age range filter — min and max", %{user2: user2, admin_opts: opts} do
      # user2 born 1990 → ~36 yrs. user1 born 2000 → ~25. user3 born 1985 → ~40.
      {:ok, result} = Accounts.list_users(%{min_age: 33, max_age: 38}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user2.id in ids
      assert result.total_count == 1
    end

    test "only min_age filter", %{user2: user2, user3: user3, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{min_age: 30}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user2.id in ids
      assert user3.id in ids
    end

    test "only max_age filter", %{user1: user1, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{max_age: 27}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
    end

    test "only IDs filter", %{user1: user1, user3: user3, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{ids: [user1.id, user3.id]}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
      assert user3.id in ids
      assert result.total_count == 2
    end

    test "empty IDs array returns all", %{admin_opts: opts} do
      # Empty array should not restrict (Go: params.Ids = []string{})
      {:ok, result} = Accounts.list_users(%{ids: []}, %{}, opts)
      assert result.total_count >= 3
    end

    test "only project_id filter", %{user1: user1, user2: user2, admin_opts: opts} do
      project = create_project()
      create_user_project(user1, project)

      {:ok, result} = Accounts.list_users(%{project_id: project.id}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
      refute user2.id in ids
      assert result.total_count == 1
    end

    test "only event_id filter", %{user1: user1, user2: user2, admin_opts: opts} do
      project = create_project()
      event = create_event(project)
      create_user_event(user2, event)

      {:ok, result} = Accounts.list_users(%{event_id: event.id}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user2.id in ids
      refute user1.id in ids
      assert result.total_count == 1
    end

    test "query text search matches name", %{user1: user1, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{query: "Alice"}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user1.id in ids
    end

    test "query text search matches email", %{user2: user2, admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{query: "bob@"}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert user2.id in ids
    end

    test "query text search returns nothing for non-matching", %{admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{query: "zzz_nomatch_zzz"}, %{}, opts)
      assert result.total_count == 0
      assert result.edges == []
    end

    test "empty filter returns all", %{admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{}, %{}, opts)
      assert result.total_count >= 3
    end

    test "nil filter returns all", %{admin_opts: opts} do
      {:ok, result} = Accounts.list_users(nil, %{}, opts)
      assert result.total_count >= 3
    end

    test "project and event filters together", %{user1: user1, admin_opts: opts} do
      project = create_project()
      event = create_event(project)
      create_user_project(user1, project)
      create_user_event(user1, event)

      {:ok, result} =
        Accounts.list_users(%{project_id: project.id, event_id: event.id}, %{}, opts)

      assert result.total_count == 1
      assert hd(result.edges).node.id == user1.id
    end

    test "combined church + gender filter", %{
      user1: user1,
      church1: church1,
      admin_opts: opts
    } do
      {:ok, result} =
        Accounts.list_users(%{church_id: church1.id, gender: "FEMALE"}, %{}, opts)

      assert result.total_count == 1
      assert hd(result.edges).node.id == user1.id
    end

    test "non-existent church_id returns empty", %{admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{church_id: "CH00000000000000000000000000"}, %{}, opts)
      assert result.total_count == 0
    end

    test "non-existent project_id returns empty", %{admin_opts: opts} do
      {:ok, result} =
        Accounts.list_users(%{project_id: "PR00000000000000000000000000"}, %{}, opts)

      assert result.total_count == 0
    end

    test "non-existent event_id returns empty", %{admin_opts: opts} do
      {:ok, result} = Accounts.list_users(%{event_id: "EV00000000000000000000000000"}, %{}, opts)
      assert result.total_count == 0
    end
  end

  # ── buildUserFilterParamsCursor equivalent ──
  # Go: TestBuildUserFilterParamsCursor (8 tests) — cursor pagination param building
  # Tests forward/backward, default page size, limit+1 for hasMore, etc.

  describe "list_users — cursor pagination (buildUserFilterParamsCursor equivalent)" do
    setup do
      church = create_church()

      users =
        for i <- 1..7 do
          # Small sleep to ensure distinct inserted_at for ordering
          Process.sleep(10)
          create_user(%{name: "CursorUser #{i}", church_id: church.id})
        end

      admin_opts = [roles: ["admin"]]
      %{church: church, users: users, admin_opts: admin_opts}
    end

    test "forward pagination — first page", %{church: church, admin_opts: opts} do
      {:ok, page} = Accounts.list_users(%{church_id: church.id}, %{first: 3}, opts)

      assert length(page.edges) == 3
      assert page.page_info.has_next_page == true
      assert page.page_info.has_previous_page == false
      assert page.total_count == 7
    end

    test "forward pagination — subsequent page with after cursor", %{
      church: church,
      admin_opts: opts
    } do
      {:ok, page1} = Accounts.list_users(%{church_id: church.id}, %{first: 3}, opts)
      end_cursor = page1.page_info.end_cursor

      {:ok, page2} =
        Accounts.list_users(%{church_id: church.id}, %{first: 3, after: end_cursor}, opts)

      assert length(page2.edges) == 3
      assert page2.page_info.has_next_page == true
      assert page2.page_info.has_previous_page == true

      # No overlap between pages
      ids1 = Enum.map(page1.edges, & &1.node.id)
      ids2 = Enum.map(page2.edges, & &1.node.id)
      assert MapSet.disjoint?(MapSet.new(ids1), MapSet.new(ids2))
    end

    test "forward pagination — last page has has_next_page false", %{
      church: church,
      admin_opts: opts
    } do
      {:ok, page1} = Accounts.list_users(%{church_id: church.id}, %{first: 5}, opts)

      {:ok, page2} =
        Accounts.list_users(
          %{church_id: church.id},
          %{first: 5, after: page1.page_info.end_cursor},
          opts
        )

      assert length(page2.edges) == 2
      assert page2.page_info.has_next_page == false
    end

    test "backward pagination — last N items", %{church: church, admin_opts: opts} do
      {:ok, page} = Accounts.list_users(%{church_id: church.id}, %{last: 3}, opts)

      assert length(page.edges) == 3
      assert page.page_info.has_previous_page == true
      assert page.total_count == 7
    end

    test "backward pagination — with before cursor", %{church: church, admin_opts: opts} do
      # Get last page
      {:ok, last_page} = Accounts.list_users(%{church_id: church.id}, %{last: 3}, opts)
      start_cursor = last_page.page_info.start_cursor

      {:ok, prev_page} =
        Accounts.list_users(%{church_id: church.id}, %{last: 3, before: start_cursor}, opts)

      assert length(prev_page.edges) == 3

      # No overlap
      last_ids = Enum.map(last_page.edges, & &1.node.id)
      prev_ids = Enum.map(prev_page.edges, & &1.node.id)
      assert MapSet.disjoint?(MapSet.new(last_ids), MapSet.new(prev_ids))
    end

    test "default pagination — no first or last returns default page size", %{
      church: church,
      admin_opts: opts
    } do
      {:ok, page} = Accounts.list_users(%{church_id: church.id}, %{}, opts)

      # Default page size is 10, we have 7 users
      assert length(page.edges) == 7
      assert page.page_info.has_next_page == false
    end

    test "total_count is consistent across pages", %{church: church, admin_opts: opts} do
      {:ok, page1} = Accounts.list_users(%{church_id: church.id}, %{first: 3}, opts)
      {:ok, page2} = Accounts.list_users(%{church_id: church.id}, %{last: 3}, opts)

      assert page1.total_count == 7
      assert page2.total_count == 7
    end

    test "total_count reflects filters accurately", %{
      church: church,
      users: users,
      admin_opts: opts
    } do
      project = create_project()
      create_user_project(hd(users), project)
      create_user_project(Enum.at(users, 1), project)

      {:ok, unfiltered} = Accounts.list_users(%{church_id: church.id}, %{}, opts)
      {:ok, filtered} = Accounts.list_users(%{project_id: project.id}, %{}, opts)

      assert unfiltered.total_count == 7
      assert filtered.total_count == 2
    end

    test "pagination with filters", %{church: church, users: users, admin_opts: opts} do
      project = create_project()

      for u <- Enum.take(users, 5) do
        create_user_project(u, project)
      end

      {:ok, page1} =
        Accounts.list_users(
          %{church_id: church.id, project_id: project.id},
          %{first: 2},
          opts
        )

      assert length(page1.edges) == 2
      assert page1.total_count == 5
      assert page1.page_info.has_next_page == true
    end

    test "empty result set", %{admin_opts: opts} do
      {:ok, result} =
        Accounts.list_users(%{church_id: "CH00000000000000000000000000"}, %{first: 10}, opts)

      assert result.edges == []
      assert result.total_count == 0
      assert result.page_info.has_next_page == false
      assert result.page_info.has_previous_page == false
      assert result.page_info.start_cursor == nil
      assert result.page_info.end_cursor == nil
    end
  end

  # ── CRUD operations ──

  describe "assign/remove user to/from project" do
    test "assign and remove user from project" do
      user = create_user()
      project = create_project()

      assert {:ok, returned_user} = Accounts.assign_user_to_project(user.id, project.id)
      assert returned_user.id == user.id

      assert Repo.get_by(ElixirBackend.Accounts.UserProject,
               user_id: user.id,
               project_id: project.id
             )

      assert {:ok, _} = Accounts.remove_user_from_project(user.id, project.id)

      refute Repo.get_by(ElixirBackend.Accounts.UserProject,
               user_id: user.id,
               project_id: project.id
             )
    end

    test "assigning twice is idempotent" do
      user = create_user()
      project = create_project()

      assert {:ok, _} = Accounts.assign_user_to_project(user.id, project.id)
      assert {:ok, _} = Accounts.assign_user_to_project(user.id, project.id)
    end

    test "removing non-existent assignment doesn't error" do
      user = create_user()
      project = create_project()

      assert {:ok, _} = Accounts.remove_user_from_project(user.id, project.id)
    end
  end

  describe "assign user to event" do
    test "assigns user to event" do
      user = create_user()
      project = create_project()
      event = create_event(project)

      assert {:ok, returned_user} = Accounts.assign_user_to_event(user.id, event.id)
      assert returned_user.id == user.id

      assert Repo.get_by(ElixirBackend.Accounts.UserEvent,
               user_id: user.id,
               event_id: event.id
             )
    end

    test "assigning to event twice is idempotent" do
      user = create_user()
      project = create_project()
      event = create_event(project)

      assert {:ok, _} = Accounts.assign_user_to_event(user.id, event.id)
      assert {:ok, _} = Accounts.assign_user_to_event(user.id, event.id)
    end
  end

  describe "lock/unlock user church" do
    test "lock sets church_locked_until ~6 months in future" do
      user = create_user()

      assert {:ok, locked} = Accounts.lock_user_church(user.id)
      assert locked.church_locked_until != nil

      diff = DateTime.diff(locked.church_locked_until, DateTime.utc_now(), :day)
      assert diff >= 179 and diff <= 181
    end

    test "unlock clears church_locked_until" do
      user = create_user()

      {:ok, _locked} = Accounts.lock_user_church(user.id)
      {:ok, unlocked} = Accounts.unlock_user_church(user.id)

      assert unlocked.church_locked_until == nil
    end

    test "lock on non-existent user returns error" do
      assert {:error, :not_found} = Accounts.lock_user_church("US00000000000000000000000000")
    end

    test "unlock on non-existent user returns error" do
      assert {:error, :not_found} = Accounts.unlock_user_church("US00000000000000000000000000")
    end
  end

  # ── Age calculation ──

  describe "calculate_age" do
    test "correct age for known birthdate" do
      assert Accounts.calculate_age(~D[2000-01-15]) == 26
    end

    test "nil birthdate returns nil" do
      assert Accounts.calculate_age(nil) == nil
    end

    test "birthday not yet reached this year" do
      assert Accounts.calculate_age(~D[2000-12-25]) == 25
    end

    test "birthday is today" do
      today = Date.utc_today()
      birthdate = %Date{year: today.year - 30, month: today.month, day: today.day}
      assert Accounts.calculate_age(birthdate) == 30
    end

    test "very young — born this year" do
      today = Date.utc_today()
      birthdate = %Date{year: today.year, month: 1, day: 1}
      assert Accounts.calculate_age(birthdate) == 0
    end
  end

  # ── Age filter edge cases ──
  # Go: zero age values, min without max, max without min

  describe "list_users — age filter edge cases" do
    setup do
      church = create_church()

      young = create_user(%{name: "Young", birthdate: ~D[2010-01-01], church_id: church.id})
      adult = create_user(%{name: "Adult", birthdate: ~D[1990-06-15], church_id: church.id})
      senior = create_user(%{name: "Senior", birthdate: ~D[1960-01-01], church_id: church.id})

      no_birthdate =
        create_user(%{name: "NoBirthdate", birthdate: nil, church_id: church.id})

      admin_opts = [roles: ["admin"]]

      %{
        church: church,
        young: young,
        adult: adult,
        senior: senior,
        no_birthdate: no_birthdate,
        admin_opts: admin_opts
      }
    end

    test "min_age excludes younger users", %{
      church: church,
      young: young,
      adult: adult,
      senior: senior,
      admin_opts: opts
    } do
      {:ok, result} = Accounts.list_users(%{church_id: church.id, min_age: 30}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      refute young.id in ids
      assert adult.id in ids
      assert senior.id in ids
    end

    test "max_age excludes older users", %{
      church: church,
      young: young,
      senior: senior,
      admin_opts: opts
    } do
      {:ok, result} = Accounts.list_users(%{church_id: church.id, max_age: 20}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert young.id in ids
      refute senior.id in ids
    end

    test "age filters exclude users without birthdate", %{
      church: church,
      no_birthdate: no_birthdate,
      admin_opts: opts
    } do
      {:ok, result} = Accounts.list_users(%{church_id: church.id, min_age: 0}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      refute no_birthdate.id in ids
    end

    test "narrow age range returns only matching", %{
      church: church,
      adult: adult,
      admin_opts: opts
    } do
      # Adult born 1990-06-15, age ~35
      {:ok, result} =
        Accounts.list_users(%{church_id: church.id, min_age: 34, max_age: 37}, %{}, opts)

      ids = Enum.map(result.edges, & &1.node.id)
      assert adult.id in ids
      assert result.total_count == 1
    end
  end
end
