# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for querying a specific superteam using superteam(id). This tests:
- superteam(id) to fetch a specific super team
- Basic fields: id, name, description
- parentProject relation

## Query

```graphql
query {
  superteam(id: "ST01K9Q3TTFVTYEM5WBFZM98G349") {
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
    "superteam": {
      "description": "doloremque totam optio et saepe sed omnis voluptas.",
      "id": "ST01K9Q3TTFVTYEM5WBFZM98G349",
      "name": "Kuvalis, Kuvalis and Kuvalis Alliance",
      "parentProject": {
        "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
