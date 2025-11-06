# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
  project(id: "PR01K8XV6J9H7BAEV49ZFVYS8R1K") {
    name
    description
    startDate
    endDate
  }

  event(id:"EV01K8XV6JHBAJNYHJXRYDYHD4WK") {
    id
    description
    name
    startDate
    endDate
  }

  user(id: "US01K8XV6J7WAC9WSFTPP376NPSD") {
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
    "me": {
      "id": "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
      "name": "Matjaz"
    },
    "project": {
      "name": "Summer Bible Camp 2025",
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "startDate": "2025-10-01T20:16:36+02:00",
      "endDate": "2025-12-30T20:16:36+01:00"
    },
    "event": {
      "id": "EV01K8XV6JHBAJNYHJXRYDYHD4WK",
      "description": "laboriosam rerum qui expedita enim unde ex et provident pariatur.",
      "name": "voluptatem Event",
      "startDate": "2025-10-15T20:16:36+02:00",
      "endDate": "2025-10-18T20:16:36+02:00"
    },
    "user": {
      "id": "US01K8XV6J7WAC9WSFTPP376NPSD",
      "name": "Austin Parisian",
      "projects": [
        {
          "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K"
        },
        {
          "id": "PR01K8XV6JTX5GYJH9WNP4BN2P2A"
        }
      ]
    }
  }
}
```
