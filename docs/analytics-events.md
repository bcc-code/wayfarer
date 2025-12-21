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

**Metrics & stats ideas:**

- Daily/weekly/monthly active users (unique users with `login_completed`)
- Login frequency per user over time
- Session duration (time between `login_completed` and `logout_completed`)
- Peak login times (hour of day, day of week)
- Returning vs new users ratio

### User Identity

When a user logs in, the following profile information is associated with their session:

| Trait            | Description                         | Example Values       |
| ---------------- | ----------------------------------- | -------------------- |
| `age_group`      | User's age bracket                  | "13 - 18", "19 - 25" |
| `gender`         | User's gender                       | -                    |
| `church_id`      | Unique identifier of user's church  | -                    |
| `church_name`    | Name of user's church               | "Oslo og Follo"      |
| `church_country` | Country where the church is located | "Norway", "Germany"  |

> **Privacy Note**: User IDs are hashed (SHA-256) before being sent to analytics platforms to protect user privacy.

**Metrics & stats ideas:**

- User demographics breakdown (age groups, gender distribution)
- Engagement by church size or country
- Most active churches (by user count or total activity)
- Geographic distribution of users
- Cross-segment comparisons (e.g., do younger users complete more quizzes?)

---

## Challenges & Quizzes

Challenges are tasks or activities users can complete to earn points.

### Challenge Interactions

| Event Name               | When It Fires                                | Data Captured                                                    |
| ------------------------ | -------------------------------------------- | ---------------------------------------------------------------- |
| `challenge_opened`       | User opens a challenge to view its details   | `challenge_id`, `challenge_name`, `challenge_type` (simple/quiz) |
| `challenge_link_clicked` | User clicks the action button on a challenge | `challenge_id`, `challenge_name`, `is_external` (true/false)     |

**Metrics & stats ideas:**

- Most viewed challenges (count of `challenge_opened` by `challenge_id`)
- Challenge click-through rate (`challenge_link_clicked` / `challenge_opened`)
- External vs internal link engagement (filter by `is_external`)
- Challenge discovery patterns (which challenges do users view first?)
- Challenges with high views but low clicks (potential UX issues)

### Quiz Progress

Quizzes are a type of challenge where users answer questions.

| Event Name              | When It Fires                           | Data Captured                                                                   |
| ----------------------- | --------------------------------------- | ------------------------------------------------------------------------------- |
| `quiz_started`          | User begins a quiz attempt              | `quiz_id`, `quiz_name`, `challenge_id`                                          |
| `quiz_answer_submitted` | User submits an answer to a question    | `question_id`, `is_correct`, `current_question`, `total_questions`              |
| `quiz_completed`        | User finishes all questions in a quiz   | `quiz_id`, `quiz_name`, `submission_id`, `score`, `max_score`, `points_awarded` |
| `quiz_abandoned`        | User leaves a quiz without finishing it | `quiz_id`, `quiz_name`, `questions_attempted`, `total_questions`                |

**Metrics & stats ideas:**

- Quiz completion rate (`quiz_completed` / `quiz_started`)
- Average score per quiz (mean of `score` / `max_score`)
- Quiz abandonment rate and drop-off points (`questions_attempted` from `quiz_abandoned`)
- Question difficulty analysis (% correct per `question_id`)
- Points earned from quizzes over time
- Time to complete quizzes (time between `quiz_started` and `quiz_completed`)
- Repeat quiz attempts per user
- Hardest quizzes (lowest average scores)
- Most engaging quizzes (highest completion rates)

---

## Achievements

Achievements are badges or rewards users earn for completing specific goals.

| Event Name            | When It Fires                     | Data Captured                                                    |
| --------------------- | --------------------------------- | ---------------------------------------------------------------- |
| `achievement_clicked` | User taps on an achievement badge | `achievement_id`, `achievement_name`, `is_unlocked` (true/false) |

**Metrics & stats ideas:**

- Most viewed achievements (count by `achievement_id`)
- Locked vs unlocked achievement interest (compare clicks by `is_unlocked`)
- Achievement discovery rate (which achievements get attention?)
- User interest in unearned achievements (clicks on locked badges)
- Achievement page engagement over time

