# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for querying a listening achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific listening achievement
- ListeningAchievement type with tracks array
- Achievement polymorphic type resolution

## Query

```graphql
query {
  achievement(id: "AC01K9Q3VFC9P5ZWCGBTMQAR08P4") {
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
      "description": "Listen to all tracks to earn this achievement.",
      "hidden": false,
      "id": "AC01K9Q3VFC9P5ZWCGBTMQAR08P4",
      "image": "https://placecats.com/neo/395/319",
      "name": "Listen: dolorem",
      "points": 108,
      "tracks": [
        {
          "description": "adipisci nihil dolorum quas ea ea quasi corporis.",
          "id": "LT01K9Q3VFFPZDKXNY5CNMGMB2P2",
          "image": "https://placecats.com/millie/397/304",
          "name": "aut officiis rerum."
        },
        {
          "description": "magnam dolorem sit aut velit deleniti quod nihil.",
          "id": "LT01K9Q3VFHDSHV1VTYY4QY3ZDDW",
          "image": "https://placecats.com/millie/369/373",
          "name": "ut ipsa tenetur."
        },
        {
          "description": "quisquam neque animi id odit odit architecto sed.",
          "id": "LT01K9Q3VFJ7Z90TX0VAX27J2S2Q",
          "image": "https://placecats.com/millie/322/321",
          "name": "adipisci qui dolores."
        },
        {
          "description": "aspernatur aut enim incidunt expedita rerum ipsa dolorem.",
          "id": "LT01K9Q3VFK278VRFNEPEDSFXMYN",
          "image": "https://placecats.com/millie/371/395",
          "name": "ipsa repellendus eius."
        },
        {
          "description": "quam neque voluptas placeat earum fugit sit sint.",
          "id": "LT01K9Q3VFKXC8DAWAQR7DBEWST2",
          "image": "https://placecats.com/millie/337/363",
          "name": "atque deserunt corporis."
        },
        {
          "description": "animi qui quos commodi quo eum id aperiam.",
          "id": "LT01K9Q3VFMRWSW61BTT8XC62BCF",
          "image": "https://placecats.com/millie/365/399",
          "name": "non voluptas culpa."
        },
        {
          "description": "nihil iste dolorum impedit quisquam fugit omnis nobis.",
          "id": "LT01K9Q3VFNHDBS5CGF2K9BCCAC2",
          "image": "https://placecats.com/millie/386/332",
          "name": "omnis sit perferendis."
        },
        {
          "description": "omnis neque qui nesciunt in officia et natus.",
          "id": "LT01K9Q3VFPCW0NFPFFPK05VNGSS",
          "image": "https://placecats.com/millie/353/321",
          "name": "totam possimus debitis."
        }
      ]
    }
  }
}
```
