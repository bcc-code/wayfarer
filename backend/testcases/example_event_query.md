# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for querying multiple entities including event(id). This tests:
- me query to get current user
- project(id) to fetch a specific project
- event(id) to fetch a specific event
- user(id) to fetch another user with their projects

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
    startDate
    endDate
  }

  event(id:"EV01K9Q3TSXDMF2FX52XCGAAFDE5") {
    id
    description
    name
    startDate
    endDate
  }

  user(id: "US01K9Q3TQYGR8W5JHW4GMVPWS44") {
    id
    name
    projects{
      id
    }
  }
}
```

## Expected

```json
{
  "data": {
    "event": {
      "description": "accusantium modi neque ex itaque est eligendi deleniti a dolore.",
      "endDate": "2025-10-14T15:48:29+02:00",
      "id": "EV01K9Q3TSXDMF2FX52XCGAAFDE5",
      "name": "eius Event",
      "startDate": "2025-10-11T15:48:29+02:00"
    },
    "me": {
      "id": "US01K9Q3TQYGR8W5JHW4GMVPWS44",
      "name": "Art Fritsch"
    },
    "project": {
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "endDate": "2026-01-09T15:48:29+01:00",
      "name": "Summer Bible Camp 2025",
      "startDate": "2025-10-11T15:48:29+02:00"
    },
    "user": {
      "id": "US01K9Q3TQYGR8W5JHW4GMVPWS44",
      "name": "Art Fritsch",
      "projects": [
        {
          "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP"
        }
      ]
    }
  }
}
```
