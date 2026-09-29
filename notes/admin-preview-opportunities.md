# Preview in the admin panel — where it would earn its place

Ideas note, 2026-09-29. Not a plan; nothing here is scheduled.

The admin panel already previews two things: a challenge card
(`AdminChallengeCardPreview`) and an achievement (`AdminAchievementPreview`),
both inside `AdminThemedPreview`, which paints the project's own colours behind
them. `AdminProjectThemePreview` shows the palette on a mock screen.

The value in all three is the same: the organiser is writing content for an app
they are not looking at, and the form can only tell them what they typed.

## Worth doing, roughly in order

### 1. The quiz question editor

The strongest case. An organiser writes a question in a dialog of form fields
and finds out what it looks like when a participant opens it — after publishing.
Nothing in the dialog shows how long alternatives wrap, what the ordering
question looks like once dragged, or what the betting module adds to a question.

The user layer already has the real components:
`layers/user/app/components/challenges/quiz/questions/*` plus
`QuizAlternative`, `QuizBettingModule` and `QuizProgress`.

### 2. The leaderboard config

The filter fields answer "what did I set", never "who shows up". A preview
rendering `LeaderboardList` against the real board — the configured entity type,
the limit, the filter applied — turns a page of abstract criteria into the list
it produces. This one needs data, not just layout: it is a query against the
configured filter rather than a render of form state.

### 3. Push notification text

`notificationText` on both the challenge and the achievement form is written
blind, and it is the one field that reaches a participant's lock screen. A mock
notification — app name, title, body, truncated where the OS truncates — is
cheap and needs no user components at all.

### 4. Project rules and info message

Both are markdown, rendered to HTML by the backend (`MarkdownText.html`) and
shown in a drawer (`ProjectRules`). The editor shows the markdown; the drawer
shows headings, lists and links. A preview closes that gap, and unlike the
others it can render exactly what the app will, since the HTML comes from the
same resolver.

### 5. Superteam colour and image

A superteam's colour only means something next to the other superteams. A
preview of the standings row, or the superteam card, is what makes an organiser
pick a colour that is distinguishable rather than one that looks nice in a
picker. Related: [`admin-feedback-inbox.md`](./admin-feedback-inbox.md) #22.

### Probably not worth it

- **Events** — a name and two dates; there is nothing to see.
- **Teams** — same.
- **Consents** — the body is markdown like the project rules, but a consent is
  written once per project and read by legal, not designed.

## One thing to decide first

The two previews that exist are **replicas**: `AdminChallengeCardPreview` is the
admin layer's own rendering of a challenge card, not the user layer's
`ChallengeCard`. That is a second copy of the design to keep in step, and it is
already how the panel drifted from the app once.

`frontend/CLAUDE.md` permits the other approach — the admin layer may import
from `layers/user` for exactly this reason, and nothing does today. Deciding
between "replica" and "the real component in a themed frame" before building
more previews is worth more than any single preview on the list, because every
one of them multiplies the choice.
