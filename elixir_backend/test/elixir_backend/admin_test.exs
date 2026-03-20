defmodule ElixirBackend.AdminTest do
  use ElixirBackend.DataCase, async: true

  alias ElixirBackend.Admin
  alias ElixirBackend.TestHelpers

  describe "dashboard_stats/0" do
    test "returns aggregate statistics" do
      _user = TestHelpers.create_user()
      _project = TestHelpers.create_project()

      {:ok, stats} = Admin.dashboard_stats()

      assert is_integer(stats.total_users)
      assert stats.total_users >= 1
      assert is_integer(stats.total_projects)
      assert stats.total_projects >= 1
      assert is_integer(stats.total_challenges)
      assert is_integer(stats.total_points_awarded)
      assert is_integer(stats.new_users_last_7_days)
      assert is_integer(stats.active_projects_count)
    end
  end

  describe "church_admin_statistics/1" do
    test "returns stub statistics" do
      user = TestHelpers.create_user()
      {:ok, stats} = Admin.church_admin_statistics(user.id)

      assert is_binary(stats.church_id)
      assert is_list(stats.age_groups)
      assert is_integer(stats.total_users_in_teams)
    end
  end
end
