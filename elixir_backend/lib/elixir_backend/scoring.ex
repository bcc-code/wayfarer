defmodule ElixirBackend.Scoring do
  @moduledoc """
  Context module for score journal management and leaderboard queries.
  """

  import Ecto.Query
  alias ElixirBackend.Repo
  alias ElixirBackend.ULID
  alias ElixirBackend.Scoring.ScoreJournal

  # ── Score Journal Read ──

  def get_entry(id) do
    case Repo.get(ScoreJournal, id) do
      nil -> {:error, :not_found}
      entry -> {:ok, entry}
    end
  end

  def list_entries(filter \\ %{}, pagination_opts \\ %{}) do
    query = from(s in ScoreJournal)
    query = apply_journal_filter(query, filter)
    total_count = Repo.aggregate(query, :count)

    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    items =
      query
      |> order_by([s], desc: s.created_at, desc: s.id)
      |> limit(^limit)
      |> Repo.all()

    edges = Enum.map(items, fn item -> %{cursor: item.id, node: item} end)

    {:ok,
     %{
       edges: edges,
       page_info: %{has_next_page: length(items) == limit, has_previous_page: false},
       total_count: total_count
     }}
  end

  # ── Score Journal Write ──

  def create_entry(attrs) do
    id = ULID.new_score_journal_id()
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    %ScoreJournal{}
    |> ScoreJournal.changeset(
      attrs
      |> Map.put(:id, id)
      |> Map.put(:created_at, now)
    )
    |> Repo.insert()
  end

  def create_adjustment(attrs) do
    create_entry(Map.put(attrs, :source_type, "MANUAL"))
  end

  def create_team_adjustment(attrs) do
    team_id = attrs[:team_id] || attrs["team_id"]
    project_id = attrs[:project_id] || attrs["project_id"]
    event_id = attrs[:event_id] || attrs["event_id"]
    points = attrs[:points] || attrs["points"]
    reason = attrs[:reason] || attrs["reason"]
    distribution_mode = attrs[:distribution_mode] || attrs["distribution_mode"]

    # Get team members
    members =
      from(tm in ElixirBackend.Teams.TeamMember,
        where: tm.team_id == ^team_id,
        select: tm.user_id
      )
      |> Repo.all()

    if members == [] do
      {:ok, []}
    else
      member_points =
        case distribution_mode do
          "SPLIT" -> div(points, length(members))
          "EACH" -> points
          _ -> points
        end

      entries =
        Enum.map(members, fn user_id ->
          {:ok, entry} =
            create_entry(%{
              project_id: project_id,
              user_id: user_id,
              event_id: event_id,
              points: member_points,
              source_type: "MANUAL",
              reason: reason
            })

          entry
        end)

      # Update leaderboards for each entry
      Enum.each(entries, &update_leaderboards/1)

      {:ok, entries}
    end
  end

  def delete_entry(id) do
    with {:ok, entry} <- get_entry(id) do
      Repo.delete(entry)
    end
  end

  # ── Leaderboard ──

  def get_project_leaderboard(project_id, entity_type, filter \\ %{}, pagination_opts \\ %{}) do
    case entity_type do
      "PERSONS" -> get_project_person_leaderboard(project_id, filter, pagination_opts)
      "TEAMS" -> get_project_team_leaderboard(project_id, filter, pagination_opts)
      "SUPERTEAMS" -> get_project_superteam_leaderboard(project_id, filter, pagination_opts)
      "CHURCHES" -> get_project_church_leaderboard(project_id, filter, pagination_opts)
      _ -> {:error, "invalid entity type"}
    end
  end

  def get_event_leaderboard(event_id, entity_type, filter \\ %{}, pagination_opts \\ %{}) do
    case entity_type do
      "PERSONS" -> get_event_person_leaderboard(event_id, filter, pagination_opts)
      "TEAMS" -> get_event_team_leaderboard(event_id, filter, pagination_opts)
      "SUPERTEAMS" -> get_event_superteam_leaderboard(event_id, filter, pagination_opts)
      "CHURCHES" -> get_event_church_leaderboard(event_id, filter, pagination_opts)
      _ -> {:error, "invalid entity type"}
    end
  end

  # Update leaderboard tables after a score journal entry is created
  def update_leaderboards(%ScoreJournal{} = entry) do
    update_person_leaderboard(entry)
    update_church_leaderboard(entry)
    update_team_leaderboards(entry)
  end

  # ── Private: Leaderboard Queries ──

  defp get_project_person_leaderboard(project_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(lp in "leaderboard_project_persons",
        join: u in ElixirBackend.Accounts.User,
        on: u.id == lp.user_id,
        where: lp.project_id == ^project_id and lp.score > 0,
        select: %{
          id: lp.user_id,
          name: u.name,
          score: lp.score,
          last_score_at: lp.last_score_at
        },
        order_by: [desc: lp.score, desc: lp.last_score_at, asc: u.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  defp get_project_team_leaderboard(project_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(lt in "leaderboard_project_teams",
        join: t in ElixirBackend.Teams.Team,
        on: t.id == lt.team_id,
        where: lt.project_id == ^project_id and lt.score > 0,
        select: %{
          id: lt.team_id,
          name: t.name,
          score: lt.score,
          last_score_at: lt.last_score_at
        },
        order_by: [desc: lt.score, desc: lt.last_score_at, asc: t.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  defp get_project_superteam_leaderboard(project_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(ls in "leaderboard_project_superteams",
        join: st in ElixirBackend.Teams.SuperTeam,
        on: st.id == ls.super_team_id,
        where: ls.project_id == ^project_id and ls.score > 0,
        select: %{
          id: ls.super_team_id,
          name: st.name,
          score: ls.score,
          last_score_at: ls.last_score_at
        },
        order_by: [desc: ls.score, desc: ls.last_score_at, asc: st.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  defp get_project_church_leaderboard(project_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(lc in "leaderboard_project_churches",
        join: c in ElixirBackend.Churches.Church,
        on: c.id == lc.church_id,
        where: lc.project_id == ^project_id and lc.score > 0,
        select: %{
          id: lc.church_id,
          name: c.name,
          score: lc.score,
          last_score_at: lc.last_score_at
        },
        order_by: [desc: lc.score, desc: lc.last_score_at, asc: c.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  defp get_event_person_leaderboard(event_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(lp in "leaderboard_event_persons",
        join: u in ElixirBackend.Accounts.User,
        on: u.id == lp.user_id,
        where: lp.event_id == ^event_id and lp.score > 0,
        select: %{
          id: lp.user_id,
          name: u.name,
          score: lp.score,
          last_score_at: lp.last_score_at
        },
        order_by: [desc: lp.score, desc: lp.last_score_at, asc: u.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  defp get_event_team_leaderboard(event_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(lt in "leaderboard_event_teams",
        join: t in ElixirBackend.Teams.Team,
        on: t.id == lt.team_id,
        where: lt.event_id == ^event_id and lt.score > 0,
        select: %{
          id: lt.team_id,
          name: t.name,
          score: lt.score,
          last_score_at: lt.last_score_at
        },
        order_by: [desc: lt.score, desc: lt.last_score_at, asc: t.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  defp get_event_superteam_leaderboard(event_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(ls in "leaderboard_event_superteams",
        join: st in ElixirBackend.Teams.SuperTeam,
        on: st.id == ls.super_team_id,
        where: ls.event_id == ^event_id and ls.score > 0,
        select: %{
          id: ls.super_team_id,
          name: st.name,
          score: ls.score,
          last_score_at: ls.last_score_at
        },
        order_by: [desc: ls.score, desc: ls.last_score_at, asc: st.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  defp get_event_church_leaderboard(event_id, _filter, pagination_opts) do
    limit = pagination_opts[:first] || pagination_opts[:last] || 25

    query =
      from(lc in "leaderboard_event_churches",
        join: c in ElixirBackend.Churches.Church,
        on: c.id == lc.church_id,
        where: lc.event_id == ^event_id and lc.score > 0,
        select: %{
          id: lc.church_id,
          name: c.name,
          score: lc.score,
          last_score_at: lc.last_score_at
        },
        order_by: [desc: lc.score, desc: lc.last_score_at, asc: c.name],
        limit: ^limit
      )

    items = Repo.all(query)
    build_leaderboard_connection(items, limit)
  end

  # ── Private: Leaderboard Updates ──

  defp update_person_leaderboard(entry) do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    # Project person
    Repo.insert_all(
      "leaderboard_project_persons",
      [
        %{
          project_id: entry.project_id,
          user_id: entry.user_id,
          score: entry.points,
          last_score_at: entry.created_at,
          updated_at: now
        }
      ],
      on_conflict:
        from(lp in "leaderboard_project_persons",
          update: [
            set: [
              score: fragment("? + ?", lp.score, ^entry.points),
              last_score_at: ^entry.created_at,
              updated_at: ^now
            ]
          ]
        ),
      conflict_target: [:project_id, :user_id]
    )

    # Event person (if event_id present)
    if entry.event_id do
      Repo.insert_all(
        "leaderboard_event_persons",
        [
          %{
            event_id: entry.event_id,
            user_id: entry.user_id,
            score: entry.points,
            last_score_at: entry.created_at,
            updated_at: now
          }
        ],
        on_conflict:
          from(lp in "leaderboard_event_persons",
            update: [
              set: [
                score: fragment("? + ?", lp.score, ^entry.points),
                last_score_at: ^entry.created_at,
                updated_at: ^now
              ]
            ]
          ),
        conflict_target: [:event_id, :user_id]
      )
    end
  end

  defp update_church_leaderboard(entry) do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    user = Repo.get(ElixirBackend.Accounts.User, entry.user_id)

    if user && user.church_id do
      Repo.insert_all(
        "leaderboard_project_churches",
        [
          %{
            project_id: entry.project_id,
            church_id: user.church_id,
            score: entry.points,
            last_score_at: entry.created_at,
            updated_at: now
          }
        ],
        on_conflict:
          from(lc in "leaderboard_project_churches",
            update: [
              set: [
                score: fragment("? + ?", lc.score, ^entry.points),
                last_score_at: ^entry.created_at,
                updated_at: ^now
              ]
            ]
          ),
        conflict_target: [:project_id, :church_id]
      )

      if entry.event_id do
        Repo.insert_all(
          "leaderboard_event_churches",
          [
            %{
              event_id: entry.event_id,
              church_id: user.church_id,
              score: entry.points,
              last_score_at: entry.created_at,
              updated_at: now
            }
          ],
          on_conflict:
            from(lc in "leaderboard_event_churches",
              update: [
                set: [
                  score: fragment("? + ?", lc.score, ^entry.points),
                  last_score_at: ^entry.created_at,
                  updated_at: ^now
                ]
              ]
            ),
          conflict_target: [:event_id, :church_id]
        )
      end
    end
  end

  defp update_team_leaderboards(entry) do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    # Find user's team memberships in this project
    team_memberships =
      from(tm in ElixirBackend.Teams.TeamMember,
        join: t in ElixirBackend.Teams.Team,
        on: t.id == tm.team_id,
        where: tm.user_id == ^entry.user_id and t.project_id == ^entry.project_id,
        select: %{team_id: t.id, super_team_id: t.super_team_id}
      )
      |> Repo.all()

    Enum.each(team_memberships, fn membership ->
      # Project team
      Repo.insert_all(
        "leaderboard_project_teams",
        [
          %{
            project_id: entry.project_id,
            team_id: membership.team_id,
            score: entry.points,
            last_score_at: entry.created_at,
            updated_at: now
          }
        ],
        on_conflict:
          from(lt in "leaderboard_project_teams",
            update: [
              set: [
                score: fragment("? + ?", lt.score, ^entry.points),
                last_score_at: ^entry.created_at,
                updated_at: ^now
              ]
            ]
          ),
        conflict_target: [:project_id, :team_id]
      )

      # Event team
      if entry.event_id do
        Repo.insert_all(
          "leaderboard_event_teams",
          [
            %{
              event_id: entry.event_id,
              team_id: membership.team_id,
              score: entry.points,
              last_score_at: entry.created_at,
              updated_at: now
            }
          ],
          on_conflict:
            from(lt in "leaderboard_event_teams",
              update: [
                set: [
                  score: fragment("? + ?", lt.score, ^entry.points),
                  last_score_at: ^entry.created_at,
                  updated_at: ^now
                ]
              ]
            ),
          conflict_target: [:event_id, :team_id]
        )
      end

      # Superteam
      if membership.super_team_id do
        Repo.insert_all(
          "leaderboard_project_superteams",
          [
            %{
              project_id: entry.project_id,
              super_team_id: membership.super_team_id,
              score: entry.points,
              last_score_at: entry.created_at,
              updated_at: now
            }
          ],
          on_conflict:
            from(ls in "leaderboard_project_superteams",
              update: [
                set: [
                  score: fragment("? + ?", ls.score, ^entry.points),
                  last_score_at: ^entry.created_at,
                  updated_at: ^now
                ]
              ]
            ),
          conflict_target: [:project_id, :super_team_id]
        )

        if entry.event_id do
          Repo.insert_all(
            "leaderboard_event_superteams",
            [
              %{
                event_id: entry.event_id,
                super_team_id: membership.super_team_id,
                score: entry.points,
                last_score_at: entry.created_at,
                updated_at: now
              }
            ],
            on_conflict:
              from(ls in "leaderboard_event_superteams",
                update: [
                  set: [
                    score: fragment("? + ?", ls.score, ^entry.points),
                    last_score_at: ^entry.created_at,
                    updated_at: ^now
                  ]
                ]
              ),
            conflict_target: [:event_id, :super_team_id]
          )
        end
      end
    end)
  end

  # ── Private: Helpers ──

  defp build_leaderboard_connection(items, limit) do
    ranked =
      items
      |> Enum.with_index(1)
      |> Enum.map(fn {item, rank} ->
        Map.merge(item, %{rank: rank, description: "", tags: []})
      end)

    edges = Enum.map(ranked, fn item -> %{cursor: "#{item.rank}", node: item} end)

    {:ok,
     %{
       edges: edges,
       page_info: %{has_next_page: length(items) == limit, has_previous_page: false},
       total_count: length(items),
       me: nil
     }}
  end

  defp apply_journal_filter(query, filter) when is_map(filter) do
    Enum.reduce(filter, query, fn
      {:project_id, pid}, q when is_binary(pid) ->
        where(q, [s], s.project_id == ^pid)

      {:user_id, uid}, q when is_binary(uid) ->
        where(q, [s], s.user_id == ^uid)

      {:event_id, eid}, q when is_binary(eid) ->
        where(q, [s], s.event_id == ^eid)

      {:challenge_id, cid}, q when is_binary(cid) ->
        where(q, [s], s.challenge_id == ^cid)

      {:source_type, st}, q when is_binary(st) ->
        where(q, [s], s.source_type == ^st)

      {:ids, ids}, q when is_list(ids) and ids != [] ->
        where(q, [s], s.id in ^ids)

      _, q ->
        q
    end)
  end

  defp apply_journal_filter(query, _), do: query
end
