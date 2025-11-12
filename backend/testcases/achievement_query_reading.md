# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

## Description

Test for querying a reading achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific reading achievement
- ReadingAchievement type with articles array
- Achievement polymorphic type resolution

## Query

```graphql
query {
  achievement(id: "AC01K9Q3VEY4007WZ2WSA3Z7B2BH") {
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
      "articles": [
        {
          "author": "Mr. Niko Goodwin V",
          "id": "RA01K9Q3VF1EDWDDGSWAH3NQ3SSE",
          "title": "nisi quisquam esse voluptatem odit.",
          "url": "http://www.ryz.net/gdmz-mwzvwe"
        },
        {
          "author": "Ernestine Collier",
          "id": "RA01K9Q3VF34VW38765GG4ZR7X2Y",
          "title": "veritatis fugiat qui tempora voluptatum.",
          "url": "http://ltv.com/xjjti-n"
        },
        {
          "author": "Leopoldo Lubowitz",
          "id": "RA01K9Q3VF42FG9T7127V3ECPE65",
          "title": "voluptas quia blanditiis voluptas voluptates.",
          "url": "http://vws.com/"
        },
        {
          "author": "Ms. Tianna Schiller",
          "id": "RA01K9Q3VF4WHY80RFC28T5ZM53T",
          "title": "ratione accusamus voluptatum voluptas velit.",
          "url": "https://sum.net/qo-gvtizr.html"
        },
        {
          "author": "Meta Zboncak V",
          "id": "RA01K9Q3VF5SBY2E98KS5131B7SD",
          "title": "fugit quia et ratione ullam.",
          "url": "http://yyk.org/glpa-xcwxjo.html"
        },
        {
          "author": "Aglae Schinner DDS",
          "id": "RA01K9Q3VF6JXY6RPWD4B9CAV53G",
          "title": "veniam quos esse quod fugiat.",
          "url": "http://www.fou.net/awum-gbih"
        }
      ],
      "description": "Complete all articles to earn this achievement.",
      "hidden": false,
      "id": "AC01K9Q3VEY4007WZ2WSA3Z7B2BH",
      "image": "https://placecats.com/g/345/343",
      "name": "Read: reprehenderit",
      "points": 131
    }
  }
}
```
