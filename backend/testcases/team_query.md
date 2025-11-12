# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for querying a team by ID using the team(id) query. This tests:
- team(id) to fetch a specific team
- Team fields including id, name, description, and parentProject

## Query

```graphql
query {
  team(id: "TM01K9Q3TTK5X2NPW8ABQWRGM9VK") {
    id
    name
    description
    parentProject {
      id
      name
    }
  }
}
```

## Expected

```json
{
  "data": {
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
