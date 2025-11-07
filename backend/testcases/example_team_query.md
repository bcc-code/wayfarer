# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

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
  project(id: "PR01K9DGS50S1RZSE5HGN8JQ1XDC") {
    name
    description
  }
  team(id: "TM01K9DGS69XTYTDE9NW5CZ222WB") {
    id
    name
    description
    parentProject {
      id
      name
    }
  }
  event(id: "EV01K9DGS584AMYYY4G2D130FWVC") {
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
      "description": "sed nostrum odit tempore illum molestias est est ut voluptas.",
      "id": "EV01K9DGS584AMYYY4G2D130FWVC",
      "name": "esse Event"
    },
    "me": {
      "id": "US01K9DGS18D92WBMV3X7ETHNPMN",
      "name": "Alaina King"
    },
    "project": {
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "name": "Summer Bible Camp 2025"
    },
    "team": {
      "description": "vel voluptas commodi non qui consequatur voluptate libero illo asperiores.",
      "id": "TM01K9DGS69XTYTDE9NW5CZ222WB",
      "name": "Coremaking Machine Operator Team",
      "parentProject": {
        "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
