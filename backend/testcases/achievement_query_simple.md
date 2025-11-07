# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for querying a simple achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific simple achievement
- Achievement fields including id, name, description, image, points, and hidden
- Polymorphic type resolution (SimpleAchievement)

## Query

```graphql
query {
  achievement(id: "AC01K9DGTA7733PPCMDSYW22GXZ4") {
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
      "description": "quia ipsam et ullam possimus placeat possimus commodi enim qui.",
      "hidden": false,
      "id": "AC01K9DGTA7733PPCMDSYW22GXZ4",
      "image": "https://placecats.com/302/313",
      "name": "adipisci Achievement",
      "points": 55,
      "project": {
        "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
