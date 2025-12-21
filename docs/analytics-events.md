# Analytics Events Reference

This document describes all analytics events tracked in Wayfarer. Use this reference to understand what user actions are being tracked and what data is available for analysis.

## Overview

Events are sent to both **RudderStack** and **PostHog** when users interact with the application. Each event captures a specific user action along with relevant context.

### How to Read This Document

Each section below lists events in a table with:

- **Event Name**: The identifier used in analytics platforms (use this when building reports)
- **When It Fires**: Plain-language description of the user action that triggers this event
- **Data Captured**: Additional information sent with the event

---

## User Sessions

### Authentication

| Event Name         | When It Fires                  | Data Captured |
| ------------------ | ------------------------------ | ------------- |
| `login_completed`  | User successfully logs in      | None          |
| `logout_completed` | User logs out of their account | None          |

### User Identity

When a user logs in, the following profile information is associated with their session:

| Trait            | Description                         | Example Values         |
| ---------------- | ----------------------------------- | ---------------------- |
| `age_group`      | User's age bracket                  | "13 - 18", "19 - 25"   |
| `gender`         | User's gender                       | -                      |
| `church_id`      | Unique identifier of user's church  | -                      |
| `church_name`    | Name of user's church               | "First Baptist Church" |
| `church_country` | Country where the church is located | "Norway", "Sweden"     |

> **Privacy Note**: User IDs are hashed (SHA-256) before being sent to analytics platforms to protect user privacy.

---

## Challenges & Quizzes

Challenges are tasks or activities users can complete to earn points.

### Challenge Interactions

| Event Name               | When It Fires                                | Data Captured                                                    |
| ------------------------ | -------------------------------------------- | ---------------------------------------------------------------- |
| `challenge_opened`       | User opens a challenge to view its details   | `challenge_id`, `challenge_name`, `challenge_type` (simple/quiz) |
| `challenge_link_clicked` | User clicks the action button on a challenge | `challenge_id`, `challenge_name`, `is_external` (true/false)     |

### Quiz Progress

Quizzes are a type of challenge where users answer questions.

| Event Name              | When It Fires                           | Data Captured                                                                   |
| ----------------------- | --------------------------------------- | ------------------------------------------------------------------------------- |
| `quiz_started`          | User begins a quiz attempt              | `quiz_id`, `quiz_name`, `challenge_id`                                          |
| `quiz_answer_submitted` | User submits an answer to a question    | `question_id`, `is_correct`, `current_question`, `total_questions`              |
| `quiz_completed`        | User finishes all questions in a quiz   | `quiz_id`, `quiz_name`, `submission_id`, `score`, `max_score`, `points_awarded` |
| `quiz_abandoned`        | User leaves a quiz without finishing it | `quiz_id`, `quiz_name`, `questions_attempted`, `total_questions`                |

**Useful metrics you can derive:**

- Quiz completion rate: `quiz_completed` / `quiz_started`
- Average score: Mean of `score` / `max_score` from `quiz_completed` events
- Drop-off point: `questions_attempted` from `quiz_abandoned` events

---

## Achievements

Achievements are badges or rewards users earn for completing specific goals.

| Event Name            | When It Fires                     | Data Captured                                                    |
| --------------------- | --------------------------------- | ---------------------------------------------------------------- |
| `achievement_clicked` | User taps on an achievement badge | `achievement_id`, `achievement_name`, `is_unlocked` (true/false) |

---

## Leaderboards

Leaderboards show rankings of users, teams, or super teams by points.

| Event Name                | When It Fires                             | Data Captured                         |
| ------------------------- | ----------------------------------------- | ------------------------------------- |
| `leaderboard_tab_changed` | User switches between leaderboard views   | `from` (previous tab), `to` (new tab) |
| `team_leaderboard_viewed` | User views the leaderboard for their team | `team_id`                             |

---

## Points & Rewards

| Event Name                 | When It Fires                                    | Data Captured |
| -------------------------- | ------------------------------------------------ | ------------- |
| `points_history_opened`    | User opens their points history log              | None          |
| `how_to_get_points_opened` | User opens the help section about earning points | None          |

---

## Push Notifications

| Event Name                   | When It Fires                                    | Data Captured                                          |
| ---------------------------- | ------------------------------------------------ | ------------------------------------------------------ |
| `push_permission_requested`  | App asks the browser for notification permission | `permission_granted` (true/false), `permission_result` |
| `push_subscription_enabled`  | User successfully enables push notifications     | None                                                   |
| `push_subscription_disabled` | User successfully disables push notifications    | None                                                   |
| `push_notifications_toggled` | User flips the notifications toggle in settings  | `enabled` (true/false)                                 |

---

## User Preferences

| Event Name           | When It Fires                 | Data Captured                              |
| -------------------- | ----------------------------- | ------------------------------------------ |
| `language_changed`   | User changes the app language | `from` (old language), `to` (new language) |
| `color_mode_changed` | User changes light/dark mode  | `from` (old mode), `to` (new mode)         |

---

## Consent Management

| Event Name         | When It Fires                 | Data Captured |
| ------------------ | ----------------------------- | ------------- |
| `consent_accepted` | User accepts a consent prompt | `consent_id`  |
| `consent_rejected` | User rejects a consent prompt | `consent_id`  |

---

## Page Views

Page views are tracked automatically whenever users navigate to a new page. These events capture:

- The page URL
- The page title
- Standard session information

No manual instrumentation is required for page view tracking.

---

## Common Analysis Queries

Here are some questions this data can help answer:

1. **Engagement**: Which challenges are most popular? (Count of `challenge_opened` by `challenge_id`)
2. **Quiz performance**: What's the average quiz score? (Mean of `score`/`max_score` from `quiz_completed`)
3. **Retention signals**: How many users enable push notifications? (`push_subscription_enabled` count)
4. **Feature adoption**: Are users exploring leaderboards? (`leaderboard_tab_changed` count)
5. **Accessibility**: Which languages are most used? (`language_changed` events + user traits)
