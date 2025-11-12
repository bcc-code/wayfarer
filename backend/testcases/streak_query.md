# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for the `streak(id)` query. This should return a streak by its ID with nested project resolution.

## Query

```graphql
query {
  streak(id: "SK01K9Q3TT2G1V3SN647X7QGJ7D9") {
    id
    name
    description
    project {
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
    "streak": {
      "description": "Maintain your streak by completing activities daily!",
      "id": "SK01K9Q3TT2G1V3SN647X7QGJ7D9",
      "name": "voluptas Streak",
      "project": {
        "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
