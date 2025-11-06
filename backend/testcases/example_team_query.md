# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
  project(id: "PR01K8XV6J9H7BAEV49ZFVYS8R1K") {
    name
    description
  }
  team(id: "TM01K8XV6VK9ED2GBZSQ2VDTAT8T") {
    id
    name
    description
    parentProject {
      id
      name
    }
  }
  event(id: "EV01K8XV6JHBAJNYHJXRYDYHD4WK") {
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
    "me": {
      "id": "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
      "name": "Matjaz"
    },
    "project": {
      "name": "Summer Bible Camp 2025",
      "description": "Join us for an amazing summer adventure exploring God's word!"
    },
    "team": {
      "id": "TM01K8XV6VK9ED2GBZSQ2VDTAT8T",
      "name": "Butcher Team",
      "description": "voluptatem quos similique odit accusantium velit unde praesentium consequuntur ut.",
      "parentProject": {
        "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
        "name": "Summer Bible Camp 2025"
      }
    },
    "event": {
      "id": "EV01K8XV6JHBAJNYHJXRYDYHD4WK",
      "name": "voluptatem Event",
      "description": "laboriosam rerum qui expedita enim unde ex et provident pariatur."
    }
  }
}
```