---

## Leaderboards

Leaderboards show rankings of users, teams, or super teams by points.

| Event Name                | When It Fires                             | Data Captured                         |
| ------------------------- | ----------------------------------------- | ------------------------------------- |
| `leaderboard_tab_changed` | User switches between leaderboard views   | `from` (previous tab), `to` (new tab) |
| `team_leaderboard_viewed` | User views the leaderboard for their team | `team_id`                             |

**Metrics & stats ideas:**

- Most popular leaderboard views (which tabs are viewed most?)
- Tab switching patterns (common navigation flows)
- Team leaderboard engagement by team size
- Leaderboard views over time (do users check more during events?)
- Correlation between leaderboard views and challenge completion

---

## Points & Rewards

| Event Name                 | When It Fires                                    | Data Captured |
| -------------------------- | ------------------------------------------------ | ------------- |
| `points_history_opened`    | User opens their points history log              | None          |
| `how_to_get_points_opened` | User opens the help section about earning points | None          |

**Metrics & stats ideas:**

- Points history engagement (how often do users check their history?)
- Help section usage (are users confused about earning points?)
- Correlation between viewing help and subsequent challenge completion
- User segments most likely to check points history (new vs returning users)

---

## Push Notifications

| Event Name                   | When It Fires                                    | Data Captured                                          |
| ---------------------------- | ------------------------------------------------ | ------------------------------------------------------ |
| `push_permission_requested`  | App asks the browser for notification permission | `permission_granted` (true/false), `permission_result` |
| `push_subscription_enabled`  | User successfully enables push notifications     | None                                                   |
| `push_subscription_disabled` | User successfully disables push notifications    | None                                                   |
| `push_notifications_toggled` | User flips the notifications toggle in settings  | `enabled` (true/false)                                 |

**Metrics & stats ideas:**

- Push notification opt-in rate (`permission_granted` = true / total requests)
- Subscription churn (users who enable then disable)
- Permission denial rate and trends over time
- Opt-in rate by user segment (age group, church, etc.)
- Impact of push notifications on engagement (compare active users with/without push enabled)

---

## User Preferences

| Event Name           | When It Fires                 | Data Captured                              |
| -------------------- | ----------------------------- | ------------------------------------------ |
| `language_changed`   | User changes the app language | `from` (old language), `to` (new language) |
| `color_mode_changed` | User changes light/dark mode  | `from` (old mode), `to` (new mode)         |

**Metrics & stats ideas:**

- Language distribution across users
- Most common language switches (which `from` → `to` pairs?)
- Dark mode adoption rate
- Preference changes over time (are more users switching to dark mode?)
- Correlation between preferences and engagement

---

## Consent Management

| Event Name         | When It Fires                 | Data Captured |
| ------------------ | ----------------------------- | ------------- |
| `consent_accepted` | User accepts a consent prompt | `consent_id`  |
| `consent_rejected` | User rejects a consent prompt | `consent_id`  |

**Metrics & stats ideas:**

- Consent acceptance rate per consent type
- Rejection rate trends over time
- Time to consent decision (how long do users wait before accepting/rejecting?)
- Impact of consent rejection on feature usage

---

## Page Views

Page views are tracked automatically whenever users navigate to a new page. These events capture:

- The page URL
- The page title
- Standard session information

No manual instrumentation is required for page view tracking.

**Metrics & stats ideas:**

- Most visited pages
- User navigation flows (common paths through the app)
- Bounce rate per page
- Average pages per session
- Entry and exit pages
- Time spent on each page
- Page views by time of day or day of week

---

## Common Analysis Queries

Here are some questions this data can help answer:

1. **Engagement**: Which challenges are most popular? (Count of `challenge_opened` by `challenge_id`)
2. **Quiz performance**: What's the average quiz score? (Mean of `score`/`max_score` from `quiz_completed`)
3. **Retention signals**: How many users enable push notifications? (`push_subscription_enabled` count)
4. **Feature adoption**: Are users exploring leaderboards? (`leaderboard_tab_changed` count)
5. **Accessibility**: Which languages are most used? (`language_changed` events + user traits)
