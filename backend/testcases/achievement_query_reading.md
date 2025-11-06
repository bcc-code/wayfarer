# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for querying a reading achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific reading achievement
- ReadingAchievement type with articles array
- Achievement polymorphic type resolution

## Query

```graphql
query {
  achievement(id: "AC01K8XV7TABCMXCXSE3PJCT33HQ") {
    id
    name
    description
    image
    points
    hidden
    ... on ReadingAchievement {
      articles {
        id
        title
        author
        url
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
      "id": "AC01K8XV7TABCMXCXSE3PJCT33HQ",
      "name": "Read: ut",
      "description": "Complete all articles to earn this achievement.",
      "image": "https://placecats.com/g/399/341",
      "points": 88,
      "hidden": false,
      "articles": [
        {
          "id": "RA01K8XV7TDKJ07B66KS6RZH73WB",
          "title": "necessitatibus voluptatem tenetur facere iusto.",
          "author": "Mitchel McClure",
          "url": "http://www.xkq.com/"
        },
        {
          "id": "RA01K8XV7TF9JQKEB2AZWGWD4T66",
          "title": "molestiae aperiam facere et eos.",
          "author": "Braden Wunsch I",
          "url": "http://fzp.com/dpzh-xnqxpy"
        },
        {
          "id": "RA01K8XV7TGX66CBH66PQF7WCNBP",
          "title": "aut nulla distinctio est necessitatibus.",
          "author": "Mr. Anibal Yundt Jr.",
          "url": "http://hoc.net/y-bweon"
        },
        {
          "id": "RA01K8XV7TJKWQPHE1ERXTXE9NJC",
          "title": "magnam dolores nulla quia labore.",
          "author": "Madalyn Pollich",
          "url": "http://svv.org/sv-smfvh.html"
        },
        {
          "id": "RA01K8XV7TM7W0VKP2H43RY4R5AJ",
          "title": "quos et quos omnis magnam.",
          "author": "Vincenzo Powlowski",
          "url": "http://muw.com/dp-jyo"
        },
        {
          "id": "RA01K8XV7TNVGDACQVVK69HC6ESR",
          "title": "asperiores qui necessitatibus et voluptatem.",
          "author": "Reta Grady",
          "url": "http://lan.biz/"
        }
      ]
    }
  }
}
```
