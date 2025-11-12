# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for querying multiple entities including team(id). This tests:
- me query to get current user
- project(id) to fetch a specific project
- team(id) to fetch a specific team
- event(id) to fetch a specific event

## Query

```graphql
query {
  me {
    id
    name
  }
  project(id: "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP") {
    name
    description
  }
  team(id: "TM01K9Q3TTK5X2NPW8ABQWRGM9VK") {
    id
    name
    description
    parentProject {
      id
      name
    }
  }
  event(id: "EV01K9Q3TSXDMF2FX52XCGAAFDE5") {
    id
    name
    description
  }
}
```

## Expected

```json
{
  "data": {
    "event": {
      "description": "accusantium modi neque ex itaque est eligendi deleniti a dolore.",
      "id": "EV01K9Q3TSXDMF2FX52XCGAAFDE5",
      "name": "eius Event"
    },
    "me": {
      "id": "US01K9Q3TQYGR8W5JHW4GMVPWS44",
      "name": "Art Fritsch"
    },
    "project": {
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "name": "Summer Bible Camp 2025"
    },
    "team": {
      "description": "amet sint ab molestiae velit ex aut impedit doloribus ea.",
      "id": "TM01K9Q3TTK5X2NPW8ABQWRGM9VK",
      "name": "Signal Repairer OR Track Switch Repairer Team",
      "parentProject": {
        "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
