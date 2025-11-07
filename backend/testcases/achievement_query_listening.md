# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for querying a listening achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific listening achievement
- ListeningAchievement type with tracks array
- Achievement polymorphic type resolution

## Query

```graphql
query {
  achievement(id: "AC01K9DGTB8S9CJCSJFPD0X8Q86M") {
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
      "id": "AC01K9DGTB8S9CJCSJFPD0X8Q86M",
      "image": "https://placecats.com/neo/357/374",
      "name": "Listen: harum",
      "points": 71,
      "tracks": [
        {
          "description": "atque ea laborum nihil eaque possimus quis culpa.",
          "id": "LT01K9DGTBEN77ZAMA3JMAK4SHXV",
          "image": "https://placecats.com/millie/316/386",
          "name": "eum et consequuntur."
        },
        {
          "description": "in culpa voluptatem velit magnam non id voluptas.",
          "id": "LT01K9DGTBHKAMFGZY8V5V4TFSYH",
          "image": "https://placecats.com/millie/339/342",
          "name": "et autem repudiandae."
        },
        {
          "description": "iure molestiae et omnis labore sit voluptates quos.",
          "id": "LT01K9DGTBK64WNWPYWVP0744SQM",
          "image": "https://placecats.com/millie/345/342",
          "name": "animi dolor dicta."
        },
        {
          "description": "in vitae occaecati ut rerum autem voluptas expedita.",
          "id": "LT01K9DGTBMSPYBR4HTG65PZS2R0",
          "image": "https://placecats.com/millie/333/389",
          "name": "tempora aspernatur voluptatum."
        },
        {
          "description": "ut esse voluptatibus aliquid rerum velit dolorem suscipit.",
          "id": "LT01K9DGTBPBXZCVRH1RZQ8YSK15",
          "image": "https://placecats.com/millie/372/313",
          "name": "quasi explicabo voluptate."
        },
        {
          "description": "quibusdam quia error unde ut qui ea et.",
          "id": "LT01K9DGTBQWWATE0JNA2BHRYD0Z",
          "image": "https://placecats.com/millie/350/362",
          "name": "aut quo ut."
        },
        {
          "description": "sunt autem incidunt nulla et ad alias sint.",
          "id": "LT01K9DGTBSD1N99BART3PN7YE7A",
          "image": "https://placecats.com/millie/326/398",
          "name": "nostrum aliquam corrupti."
        },
        {
          "description": "perspiciatis molestias reiciendis voluptas consectetur voluptatem explicabo aut.",
          "id": "LT01K9DGTBV01T5GPAP81ZF5ABNN",
          "image": "https://placecats.com/millie/355/354",
          "name": "voluptatibus velit repellat."
        }
      ]
    }
  }
}
```
