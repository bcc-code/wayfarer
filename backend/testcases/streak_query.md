# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for the `streak(id)` query. This should return a streak by its ID with nested project resolution.

## Query

```graphql
query {
  streak(id: "SK01K9DGS5S4CEFDC3XQC2HQ7HGW") {
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
      "id": "SK01K9DGS5S4CEFDC3XQC2HQ7HGW",
      "name": "similique Streak",
      "description": "Maintain your streak by completing activities daily!",
      "project": {
        "id": "PR01K9DGS5HA9SWE7T8FCJ975JZT",
        "name": "Youth Winter Retreat 2024"
      }
    }
  }
}
```
