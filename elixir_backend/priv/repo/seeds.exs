# Script for populating the database. You can run it as:
#
#     mix run priv/repo/seeds.exs
#
# Idempotent: can be re-run safely (uses upsert/on_conflict where possible).

alias ElixirBackend.Repo
alias ElixirBackend.ULID

# ── 1. Churches ──

alias ElixirBackend.Churches.Church

churches =
  [
    %{id: "CH00000000000000000000OSLO", name: "Oslo Kirke", country: "NO", category: "XL"},
    %{id: "CH0000000000000000000BERGEN", name: "Bergen Kirke", country: "NO", category: "L"},
    %{id: "CH000000000000000000TROMSO", name: "Tromsø Kirke", country: "NO", category: "S"}
  ]
  |> Enum.map(fn attrs ->
    %Church{}
    |> Church.changeset(attrs)
    |> Repo.insert!(on_conflict: :nothing, conflict_target: :id)
  end)

[oslo_church, bergen_church, tromso_church] = churches
IO.puts("✓ Created #{length(churches)} churches")

# ── 2. Users ──

alias ElixirBackend.Accounts.User

users =
  [
    %{
      id: "US00000000000000000000ADMIN",
      name: "Admin User",
      members_id: "seed-admin-001",
      email: "admin@wayfarer.test",
      gender: "MALE",
      birthdate: ~D[1990-05-15],
      church_id: oslo_church.id,
      language: "no"
    },
    %{
      id: "US0000000000000000000ALICE",
      name: "Alice Hansen",
      members_id: "seed-alice-001",
      email: "alice@wayfarer.test",
      gender: "FEMALE",
      birthdate: ~D[2002-08-20],
      church_id: oslo_church.id,
      language: "en"
    },
    %{
      id: "US00000000000000000000BOB",
      name: "Bob Olsen",
      members_id: "seed-bob-001",
      email: "bob@wayfarer.test",
      gender: "MALE",
      birthdate: ~D[2001-03-10],
      church_id: bergen_church.id,
      language: "no"
    },
    %{
      id: "US0000000000000000000CAROL",
      name: "Carol Nilsen",
      members_id: "seed-carol-001",
      email: "carol@wayfarer.test",
      gender: "FEMALE",
      birthdate: ~D[2003-11-01],
      church_id: tromso_church.id,
      language: "en"
    }
  ]
  |> Enum.map(fn attrs ->
    %User{}
    |> User.create_changeset(attrs)
    |> Repo.insert!(on_conflict: :nothing, conflict_target: :id)
  end)

[admin_user, alice, bob, carol] = users
IO.puts("✓ Created #{length(users)} users")

# ── 3. Projects ──

alias ElixirBackend.Projects.Project

projects =
  [
    %{
      id: "PR000000000000000000BIBEL26",
      name: "Bibelcamp 2026",
      description: "Årets store bibelcamp med aktiviteter, quizer og konkurranser.",
      start_date: ~U[2026-01-01 00:00:00Z],
      end_date: ~U[2026-12-31 23:59:59Z],
      rules: "Vær snill og ha det gøy!",
      info_message: "Velkommen til Bibelcamp 2026!",
      info_message_start: ~U[2026-01-01 00:00:00Z],
      info_message_end: ~U[2026-01-31 23:59:59Z]
    },
    %{
      id: "PR00000000000000000SOMMER26",
      name: "Sommerleir 2026",
      description: "Sommerleir med fokus på fellesskap og bibel.",
      start_date: ~U[2026-06-01 00:00:00Z],
      end_date: ~U[2026-08-31 23:59:59Z]
    }
  ]
  |> Enum.map(fn attrs ->
    %Project{}
    |> Project.changeset(attrs)
    |> Repo.insert!(on_conflict: :nothing, conflict_target: :id)
  end)

[bibelcamp, sommerleir] = projects
IO.puts("✓ Created #{length(projects)} projects")

