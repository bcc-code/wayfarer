# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

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
  project(id: "PR01K9DGS50S1RZSE5HGN8JQ1XDC") {
    name
    description
    startDate
    endDate
  }

  event(id:"EV01K9DGS584AMYYY4G2D130FWVC") {
    id
    description
    name
    startDate
    endDate
  }

  user(id: "US01K9DGS18D92WBMV3X7ETHNPMN") {
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
      "description": "sed nostrum odit tempore illum molestias est est ut voluptas.",
      "endDate": "2025-10-24T22:22:22+02:00",
      "id": "EV01K9DGS584AMYYY4G2D130FWVC",
      "name": "esse Event",
      "startDate": "2025-10-21T22:22:22+02:00"
    },
    "me": {
      "id": "US01K9DGS18D92WBMV3X7ETHNPMN",
      "name": "Alaina King"
    },
    "project": {
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "endDate": "2026-01-05T22:22:22+01:00",
      "name": "Summer Bible Camp 2025",
      "startDate": "2025-10-07T22:22:22+02:00"
    },
    "user": {
      "id": "US01K9DGS18D92WBMV3X7ETHNPMN",
      "name": "Alaina King",
      "projects": [
        {
          "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC"
        },
        {
          "id": "PR01K9DGS5W604EQNCS8V4EG3V3Z"
        }
      ]
    }
  }
}
```
