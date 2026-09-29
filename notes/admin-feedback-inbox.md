# Admin UI feedback — inbox

Feedback on the admin interface from three testers (2026-09-29), with a focus on
the quiz admin. The raw, verbatim feedback is preserved at the bottom; the
sections above are the consolidated and deduplicated version we work from.

Related notes:

- [`admin-ux-improvements.md`](./admin-ux-improvements.md) — per-page UX tracker
  (structure, conventions, scope decisions). Anything we decide to fix should be
  carried over there.
- [`quiz-system-implementation.md`](./quiz-system-implementation.md) — how the
  quiz system works
- [`quiz-results-admin.md`](./quiz-results-admin.md) — the results/analytics view

Severity: `!!` correctness/data problem · `!` real friction · `~` nice-to-have
Source: `P1`, `P2`, `P3` = tester 1/2/3 (see raw feedback); `muntlig` =
relayed in conversation, not written down below.

## Decisions (2026-09-29)

Settled with Sigve; these drive the verdicts below.

1. **Audience: project / study organizers.** Not local church admins. Only a
   handful of people will use it for now, so polish is not the goal — but it
   must be intuitive enough that it does not block them from setting up a
   project on their own.
2. **Quiz retakes should not award points.** Agreed in principle, but it is a
   bigger product decision and is deliberately deferred — see #19. Until then,
   treat "Tillat brukere å ta quiz på nytt" as granting points again on every
   attempt.
3. **Hide what is not relevant today, explain what remains.** Unused fields come
   out of the UI; every field that stays gets a help text a non-developer can
   read.

**Hiding means hiding in the UI only** — the columns, GraphQL inputs and
resolvers stay as they are. Anything we hide still needs a sane value sent on
create (event `null`, publish time `now`, ordering default), so hide the input,
keep the payload.

## Findings

### Cross-cutting