# ── 4. Events ──

alias ElixirBackend.Events.Event

events =
  [
    %{
      id: "EV00000000000000000HELG0126",
      name: "Januarhelg",
      description: "Kickoff-helg for Bibelcamp 2026.",
      start_date: ~U[2026-01-17 16:00:00Z],
      end_date: ~U[2026-01-19 14:00:00Z],
      project_id: bibelcamp.id
    },
    %{
      id: "EV00000000000000000HELG0226",
      name: "Påskehelg",
      description: "Påskeleir med quiz og aktiviteter.",
      start_date: ~U[2026-04-03 16:00:00Z],
      end_date: ~U[2026-04-06 14:00:00Z],
      project_id: bibelcamp.id
    },
    %{
      id: "EV0000000000000000SOMCAMP26",
      name: "Sommercamp",
      description: "Hovedsamlingen for sommerleiren.",
      start_date: ~U[2026-07-01 10:00:00Z],
      end_date: ~U[2026-07-07 16:00:00Z],
      project_id: sommerleir.id
    }
  ]
  |> Enum.map(fn attrs ->
    %Event{}
    |> Event.changeset(attrs)
    |> Repo.insert!(on_conflict: :nothing, conflict_target: :id)
  end)

[januarhelg, paskehelg, sommercamp] = events
IO.puts("✓ Created #{length(events)} events")

# ── 5. Streaks ──

alias ElixirBackend.Streaks

_streaks =
  [
    {:ok, streak1} =
      Streaks.create_streak(%{
        name: "Daglig bibellesing",
        description: "Les bibelen hver dag gjennom hele året.",
        project_id: bibelcamp.id,
        relevant_days: [
          %{start: ~D[2026-01-01], end: ~D[2026-06-30]},
          %{start: ~D[2026-07-01], end: ~D[2026-12-31]}
        ]
      }),
    {:ok, streak2} =
      Streaks.create_streak(%{
        name: "Sommerlesing",
        description: "Les hver dag gjennom sommerleiren.",
        project_id: sommerleir.id,
        relevant_days: [
          %{start: ~D[2026-06-01], end: ~D[2026-08-31]}
        ]
      })
  ]

daily_reading = streak1
summer_reading = streak2
IO.puts("✓ Created 2 streaks")

# ── 6. External Content ──

alias ElixirBackend.ExternalContent

contents =
  [
    {:ok, content1} =
      ExternalContent.upsert_content(%{
        id: "EC00000000000000000ARTICLE1",
        plan_id: "plan-bibelcamp-2026",
        task_id: "task-read-genesis",
        content_type: "ARTICLE",
        source: "ssf",
        url: "https://content.example.com/genesis-1"
      }),
    {:ok, content2} =
      ExternalContent.upsert_content(%{
        id: "EC00000000000000000ARTICLE2",
        plan_id: "plan-bibelcamp-2026",
        task_id: "task-read-exodus",
        content_type: "ARTICLE",
        source: "ssf",
        url: "https://content.example.com/exodus-1"
      }),
    {:ok, content3} =
      ExternalContent.upsert_content(%{
        id: "EC000000000000000000MEDIA1",
        plan_id: "plan-bibelcamp-2026",
        task_id: "task-listen-worship",
        content_type: "MEDIA",
        source: "ssf",
        url: "https://content.example.com/worship-song-1"
      })
  ]

article1 = content1
article2 = content2
_media1 = content3
IO.puts("✓ Created #{length(contents)} external content items")

# ── 7. Challenges ──

alias ElixirBackend.Challenges

