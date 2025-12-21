# Analytics Events

This document describes all analytics events tracked in the Interact frontend. Events are sent to both RudderStack and PostHog via the `useAnalytics()` composable.

## Authentication

| Event              | Description                 | Properties |
| ------------------ | --------------------------- | ---------- |
| `login_completed`  | User successfully logged in | -          |
| `logout_completed` | User logged out             | -          |

## Challenges

| Event                    | Description                                 | Properties                                                                  |
| ------------------------ | ------------------------------------------- | --------------------------------------------------------------------------- |
| `challenge_opened`       | User opened a challenge detail page         | `challenge_id`, `challenge_name`, `challenge_type` (`'simple'` \| `'quiz'`) |
| `challenge_link_clicked` | User clicked the button on a challenge card | `challenge_id`, `challenge_name`, `is_external`                             |

## Quiz

| Event                   | Description                             | Properties                                                                      |
| ----------------------- | --------------------------------------- | ------------------------------------------------------------------------------- |
| `quiz_started`          | User started a new quiz attempt         | `quiz_id`, `quiz_name`, `challenge_id`                                          |
| `quiz_answer_submitted` | User submitted an answer to a question  | `question_id`, `is_correct`, `current_question`, `total_questions`              |
| `quiz_completed`        | User finished a quiz                    | `quiz_id`, `quiz_name`, `submission_id`, `score`, `max_score`, `points_awarded` |
| `quiz_abandoned`        | User closed a quiz before completing it | `quiz_id`, `quiz_name`, `questions_attempted`, `total_questions`                |

## Achievements

| Event                 | Description                          | Properties                                          |
| --------------------- | ------------------------------------ | --------------------------------------------------- |
| `achievement_clicked` | User clicked on an achievement badge | `achievement_id`, `achievement_name`, `is_unlocked` |

## Leaderboard

| Event                     | Description                          | Properties   |
| ------------------------- | ------------------------------------ | ------------ |
| `leaderboard_tab_changed` | User switched leaderboard tabs       | `from`, `to` |
| `team_leaderboard_viewed` | User viewed their team's leaderboard | `team_id`    |

## Points

| Event                      | Description                              | Properties |
| -------------------------- | ---------------------------------------- | ---------- |
| `points_history_opened`    | User opened their points history         | -          |
| `how_to_get_points_opened` | User opened the "how to get points" info | -          |

## Push Notifications

| Event                        | Description                                            | Properties                                |
| ---------------------------- | ------------------------------------------------------ | ----------------------------------------- |
| `push_permission_requested`  | Browser permission was requested                       | `permission_granted`, `permission_result` |
| `push_subscription_enabled`  | User successfully subscribed to push notifications     | -                                         |
| `push_subscription_disabled` | User successfully unsubscribed from push notifications | -                                         |
| `push_notifications_toggled` | User toggled the notifications switch in settings      | `enabled`                                 |

## Settings

| Event                | Description                   | Properties   |
| -------------------- | ----------------------------- | ------------ |
| `language_changed`   | User changed the app language | `from`, `to` |
| `color_mode_changed` | User changed the color mode   | `from`, `to` |

## Consent

| Event              | Description             | Properties   |
| ------------------ | ----------------------- | ------------ |
| `consent_accepted` | User accepted a consent | `consent_id` |
| `consent_rejected` | User rejected a consent | `consent_id` |

## Page Views

Page views are automatically tracked on every route navigation via `router.afterEach()`. This uses RudderStack's `page()` method which captures the current URL and page title.

## User Identification

When a user logs in, they are identified with the following traits:

- `age_group` - Age bracket (e.g., "13 - 18", "19 - 25")
- `gender` - User's gender
- `church_id` - ID of the user's church
- `church_name` - Name of the user's church
- `church_country` - Country of the user's church

Note: User IDs are hashed with SHA-256 before being sent to analytics platforms.
