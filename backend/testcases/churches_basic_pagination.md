# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test basic pagination for the `churches` query without filters.

## Query

```graphql
query {
  churches(first: 5) {
    edges {
      cursor
      node {
        id
        name
        country
        category
      }
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    totalCount
  }
}
```

## Expected

```json
{
  "data": {
    "churches": {
      "edges": [
        {
          "cursor": "Q0gwMUs5UTNUUU4xNjk5RzQ1MUdQNEcwV1FQRw==",
          "node": {
            "category": "L",
            "country": "Norway",
            "id": "CH01K9Q3TQN1699G451GP4G0WQPG",
            "name": "Oslo Community Church"
          }
        },
        {
          "cursor": "Q0gwMUs5UTNUUVBSNTI4VEJLNEdGRTNCUTNBWQ==",
          "node": {
            "category": "S",
            "country": "Norway",
            "id": "CH01K9Q3TQPR528TBK4GFE3BQ3AY",
            "name": "Bergen Fellowship"
          }
        },
        {
          "cursor": "Q0gwMUs5UTNUUVFKU1hENjJENUtaM1hZTUFCNA==",
          "node": {
            "category": "S",
            "country": "Norway",
            "id": "CH01K9Q3TQQJSXD62D5KZ3XYMAB4",
            "name": "Stavanger Baptist Church"
          }
        },
        {
          "cursor": "Q0gwMUs5UTNUUVJDMUo2MllDMDE1VlZEOVRRUA==",
          "node": {
            "category": "XL",
            "country": "Sweden",
            "id": "CH01K9Q3TQRC1J62YC015VVD9TQP",
            "name": "Stockholm Evangelical Church"
          }
        },
        {
          "cursor": "Q0gwMUs5UTNUUVM4WUdZOVhINjhRSDhBVlM0MA==",
          "node": {
            "category": "L",
            "country": "Sweden",
            "id": "CH01K9Q3TQS8YGY9XH68QH8AVS40",
            "name": "Göteborg Christian Center"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "Q0gwMUs5UTNUUVM4WUdZOVhINjhRSDhBVlM0MA==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "Q0gwMUs5UTNUUU4xNjk5RzQ1MUdQNEcwV1FQRw=="
      },
      "totalCount": 8
    }
  }
}
```