challenges =
  [
    {:ok, _ch1} =
      Challenges.create_challenge(%{
        project_id: bibelcamp.id,
        challenge_type: "SIMPLE",
        name: "Del et bibelvers",
        description: "<p>Del ditt favoritt bibelvers med gruppen din.</p>",
        button_text: "Fullfør",
        notification_text: "Noen delte et bibelvers!",
        allow_self_completion: true,
        published_at: ~U[2026-01-01 00:00:00Z]
      }),
    {:ok, _ch2} =
      Challenges.create_challenge(%{
        project_id: bibelcamp.id,
        event_id: januarhelg.id,
        challenge_type: "SIMPLE",
        name: "Bli kjent-lek",
        description: "<p>Delta i bli kjent-leken under januarhelgen.</p>",
        button_text: "Jeg deltok!",
        notification_text: "Noen deltok i bli kjent-leken!",
        allow_self_completion: true,
        published_at: ~U[2026-01-17 16:00:00Z]
      }),
    {:ok, _ch3} =
      Challenges.create_challenge(%{
        project_id: bibelcamp.id,
        challenge_type: "EXTERNAL",
        name: "Les en artikkel",
        description: "<p>Les denne artikkelen om troens grunnlag.</p>",
        button_text: "Åpne artikkel",
        notification_text: "Noen leste artikkelen!",
        url: "https://content.example.com/faith-basics",
        published_at: ~U[2026-01-01 00:00:00Z]
      }),
    {:ok, _ch4} =
      Challenges.create_challenge(%{
        project_id: sommerleir.id,
        event_id: sommercamp.id,
        challenge_type: "SIMPLE",
        name: "Morgentrim",
        description: "<p>Delta i morgentrim kl 07:00.</p>",
        button_text: "Gjennomført",
        notification_text: "Noen var med på morgentrim!",
        allow_self_completion: true,
        published_at: ~U[2026-07-01 06:00:00Z]
      }),
    {:ok, _ch5} =
      Challenges.create_challenge(%{
        project_id: bibelcamp.id,
        challenge_type: "QUIZ",
        name: "Bibelquiz",
        description: "<p>Ta bibelquizen og test kunnskapen din!</p>",
        button_text: "Start quiz",
        notification_text: "Noen tok bibelquizen!",
        published_at: ~U[2026-01-01 00:00:00Z]
      })
  ]

[share_verse, bli_kjent, _read_article, _morning_exercise, quiz_challenge] =
  Enum.map(challenges, fn {:ok, c} -> c end)

IO.puts("✓ Created #{length(challenges)} challenges")

# ── 8. Super Teams & Teams ──

alias ElixirBackend.Teams

{:ok, super_team_north} =
  Teams.create_super_team(bibelcamp.id, %{
    name: "Nord",
    description: "Lag fra Nord-Norge",
    color: "#1E88E5"
  })

{:ok, super_team_south} =
  Teams.create_super_team(bibelcamp.id, %{
    name: "Sør",
    description: "Lag fra Sør-Norge",
    color: "#E53935"
  })

IO.puts("✓ Created 2 super teams")

{:ok, team_oslo} =
  Teams.create_team(bibelcamp.id, %{
    name: "Oslo Ungdom",
    description: "Ungdomsgruppen fra Oslo",
    super_team_id: super_team_south.id
  })

{:ok, team_bergen} =
  Teams.create_team(bibelcamp.id, %{
    name: "Bergen Ungdom",
    description: "Ungdomsgruppen fra Bergen",
    super_team_id: super_team_south.id
  })

{:ok, team_tromso} =
  Teams.create_team(bibelcamp.id, %{
    name: "Tromsø Ungdom",
    description: "Ungdomsgruppen fra Tromsø",
    super_team_id: super_team_north.id
  })

{:ok, team_sommer} =
  Teams.create_team(sommerleir.id, %{
    name: "Sommerteam Alpha",
    description: "Et av lagene på sommerleiren"
  })

IO.puts("✓ Created 4 teams")

# Add members to teams
Teams.add_members(team_oslo.id, [admin_user.id, alice.id])
Teams.add_members(team_bergen.id, [bob.id])
Teams.add_members(team_tromso.id, [carol.id])
Teams.add_members(team_sommer.id, [alice.id, bob.id])
IO.puts("✓ Assigned users to teams")

