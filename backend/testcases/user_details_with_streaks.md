# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for comprehensive user details query including current project with streaks showing activity status. This verifies that:
- User details are correctly resolved
- Project relationships work correctly
- Streak status calculation shows correct values
- ListenedDays only includes dates from today backwards (no future dates)
- Active status is correctly marked based on user activity

## Query

```graphql
query userDetails {
  me {
    id
    name
    projects {
      myTeam {
        name
      }
    }
  }
  currentProject {
    id
    name
    description
    challenges {
      name
    }
    events {
      name
    }
    startDate
    endDate
    branding {
      colors {
        secondary
      }

    }
    teams {
      name
      superTeam
      {
        name
      }
    }
    myTeam {
      name
    }
    streaks {
      name
      id
      relevantDays {
        start
        end
      }
      listenedDays(last: 5) {
        date
        active
      }
      project {
        name
      }
      status
    }
  }
}
```

## Expected

```json
{
  "data": {
    "currentProject": {
      "branding": {
        "colors": {
          "secondary": "#FFF1EB"
        }
      },
      "challenges": [
        {
          "name": "nam Challenge"
        },
        {
          "name": "non Challenge"
        },
        {
          "name": "alias Challenge"
        },
        {
          "name": "fugiat Challenge"
        },
        {
          "name": "aut Challenge"
        },
        {
          "name": "aperiam Challenge"
        },
        {
          "name": "assumenda Challenge"
        },
        {
          "name": "perferendis Challenge"
        },
        {
          "name": "accusantium Challenge"
        },
        {
          "name": "ut Challenge"
        },
        {
          "name": "inventore Challenge"
        },
        {
          "name": "veniam Challenge"
        },
        {
          "name": "earum Challenge"
        },
        {
          "name": "sint Challenge"
        },
        {
          "name": "nisi Challenge"
        },
        {
          "name": "blanditiis Challenge"
        },
        {
          "name": "suscipit Challenge"
        }
      ],
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "endDate": "2026-01-09T15:48:29+01:00",
      "events": [
        {
          "name": "nihil Event"
        },
        {
          "name": "id Event"
        },
        {
          "name": "quos Event"
        },
        {
          "name": "animi Event"
        },
        {
          "name": "eius Event"
        }
      ],
      "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP",
      "myTeam": {
        "name": "Semiconductor Processor Team"
      },
      "name": "Summer Bible Camp 2025",
      "startDate": "2025-10-11T15:48:29+02:00",
      "streaks": [
        {
          "id": "SK01K9Q3TT46W57DX96ARTPETEAB",
          "listenedDays": [],
          "name": "et Streak",
          "project": {
            "name": "Summer Bible Camp 2025"
          },
          "relevantDays": [],
          "status": 0
        },
        {
          "id": "SK01K9Q3TT2G1V3SN647X7QGJ7D9",
          "listenedDays": [],
          "name": "voluptas Streak",
          "project": {
            "name": "Summer Bible Camp 2025"
          },
          "relevantDays": [],
          "status": 0
        }
      ],
      "teams": [
        {
          "name": "Parts Salesperson Team",
          "superTeam": {
            "name": "Kovacek Group Alliance"
          }
        },
        {
          "name": "Park Naturalist Team",
          "superTeam": null
        },
        {
          "name": "Hand Trimmer Team",
          "superTeam": null
        },
        {
          "name": "Fish Game Warden Team",
          "superTeam": null
        },
        {
          "name": "Semiconductor Processor Team",
          "superTeam": null
        },
        {
          "name": "Electrician Team",
          "superTeam": {
            "name": "Kovacek Group Alliance"
          }
        },
        {
          "name": "Forester Team",
          "superTeam": {
            "name": "Kovacek Group Alliance"
          }
        },
        {
          "name": "Communication Equipment Repairer Team",
          "superTeam": null
        },
        {
          "name": "Well and Core Drill Operator Team",
          "superTeam": null
        },
        {
          "name": "Hand Presser Team",
          "superTeam": null
        },
        {
          "name": "Signal Repairer OR Track Switch Repairer Team",
          "superTeam": null
        }
      ]
    },
    "me": {
      "id": "US01K9Q3TQYGR8W5JHW4GMVPWS44",
      "name": "Art Fritsch",
      "projects": [
        {
          "myTeam": {
            "name": "Semiconductor Processor Team"
          }
        }
      ]
    }
  }
}
```
