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

**Caveat found later:** there is no real component to reuse here — a push
notification is drawn by the OS, not by the app — so this one is a mock by
necessity, the only place on the list where the replica rule does not apply.

`notificationText` on both the challenge and the achievement form is written
blind, and it is the one field that reaches a participant's lock screen. A mock
notification — app name, title, body, truncated where the OS truncates — is
cheap and needs no user components at all.

### 4. Project rules and info message

**Caveat found later:** `ProjectInfoBanner` takes `{ markdown, html }` and the
HTML comes from the backend (`MarkdownText.html`). A preview of _unsaved_
markdown therefore needs either a round trip or a client-side renderer, and no
markdown library is installed. Previewing the last-saved version is free; the
live draft is not.

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

### 6. Project branding

`AdminProjectThemeEditor` already previews the palette as swatches on a mock
screen. Rendering a real screen in the chosen colours — the challenge card and
a standings row, which is most of what a participant looks at — would answer
"does this palette work" rather than "what are these colours". The pieces are
in place: both components are already used in previews elsewhere, and
`AdminThemedPreview` is exactly the frame for it.

### Probably not worth it

- **Events** — a name and two dates; there is nothing to see.
- **Teams** — same.
- **Consents** — the body is markdown like the project rules, but a consent is
  written once per project and read by legal, not designed.

## Decided: the real components, not replicas (2026-09-29)

Every preview renders the participant app's own components, imported through
`#layers/user/app/...`. What stays in the admin layer is the **adapter**: the
form holds a draft, the components are typed against the query that feeds the
app, and the preview maps one to the other.

`AdminThemedPreview` is a width and the project's palette, nothing else. The
390px cap is the viewport the user layer's breakpoints are drawn for, so the
content is judged at a width a participant will actually see.

It deliberately paints no surface. A phone mockup with a bezel was tried and
reverted — at a real phone's aspect ratio the frame is mostly empty box — and
so was the plain bordered frame that replaced it: a challenge card already
carries its own rounded surface, and a frame around it is just two borders. A
preview that _is_ a screen rather than a card (the quiz question, the
achievement) paints its own background.

Three things this needs, learned building the first two:

- `AdminThemedPreview` is `inert`. The real components carry real links and
  buttons; without it, clicking a preview navigates the admin into the
  participant app, and tabbing strands the keyboard inside a picture. The one
  opt-out (`interactive`) exists for `AdminAchievementPreview`, which still
  keeps a state switcher inside the screen, and goes away with it.
- A draft has no ids. The quiz preview gives answers and ordering items
  positional ones, because the components key on them.
- Shared markup that is not a component yet has to become one. The question
  heading lived inline in `QuizChallenge`; it is now
  `QuizQuestionHeading`, used by both. Extracting beats copying — the copy is
  what this note was written to stop.

**Done.** Every preview on the list below is built, except the project rules
and the superteam (the latter set aside as not needed yet):

| Preview           | Where                   | Real component                          |
| ----------------- | ----------------------- | --------------------------------------- |
| Challenge card    | challenge form          | `ChallengeCard`                         |
| Quiz question     | question dialog         | `Quiz*Question` + `QuizQuestionHeading` |
| Leaderboard       | leaderboard edit page   | `StandingsBoard`                        |
| Achievement       | achievement form        | `AchievementDetails`                    |
| Project home page | project settings        | `ProfileProjectCard`                    |
| Push notification | challenge + achievement | none — see below                        |

Two of them needed a component extracted first, the same move both times: the
markup existed but was inline. `QuizQuestionHeading` came out of
`QuizChallenge`, and `AchievementDetails` out of `AchievementBadge` — the badge
keeps the drawer, the analytics and the confetti, and the contents are now
shared. That also retired `AdminThemedPreview`'s `interactive` escape hatch:
the achievement's pending/completed switcher moved out to the form, where
preview chrome belongs.

Two limits worth knowing:

- **The leaderboard preview shows the saved configuration.** The server
  computes a board from the stored filter and there is no query that takes an
  unsaved one, so edits appear after saving. The label above the preview says
  so. A `previewLeaderboard(filter:)` query would fix it; that is backend work.
- **The push notification is a mock.** A notification is drawn by the OS, so
  there is no component to reuse. It mirrors the payload the backend sends —
  title from the name, body from the notification text — and clamps to two
  lines as a phone does.

The project settings preview is fed from the **draft**, not from what is
stored, so a banner and a palette can be judged before saving. That was the
point of the branding idea and it came free with the home-page preview.