# ── 9. Quizzes ──

alias ElixirBackend.Quizzes

{:ok, bible_quiz} =
  Quizzes.create_quiz(%{
    name: "Bibelquiz: 1. Mosebok",
    description: "Test din kunnskap om Første Mosebok!",
    project_id: bibelcamp.id,
    challenge_id: quiz_challenge.id,
    timeout_seconds: 600,
    randomize_questions: true,
    reveal_correct_answers: true,
    allow_retakes: false,
    completion_points: 50
  })

{:ok, _q1} =
  Quizzes.add_question(bible_quiz.id, %{
    question_type: "PREDEFINED",
    question_text: "Hvem bygde arken?",
    question_order: 1,
    points: 10,
    betting_enabled: false,
    allow_multiple_selection: false,
    predefined_answers: [
      %{answer_text: "Noah", is_correct: true, answer_order: 0},
      %{answer_text: "Abraham", is_correct: false, answer_order: 1},
      %{answer_text: "Moses", is_correct: false, answer_order: 2},
      %{answer_text: "David", is_correct: false, answer_order: 3}
    ]
  })

{:ok, _q2} =
  Quizzes.add_question(bible_quiz.id, %{
    question_type: "PREDEFINED",
    question_text: "Hvor mange dager brukte Gud på skapelsen?",
    question_order: 2,
    points: 10,
    betting_enabled: false,
    allow_multiple_selection: false,
    predefined_answers: [
      %{answer_text: "5", is_correct: false, answer_order: 0},
      %{answer_text: "6", is_correct: true, answer_order: 1},
      %{answer_text: "7", is_correct: false, answer_order: 2},
      %{answer_text: "10", is_correct: false, answer_order: 3}
    ]
  })

{:ok, _q3} =
  Quizzes.add_question(bible_quiz.id, %{
    question_type: "FREE_TEXT",
    question_text: "Hva heter den første boken i Bibelen?",
    question_order: 3,
    points: 10,
    betting_enabled: false
  })

{:ok, _q4} =
  Quizzes.add_question(bible_quiz.id, %{
    question_type: "NUMBER",
    question_text: "Hvor mange sønner hadde Jakob?",
    question_order: 4,
    points: 15,
    betting_enabled: true,
    min_value: 1.0,
    max_value: 20.0,
    step_value: 1.0
  })

{:ok, _q5} =
  Quizzes.add_question(bible_quiz.id, %{
    question_type: "ORDERING",
    question_text: "Sett disse hendelsene i riktig rekkefølge:",
    question_order: 5,
    points: 20,
    betting_enabled: false,
    ordering_items: [
      %{item_text: "Skapelsen", correct_order: 1},
      %{item_text: "Syndefallet", correct_order: 2},
      %{item_text: "Noas ark", correct_order: 3},
      %{item_text: "Babels tårn", correct_order: 4}
    ]
  })

IO.puts("✓ Created 1 quiz with 5 questions")

# Second quiz for sommerleir
{:ok, summer_quiz} =
  Quizzes.create_quiz(%{
    name: "Sommerquiz",
    description: "En morsom quiz for sommerleiren.",
    project_id: sommerleir.id,
    timeout_seconds: 300,
    randomize_questions: false,
    reveal_correct_answers: true,
    allow_retakes: true,
    completion_points: 25
  })

{:ok, _sq1} =
  Quizzes.add_question(summer_quiz.id, %{
    question_type: "PREDEFINED",
    question_text: "Hva er Norges lengste elv?",
    question_order: 1,
    points: 10,
    betting_enabled: false,
    allow_multiple_selection: false,
    predefined_answers: [
      %{answer_text: "Glomma", is_correct: true, answer_order: 0},
      %{answer_text: "Drammenselva", is_correct: false, answer_order: 1},
      %{answer_text: "Tanaelva", is_correct: false, answer_order: 2}
    ]
  })

