# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for querying a reading achievement by ID using the achievement(id) query. This tests:
- achievement(id) to fetch a specific reading achievement
- ReadingAchievement type with articles array
- Achievement polymorphic type resolution

## Query

```graphql
query {
  achievement(id: "AC01K9DGTAJ116WSTKF6YY8GPBJZ") {
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
          "author": "Ayden Davis IV",
          "id": "RA01K9DGTAQYR75F2WKS90QP7AWD",
          "title": "in molestiae facere nostrum id.",
          "url": "http://jtw.com/eqqar-rq.html"
        },
        {
          "author": "Mr. Joany Jast",
          "id": "RA01K9DGTATWNNGV2C3Z650H7VV1",
          "title": "nihil sequi cumque vero quia.",
          "url": "https://www.yvk.com/om-ispty"
        },
        {
          "author": "Susana Nolan DVM",
          "id": "RA01K9DGTAWDZM6Y2DE8ETRDD7YQ",
          "title": "veniam sed magni id error.",
          "url": "http://htm.net/"
        },
        {
          "author": "Mr. Gabriel Lockman V",
          "id": "RA01K9DGTAXYGNYWTM5MYFQG63R1",
          "title": "placeat ut cumque quam quibusdam.",
          "url": "http://mah.biz/dabup-jzya.html"
        }
      ],
      "description": "Complete all articles to earn this achievement.",
      "hidden": false,
      "id": "AC01K9DGTAJ116WSTKF6YY8GPBJZ",
      "image": "https://placecats.com/g/378/330",
      "name": "Read: cum",
      "points": 91
    }
  }
}
```