| #   | Sev | Item                                                                                                                                    | Source  |
| --- | --- | --------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| 1   | `!` | Signing in lands on the normal user app (and into the bible study); had to retype `/admin` in the URL to get back                       | P1      |
| 2   | `!` | Poor contrast on checkbox borders — probably applies to all input types                                                                 | P3      |
| 3   | `!` | No confirmation when navigating away with unsaved changes                                                                               | P3      |
| 4   | `~` | Clickable things need clearer affordance: button sizes, colors on interactive elements, proper number-picker fields                     | P2      |
| 5   | `!` | Help texts need a pass overall — they must be understandable without knowing the data model (see #10, #11, #15, #16, #24 for specifics) | P2, P3  |
| 25  | `!` | Content is labelled "Navn"; it should be "Tittel" — challenges, quizzes and the like                                                    | muntlig |

Overall verdict from P2: "OK+ UI, bra nok for et adminpanel behind the scenes",
just not as friendly as it could be. Nothing in the feedback says the admin is
hard to get through — P2 got the job done quickly.

### Challenge

| #   | Sev | Item                                                                                                                                                                                              | Source |
| --- | --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 6   | —   | Creating a new challenge is easy to find — keep that                                                                                                                                              | P1     |
| 7   | `!` | Default challenge type should be Quiz. Today it is `Simple` (`challenges/new.vue:52`); the other types read as unfamiliar to a new admin                                                          | P1, P3 |
| 8   | `!` | Each challenge type needs a short explanation in the type selector                                                                                                                                | P1, P3 |
| 9   | `~` | Hide the event ("arrangement") selector — barely used so far                                                                                                                                      | P3     |
| 10  | `!` | Publishing time: help text "Når utfordringen blir synlig for brukere. Før dette ser bare påmeldte den." — "påmeldte" is not understood. P3 suggests hiding it and defaulting to now               | P1, P3 |
| 11  | `!` | Notification help text "Tekst som vises i push-varsler når admin melder bruker på utfordringen. La feltet stå tomt for ingen varsling." — "admin melder bruker på utfordringen" is not understood | P1     |

### Quiz — editor

| #   | Sev | Item                                                                                                                                                                                                                                                                | Source  |
| --- | --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| 12  | `!` | After creating a quiz, the landing page does not make the next step obvious. It opens on the quiz name you just entered (editable, but not what you came for), while description and image appear to have "disappeared". The page should lead with adding questions | P1      |
| 13  | `~` | Hide quiz title, description and image — not in use today (same page as #12; hiding them is one way to fix it)                                                                                                                                                      | P3      |
| 14  | `!` | "Legg til svaralternativ" should sit below the list of alternatives, not above it                                                                                                                                                                                   | P3      |
| 15  | `!` | Question type needs a help text                                                                                                                                                                                                                                     | P3      |
| 16  | `!` | Question points need a help text                                                                                                                                                                                                                                    | P3      |
| 26  | `~` | Hide "Tilfeldig spørsmålsrekkefølge" until the backend honours it — **premise was wrong, see the log**                                                                                                                                                              | muntlig |
| 17  | `!` | The question editor modal must not be closeable by accident — losing a half-written question to a stray click or Esc                                                                                                                                                | P3      |
| 18  | `~` | Per-question images, not just a quiz thumbnail — enables "who is this?" style questions and visual answer alternatives                                                                                                                                              | P2      |

### Quiz — behaviour

| #   | Sev  | Item                                                                                                                                                                                                                                                                                                                                                                     | Source |
| --- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| 19  | `!!` | "Tillat brukere å ta quiz på nytt" — does a retake award points again? **Verified: yes.** Each completed submission writes a new `score_journal` entry (`quizzes.resolvers.go:1271`–`1322`) and there is no per-user/per-quiz dedupe, so every retake adds `completionPoints` + answer points again. P1 wants retakes, but wants to be sure points are not granted twice | P1     |

### Achievement

| #   | Sev | Item                                                                                                                                      | Source |
| --- | --- | ----------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 20  | `!` | After creating an achievement you should land on the achievement list. Today it goes to the project overview (`achievements/new.vue:115`) | P3     |

### Superteam

| #   | Sev | Item                                         | Source |
| --- | --- | -------------------------------------------- | ------ |
| 21  | `~` | Superteam description is not relevant — hide | P3     |
| 22  | `!` | Superteam color field is not intuitive       | P3     |

### Leaderboard

| #   | Sev | Item                                                     | Source |
| --- | --- | -------------------------------------------------------- | ------ |
| 23  | `~` | Hide the event selector and the ordering field           | P3     |
| 24  | `!` | Better label and help text on the limit field ("Topp X") | P3     |

## Status

Verdict: `fix` · `later` · `wontfix` · `needs-discussion`. "Plan" is what we
intend to do; rows marked **done** have landed.

| #   | Verdict   | Plan                                                                                                                         |
| --- | --------- | ---------------------------------------------------------------------------------------------------------------------------- |
| 1   | `fix`     | Redirect admins to /admin after sign-in                                                                                      |
| 2   | `fix`     | Input border contrast, app-wide                                                                                              |
| 3   | `fix`     | Unsaved-changes guard on admin forms                                                                                         |
| 4   | `later`   | Polish — revisit once the blocking items are done                                                                            |
| 5   | `fix`     | Help-text pass; specifics in #10, #11, #15, #16, #24                                                                         |
| 7   | `fix`     | Default challenge type → Quiz — **done**                                                                                     |
| 8   | `fix`     | Explain each challenge type in the selector — **done**                                                                       |
| 9   | `fix`     | Hide event selector (decision 3) — **done**                                                                                  |
| 10  | `fix`     | Publishing time removed; "Synlig fra" replaced by a visibility choice — **done**                                             |
| 11  | `fix`     | Reword notification help text — **done**                                                                                     |
| 12  | `fix`     | Quiz page leads with questions; settings moved below — **done**                                                              |
| 13  | `fix`     | Title, description and image removed from the quiz form; inherited from the challenge and carried through on save — **done** |
| 26  | `wontfix` | The setting works — the warning saying otherwise was wrong and is gone; the checkbox now explains it                         |
| 14  | `fix`     | "Legg til svaralternativ" moved under the list (same for ordering items) — **done**                                          |
| 15  | `fix`     | Help text on question type, with the grading rule per type — **done**                                                        |
| 16  | `fix`     | Points field hidden for the types that cannot earn any; help text on the rest — **done**                                     |
| 17  | `fix`     | Question dialog is no longer dismissible by click-outside or Esc — **done**                                                  |
| 18  | `later`   | Needs backend work; no one is blocked on it                                                                                  |
| 19  | `later`   | Product decision deferred (decision 2) — write it up separately                                                              |
| 20  | `fix`     | Redirect to achievement list after create                                                                                    |
| 21  | `fix`     | Hide superteam description (decision 3)                                                                                      |
| 22  | `fix`     | Rework superteam color field                                                                                                 |
| 23  | `fix`     | Hide leaderboard event + ordering (decision 3)                                                                               |
| 24  | `fix`     | Relabel limit field + help text                                                                                              |

## Log

- 2026-09-29 — file created
- 2026-09-29 — feedback from three testers added and consolidated into #1–#24;
  #19 verified against the backend
- 2026-09-29 — open decisions settled; verdicts assigned. 20 items to fix,
  4 deferred (#4, #18, #19)
- 2026-09-29 — batch 1 (challenge creation: #7–#11) implemented. Found and
  fixed on the way: a new challenge was created with an empty `visible_at`,
  which the API reads as "only enrolled users ever see it" — new challenges are
  now visible to everyone unless the organiser says otherwise, and the choice is
  explicit in the form. Also fixed: the create page sent `endTime`/`visibleAt`
  as bare local time, which the API parses as UTC
- 2026-09-29 — batch 2 (quiz editor: #12–#17) implemented. While writing the
  help texts, confirmed against the backend that only Flervalg and Rekkefølge
  are graded automatically — free text and number answers are stored with
  `is_correct` NULL and therefore never earn their question's points. The type
  and points help texts now say so
- 2026-09-29 — copy fixes on the quiz settings: completion points now say they
  are given to everyone who finishes regardless of score, and #26 hides the
  randomise switch rather than explaining that it does nothing
- 2026-09-29 — #26 reverted after checking the backend. `randomize_questions`
  **is** honoured: `StartQuizSession` shuffles the question order into the
  submission (`quiz_sessions.resolvers.go:756`), `QuizSubmission.orderedQuestions`
  returns that stored order (`quizzes.resolvers.go:2813`), and the user app
  renders exactly that list (`QuizChallenge.vue:233`). Only the M2M
  `CreateQuizSubmission` deliberately keeps the canonical order. The admin form's
  old "Ikke i bruk ennå" warning was simply wrong; the checkbox is back with a
  description of when the order is drawn. Whether any existing quiz has it
  enabled is a database question, not answerable from the code
- 2026-09-29 — question type copy rewritten in plain language, and the points
  field is now hidden for Fritekst and Tall, which no grading path can score.
  Any points value left over from an earlier type choice is dropped on save.
  **Follow-up worth its own item:** `StartQuizSession` sums `points` over every
  question when it computes `max_score` (`quiz_sessions.resolvers.go:768`),
  with no type filter, while the M2M `CreateQuizSubmission` filters by type
  (`quizzes.resolvers.go:1506`) — and includes Number, which is not graded
  either. Existing quizzes with points on an ungradable question therefore show
  participants a max score they cannot reach
- 2026-09-29 — **bug found and fixed while testing the points field**: editing
  an existing question discarded every change. `AdminQuizQuestionEditor`
  rebuilt the question without its `localKey`, and `AdminQuizForm.saveQuestion`
  matches on exactly that key, so the edit matched nothing and the untouched
  original stayed in the list — no error, and the dialog closed as if it had
  worked. Predates this week's work; the old `points: 1` default hid part of it
  because a value was always sent. Covered by a regression test
- 2026-09-29 — #25: project content ("Navn" → "Tittel") on the challenge, quiz,
  achievement and leaderboard forms. People and groups — teams, superteams,
  users, churches — and quiz sessions keep "Navn"

---

# Raw feedback (verbatim)

## Feedback person 1

- Når jeg logget inn kom jeg tilbake til vanlig interact og inn i bibelstudie. Jeg måtte igjen bytte url tilbake til /admin for å komme inn. (Litt upraktisk)
- Vil forsikre meg - dette er admin for noen som vil sette opp og manage et bibelstudie ikkesant? Ikke en lokal admin. For ellers bør vi nok se på å redusere valg muligheter.
- Lett å komme til å opprette en ny utfordring! Bra! Men standard utfordringstype bør være Quiz. De andre kategoriene er litt fremmede for meg.
- Når jeg skal sette publiseringstidspunkt kommer det en tekst jeg ikke forstår helt "Når utfordringen blir synlig for brukere. Før dette ser bare påmeldte den." Hva betyr påmeldte?
- Under Varsling står det også en tekst som ikke er helt klar for meg: "Tekst som vises i push-varsler når admin melder bruker på utfordringen. La feltet stå tomt for ingen varsling." - Hva menes med at admin melder en bruker på utfordring?
- Når jeg opprettet Quiz kommer jeg til en side det tok lang tid å skjønne hva jeg skulle gjøre på. Jeg blir møtt med Quiz navn som jeg allerede hadde gitt. Skjønte etterhvert at jeg kunne redigere navnet på den hvis jeg ville. Men Det er ikke der jeg er nå. Jeg tenker mest på å lage spørsmål. Dessuten var beskrivelse og bilde "forsvunnet"
- Tillat brukere å ta quiz på nytt. - Spm får de da mulighet til å få poeng på nytt? Det bør klargjøres. Fint om man kan ta quiz på nytt. Men jeg som admin bør være trygg på at ikke poeng gis dobbelt.

## Feedback person 2

Synes det var OK+ UI, bra nok for et adminpanel behind the scenes, men kunne også vært noe mer brukervennlig (hjelpetekst, knappstørrelser, farger på klikkbare ting, tallvelge-felt for å nevne noen)

Gikk jo forholdsvis fort og greit, så har ikke så mye mer enn det å kommentere på😊

Annet enn det kunne det vært interessant å se på mulighet for å legge inn bilde i hvert spørsmål i en quiz (ikke bare som en thumbnail til quizen). Det utvider bruksområdet litt mer hvor man kan stille et spørsmål "hvem er dette" feks. og kan bruke svaralternativene til å ikke bare referere til oppgavetekst, men også noe visuelt.

Bare en idé 💡

## Feedback person 3

### Generelt

- Dårlig kontrast på checkbox borders (gjelder kanskje alle typer inputs)
- Bereftelse før man navigerer uten å lagre

### Challenge

- Skjul arrangement (ikke mye brukt foreløpig)
- Quiz by default
- Forklaringer på challenge type
- Skjul publiseringstidspunk kanskje? Burde basically være "Date.now"

### Quiz

- Skjul tittel, description og bilde (ikke i bruk per nå)
- Finskrive hjelpetekster. Hjelpetekster må være enkle å forstå
- Spørsmålstype hjelpetekst
- "Legg til svaralternativ" burde ligge under alternativer
- Spørsmål poeng hjelpetekst
- Spørsmål editor må ikke kunne lukkes uheldig (modal)

### Achievement

- Burde blitt tatt videre til liste etter å ha opprettet utmerkelse

### Superteam

- Superteam description er ikke relevant
- Superteam farge felt ikke intuitivt

### Leaderboard

- Skjul arrangement, skjul rekkefølge
- Bedre label og hjelpetekst på limit (topp X)