IO.puts("✓ Created 1 more quiz with 1 question")

# ── 10. Achievements ──

alias ElixirBackend.Achievements

placeholder_img = "https://placehold.co/200x200/EEE/999?text=Achievement"
placeholder_img_done = "https://placehold.co/200x200/4CAF50/FFF?text=Done"

{:ok, ach_simple} =
  Achievements.create_simple_achievement(%{
    name: "Velkommen!",
    description_pending: "Fullfør din første utfordring.",
    description_completed: "Du fullførte din første utfordring!",
    notification_text: "Noen fikk sin første achievement!",
    image_pending: placeholder_img,
    image_completed: placeholder_img_done,
    project_id: bibelcamp.id,
    points: 50,
    hidden: false,
    challenge_id: share_verse.id
  })

{:ok, _ach_event} =
  Achievements.create_simple_achievement(%{
    name: "Januarhelg-deltaker",
    description_pending: "Delta på januarhelgen.",
    description_completed: "Du var med på januarhelgen!",
    notification_text: "Noen deltok på januarhelgen!",
    image_pending: placeholder_img,
    image_completed: placeholder_img_done,
    project_id: bibelcamp.id,
    event_id: januarhelg.id,
    points: 100,
    hidden: false,
    challenge_id: bli_kjent.id
  })

{:ok, _ach_hidden} =
  Achievements.create_simple_achievement(%{
    name: "Hemmelig achievement",
    description_pending: "???",
    description_completed: "Du fant den hemmelige achievementen!",
    notification_text: "Noen fant den hemmelige achievementen!",
    image_pending: placeholder_img,
    image_completed: placeholder_img_done,
    project_id: bibelcamp.id,
    points: 200,
    hidden: true
  })

{:ok, _ach_content} =
  Achievements.create_content_achievement(%{
    name: "Bibelleser",
    description_pending: "Les alle artiklene.",
    description_completed: "Du har lest alle artiklene!",
    notification_text: "Noen fullførte bibelleser-achievementen!",
    image_pending: placeholder_img,
    image_completed: placeholder_img_done,
    project_id: bibelcamp.id,
    points: 150,
    hidden: false,
    items: [
      %{external_content_id: article1.id, sort_order: 0},
      %{external_content_id: article2.id, sort_order: 1}
    ]
  })

{:ok, _ach_streak} =
  Achievements.create_streak_achievement(%{
    name: "7-dagers streak",
    description_pending: "Les bibelen 7 dager på rad.",
    description_completed: "Du har lest 7 dager på rad!",
    notification_text: "Noen oppnådde 7-dagers streak!",
    image_pending: placeholder_img,
    image_completed: placeholder_img_done,
    project_id: bibelcamp.id,
    points: 100,
    hidden: false,
    streak_id: daily_reading.id,
    needed_streak: 7
  })

{:ok, _ach_quiz} =
  Achievements.create_quiz_achievement(%{
    name: "Quizmester",
    description_pending: "Få minst 80% på bibelquizen.",
    description_completed: "Du er en ekte quizmester!",
    notification_text: "Noen ble quizmester!",
    image_pending: placeholder_img,
    image_completed: placeholder_img_done,
    project_id: bibelcamp.id,
    points: 100,
    hidden: false,
    quiz_id: bible_quiz.id,
    min_score_percentage: 80,
    require_completion: true
  })

IO.puts("✓ Created 6 achievements (simple, event, hidden, content, streak, quiz)")

# ── 11. Consents ──

alias ElixirBackend.Consents

{:ok, privacy_consent} =
  Consents.create_consent(%{
    key: "privacy_policy",
    version: 1,
    title: "Personvernserklæring",
    short_text: "Vi behandler dine personopplysninger trygt.",
    body: """
    <h2>Personvernserklæring</h2>
    <p>Vi samler inn og behandler personopplysninger i henhold til GDPR.</p>
    <p>Dine data brukes kun for å gi deg den beste opplevelsen på plattformen.</p>
    """,
    published_at: ~U[2026-01-01 00:00:00Z]
  })

