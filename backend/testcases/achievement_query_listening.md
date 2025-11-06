# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for querying a listening achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific listening achievement
- ListeningAchievement type with tracks array
- Achievement polymorphic type resolution

## Query

```graphql
query {
  achievement(id: "AC01K8XV7WJNR51CH185GSHVZCEC") {
    id
    name
    description
    image
    points
    hidden
    ... on ListeningAchievement {
      tracks {
        id
        name
        description
        image
      }
    }
  }
}
```

## Expected

```json
{
  "data": {
    "achievement": {
      "id": "AC01K8XV7WJNR51CH185GSHVZCEC",
      "name": "Listen: perferendis",
      "description": "Listen to all tracks to earn this achievement.",
      "image": "https://placecats.com/neo/389/388",
      "points": 80,
      "hidden": false,
      "tracks": [
        {
          "id": "LT01K8XV7WNX7KRVNHN1KVNV3P5K",
          "name": "ea eligendi autem.",
          "description": "voluptatibus est magni occaecati corrupti est quaerat incidunt.",
          "image": "https://placecats.com/millie/395/356"
        },
        {
          "id": "LT01K8XV7WQJF949R6SN454G763V",
          "name": "fugit quo impedit.",
          "description": "placeat ea aut eum provident unde earum veritatis.",
          "image": "https://placecats.com/millie/348/308"
        },
        {
          "id": "LT01K8XV7WS7HM510NRC529AV9AW",
          "name": "dolores vitae laboriosam.",
          "description": "et commodi alias rerum eligendi optio necessitatibus qui.",
          "image": "https://placecats.com/millie/303/353"
        },
        {
          "id": "LT01K8XV7WTY1VND6QS7H39KMJM1",
          "name": "et debitis officia.",
          "description": "mollitia et doloremque expedita explicabo eos optio blanditiis.",
          "image": "https://placecats.com/millie/361/315"
        }
      ]
    }
  }
}
```
