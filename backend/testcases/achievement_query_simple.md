# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for querying a simple achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific simple achievement
- Achievement fields including id, name, description, image, points, and hidden
- Polymorphic type resolution (SimpleAchievement)

## Query

```graphql
query {
  achievement(id: "AC01K8XV7SF42C5MH49A7EKDZ64P") {
    id
    name
    description
    image
    points
    hidden
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
    "achievement": {
      "id": "AC01K8XV7SF42C5MH49A7EKDZ64P",
      "name": "ratione Achievement",
      "description": "iste delectus quidem corrupti numquam consectetur rerum dolor rerum dolor.",
      "image": "https://placecats.com/377/343",
      "points": 5,
      "hidden": false,
      "project": {
        "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