{:ok, terms_consent} =
  Consents.create_consent(%{
    key: "terms_of_service",
    version: 1,
    title: "Brukervilkår",
    short_text: "Aksepter brukervilkårene for å bruke tjenesten.",
    body: """
    <h2>Brukervilkår</h2>
    <p>Ved å bruke denne tjenesten aksepterer du følgende vilkår...</p>
    """,
    published_at: ~U[2026-01-01 00:00:00Z]
  })

IO.puts("✓ Created 2 consents")

# ── 12. User Participation ──

alias ElixirBackend.Accounts

# All users join bibelcamp
for user <- [admin_user, alice, bob, carol] do
  Accounts.assign_user_to_project(user.id, bibelcamp.id)
end

# Alice and Bob also join sommerleir
Accounts.assign_user_to_project(alice.id, sommerleir.id)
Accounts.assign_user_to_project(bob.id, sommerleir.id)

# Event participation
Accounts.assign_user_to_event(admin_user.id, januarhelg.id)
Accounts.assign_user_to_event(alice.id, januarhelg.id)
Accounts.assign_user_to_event(bob.id, paskehelg.id)
Accounts.assign_user_to_event(alice.id, sommercamp.id)
Accounts.assign_user_to_event(bob.id, sommercamp.id)

IO.puts("✓ Assigned users to projects and events")

# ── 13. Roles ──

alias ElixirBackend.Roles

Roles.assign_role(%{
  user_id: admin_user.id,
  role: "SUPERADMIN",
  assigned_by: admin_user.id
})

Roles.assign_role(%{
  user_id: alice.id,
  role: "PROJECT_ADMIN",
  scope_type: "PROJECT",
  scope_id: bibelcamp.id,
  assigned_by: admin_user.id
})

Roles.assign_role(%{
  user_id: bob.id,
  role: "TEAM_LEAD",
  scope_type: "TEAM",
  scope_id: team_bergen.id,
  assigned_by: admin_user.id
})

IO.puts("✓ Assigned roles")

# ── 14. Consent Acceptance ──

Consents.accept_consent(admin_user.id, privacy_consent.id)
Consents.accept_consent(admin_user.id, terms_consent.id)
Consents.accept_consent(alice.id, privacy_consent.id)
Consents.accept_consent(alice.id, terms_consent.id)
Consents.accept_consent(bob.id, privacy_consent.id)
# Bob hasn't accepted terms yet — will show as pending

IO.puts("✓ Recorded consent responses")

# ── 15. Streak Activity ──

# Alice has been reading for 5 days
for day_offset <- 0..4 do
  date = Date.add(Date.utc_today(), -day_offset)
  Streaks.record_activity(daily_reading.id, alice.id, date)
end

# Bob read for 3 days, skipped one, read again
for day_offset <- [0, 1, 2, 4, 5] do
  date = Date.add(Date.utc_today(), -day_offset)
  Streaks.record_activity(daily_reading.id, bob.id, date)
end

IO.puts("✓ Recorded streak activity")

# ── 16. Challenge Completions ──

alias ElixirBackend.Challenges

Challenges.complete_challenge(alice.id, share_verse.id)
Challenges.complete_challenge(bob.id, share_verse.id)
Challenges.complete_challenge(alice.id, bli_kjent.id)

IO.puts("✓ Recorded challenge completions")

# ── 17. Quiz Sessions ──

{:ok, _session} =
  Quizzes.create_session(bible_quiz.id, %{
    name: "Januarhelg-runde",
    state: "OPEN",
    created_by: admin_user.id,
    open_at: ~U[2026-01-18 10:00:00Z],
    lock_at: ~U[2026-01-18 11:00:00Z]
  })

