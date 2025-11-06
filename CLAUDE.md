# Wayfarer

Wayfarer is a gamification system used as the core for a multitude of projects, usually bible studies associated with a youth camp.

## Database

The database schema is defined in `schema.sql` and uses CockroachDB.

### ID Format

All primary keys use prefixed ULIDs for better readability and debugging:
- Format: `XX + 26-character ULID = 28 characters total`
- Example: `CH01ARZ3NDEKTSV4RRFFQ69G5FAV` (Church), `US01ARZ3NDEKTSV4RRFFQ69G5FAV` (User)
- ULIDs are lexicographically sortable by timestamp
- Must be generated in application code with the appropriate prefix

### ID Prefix Reference

- `CH` - Churches
- `US` - Users
- `UR` - User Roles
- `PR` - Projects
- `EV` - Events
- `ST` - SuperTeams
- `TM` - Teams
- `SK` - Streaks
- `CL` - Challenges
- `AC` - Achievements
- `RA` - Reading Achievement Articles
- `LT` - Listening Achievement Tracks
- `SA` - Score Adjustments

### Core Entities

#### Churches (`churches`)
Organizations that users belong to. Categorized by size (S, L, XL).

#### Users (`users`)
End users of the system. Each user belongs to one church and can participate in multiple projects and events.

#### Projects (`projects`)
Top-level groupings containing branding, events, achievements, challenges, and streaks. Projects have start/end dates and can be archived.

#### Events (`events`)
In-person or time-bound events within a project. Events inherit from their parent project and can have their own challenges and achievements.

#### Teams (`teams`)
Groups of users within a project. Teams can optionally belong to a SuperTeam. Users join teams via unique join codes.

#### SuperTeams (`super_teams`)
Collections of multiple teams, creating a two-level team hierarchy within a project.

#### Streaks (`streaks`)
Activity tracking over time periods. Uses `DATERANGE[]` array to define which date ranges are relevant for the streak.

#### Challenges (`challenges`)
Tasks or activities users can complete. Challenges belong to a project and optionally to an event. They can have end times and publish dates.

#### Achievements (`achievements`)
Rewards users can earn. Four types:
- **Simple**: Basic achievement with just a name, description, and points
- **Reading**: Requires reading specific articles (stored in `reading_achievement_articles`)
- **Listening**: Requires listening to specific tracks (stored in `listening_achievement_tracks`)
- **Streak**: Requires maintaining a streak for a specified duration (linked to `streaks`)

### Progress Tracking

User progress is tracked through several junction tables:
- `user_projects` / `user_events` - Project/event participation
- `team_members` - Team membership
- `user_achievements` / `team_achievements` / `super_team_achievements` - Achievement awards
- `user_challenge_completions` - Challenge completion
- `user_reading_progress` / `user_listening_progress` - Progress on reading/listening achievements
- `user_streak_activity` - Daily activity for streaks

### Scoring

Scores are calculated on-demand from achievements and score adjustments. There are no pre-aggregated score tables.

Score adjustments are logged in `score_adjustments` for audit purposes, tracking manual point changes for users, teams, or super teams.

### GraphQL APIs

The system exposes three separate GraphQL APIs defined in the `gql/` directory:
- **User API** (`user.graphqls`) - For end users (mobile/web apps)
- **Admin API** (`admin.graphqls`) - For system administrators
- **M2M API** (`m2m.graphqls`) - For external systems to notify Wayfarer about events

## Concepts

The system has some concepts that are important to understand.

### Projects

Projects are the top-level groupings of everything. Every other concept mentioned in this document are associated with _one_ project. Users are not associated with a single project, but are also top-level.

Projects contain data about the eg. bible study, such as branding (logo, colors, fonts), events, points, achievements and leaderboards.

### Events

Each project can contain multiple events, which usually is some form of in-person event. The event contains some data, such as points, achievements and leaderboards.
- In queries, use named, not numbered parameters. For example @userid::text.
- No signatures on commits
- When creating commits sign with: Assited by [MODEL] via [Tool]
- Generate unit tests for functions you write. Use the unit tests to verify correctnes. When mocking use mockery, with the config `.mockery.yml`.
- Put notes into notes folder. Before doing a big investigation check if we already have notes on the system.
- Use "make gqltest" in ./backend/ dir to run a sanity check after changes to schemas and resolvers.
- Update the notes if you make changes to schemas and other things that invalidate the notest.