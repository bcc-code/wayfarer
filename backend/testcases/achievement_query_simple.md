# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for querying a simple achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific simple achievement
- Achievement fields including id, name, description, image, points, and hidden
- Polymorphic type resolution (SimpleAchievement)

## Query

```graphql
query {
  achievement(id: "AC01K9Q3VERA16BQWMZT6BRDPXB5") {
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
      "description": "praesentium ut est possimus similique aut qui ratione voluptatum ea.",
      "hidden": false,
      "id": "AC01K9Q3VERA16BQWMZT6BRDPXB5",
      "image": "https://placecats.com/371/306",
      "name": "repudiandae Achievement",
      "points": 100,
      "project": {
        "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