IO.puts("✓ Created quiz session")

# ── 18. Translations ──

alias ElixirBackend.Translations

# Project translations (English)
Translations.upsert_translation(:project, %{
  project_id: bibelcamp.id,
  language_code: "en",
  name: "Bible Camp 2026",
  description: "This year's big Bible camp with activities, quizzes, and competitions.",
  rules: "Be kind and have fun!"
})

Translations.upsert_translation(:project, %{
  project_id: sommerleir.id,
  language_code: "en",
  name: "Summer Camp 2026",
  description: "Summer camp focusing on community and the Bible."
})

# Event translations
Translations.upsert_translation(:event, %{
  event_id: januarhelg.id,
  language_code: "en",
  name: "January Weekend",
  description: "Kickoff weekend for Bible Camp 2026."
})

Translations.upsert_translation(:event, %{
  event_id: paskehelg.id,
  language_code: "en",
  name: "Easter Weekend",
  description: "Easter camp with quizzes and activities."
})

Translations.upsert_translation(:event, %{
  event_id: sommercamp.id,
  language_code: "en",
  name: "Summer Camp",
  description: "Main gathering for the summer camp."
})

# Streak translations
Translations.upsert_translation(:streak, %{
  streak_id: daily_reading.id,
  language_code: "en",
  name: "Daily Bible reading",
  description: "Read the Bible every day throughout the year."
})

Translations.upsert_translation(:streak, %{
  streak_id: summer_reading.id,
  language_code: "en",
  name: "Summer reading",
  description: "Read every day throughout the summer camp."
})

# Challenge translations
Translations.upsert_translation(:challenge, %{
  challenge_id: share_verse.id,
  language_code: "en",
  name: "Share a Bible verse",
  description: "<p>Share your favorite Bible verse with your group.</p>",
  button_text: "Complete",
  notification_text: "Someone shared a Bible verse!"
})

# Achievement translations
Translations.upsert_translation(:achievement, %{
  achievement_id: ach_simple.id,
  language_code: "en",
  name: "Welcome!",
  description_pending: "Complete your first challenge.",
  description_completed: "You completed your first challenge!",
  notification_text: "Someone got their first achievement!"
})

# Consent translations
Translations.upsert_translation(:consent, %{
  consent_id: privacy_consent.id,
  language_code: "en",
  title: "Privacy Policy",
  short_text: "We handle your personal data safely.",
  body: """
  <h2>Privacy Policy</h2>
  <p>We collect and process personal data in accordance with GDPR.</p>
  <p>Your data is only used to give you the best experience on the platform.</p>
  """
})

Translations.upsert_translation(:consent, %{
  consent_id: terms_consent.id,
  language_code: "en",
  title: "Terms of Service",
  short_text: "Accept the terms of service to use the platform.",
  body: """
  <h2>Terms of Service</h2>
  <p>By using this service you accept the following terms...</p>
  """
})

# Quiz translations
Translations.upsert_translation(:quiz, %{
  quiz_id: bible_quiz.id,
  language_code: "en",
  name: "Bible Quiz: Genesis",
  description: "Test your knowledge of the Book of Genesis!"
})

IO.puts("✓ Created translations (en)")

IO.puts("")
IO.puts("🌱 Seeding complete!")
IO.puts("   Churches: 3")
IO.puts("   Users: 4 (admin, alice, bob, carol)")
IO.puts("   Projects: 2")
IO.puts("   Events: 3")
IO.puts("   Streaks: 2")
IO.puts("   External Content: 3")
IO.puts("   Challenges: 5 (simple, event, external, quiz, sommerleir)")
IO.puts("   Super Teams: 2, Teams: 4")
IO.puts("   Quizzes: 2 (6 questions total)")
IO.puts("   Achievements: 6 (simple, event, hidden, content, streak, quiz)")
IO.puts("   Consents: 2")
IO.puts("   Translations: 11 (English)")
