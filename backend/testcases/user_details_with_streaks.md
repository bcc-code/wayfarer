# UserID: US01K9DGS4RXAJ40NF3XJNP7M9VQ

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
    "me": {
      "id": "US01K9DGS4RXAJ40NF3XJNP7M9VQ",
      "name": "Alphonso Shields",
      "projects": [
        {
          "myTeam": {
            "name": "Coremaking Machine Operator Team"
          }
        },
        {
          "myTeam": {
            "name": "Ship Captain Team"
          }
        }
      ]
    },
    "currentProject": {
      "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
      "name": "Summer Bible Camp 2025",
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "challenges": [
        {
          "name": "explicabo Challenge"
        },
        {
          "name": "impedit Challenge"
        },
        {
          "name": "eveniet Challenge"
        },
        {
          "name": "minima Challenge"
        },
        {
          "name": "nemo Challenge"
        },
        {
          "name": "nesciunt Challenge"
        },
        {
          "name": "praesentium Challenge"
        },
        {
          "name": "corporis Challenge"
        },
        {
          "name": "itaque Challenge"
        },
        {
          "name": "dolore Challenge"
        },
        {
          "name": "unde Challenge"
        },
        {
          "name": "rerum Challenge"
        },
        {
          "name": "hic Challenge"
        },
        {
          "name": "omnis Challenge"
        }
      ],
      "events": [
        {
          "name": "rerum Event"
        },
        {
          "name": "omnis Event"
        },
        {
          "name": "esse Event"
        },
        {
          "name": "molestias Event"
        },
        {
          "name": "dolorem Event"
        }
      ],
      "startDate": "2025-10-07T22:22:22+02:00",
      "endDate": "2026-01-05T22:22:22+01:00",
      "branding": {
        "colors": {
          "secondary": "#87B534"
        }
      },
      "teams": [
        {
          "name": "Buffing and Polishing Operator Team",
          "superTeam": null
        },
        {
          "name": "Metal Fabricator Team",
          "superTeam": {
            "name": "Hills-Hills Alliance"
          }
        },
        {
          "name": "Home Team",
          "superTeam": null
        },
        {
          "name": "Telecommunications Line Installer Team",
          "superTeam": null
        },
        {
          "name": "Plating Operator OR Coating Machine Operator Team",
          "superTeam": {
            "name": "Hills-Hills Alliance"
          }
        },
        {
          "name": "Home Appliance Repairer Team",
          "superTeam": {
            "name": "Hills-Hills Alliance"
          }
        },
        {
          "name": "Job Printer Team",
          "superTeam": {
            "name": "Koelpin and Sons Alliance"
          }
        },
        {
          "name": "Accountant Team",
          "superTeam": {
            "name": "Koelpin and Sons Alliance"
          }
        },
        {
          "name": "Coremaking Machine Operator Team",
          "superTeam": null
        }
      ],
      "myTeam": {
        "name": "Coremaking Machine Operator Team"
      },
      "streaks": [
        {
          "name": "dolorum Streak",
          "id": "SK01K9DGS5FT6YD9JTJP1HQX9CG2",
          "relevantDays": [],
          "listenedDays": [],
          "project": {
            "name": "Summer Bible Camp 2025"
          },
          "status": 0
        },
        {
          "name": "nulla Streak",
          "id": "SK01K9DGS5CSE0SSC0HS59325XPN",
          "relevantDays": [
            {
              "start": "2025-11-06",
              "end": "2025-11-09"
            }
          ],
          "listenedDays": [
            {
              "date": "2025-11-07",
              "active": true
            },
            {
              "date": "2025-11-06",
              "active": false
            }
          ],
          "project": {
            "name": "Summer Bible Camp 2025"
          },
          "status": 1
        }
      ]
    }
  }
}
```
