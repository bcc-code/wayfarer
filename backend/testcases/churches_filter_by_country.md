# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test filtering churches by country.

## Query

```graphql
query {
  churches(filter: { country: "Sweden" }) {
    edges {
      node {
        id
        name
        country
        category
      }
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
          "node": {
            "category": "XL",
            "country": "Sweden",
            "id": "CH01K9Q3TQRC1J62YC015VVD9TQP",
            "name": "Stockholm Evangelical Church"
          }
        },
        {
          "node": {
            "category": "L",
            "country": "Sweden",
            "id": "CH01K9Q3TQS8YGY9XH68QH8AVS40",
            "name": "Göteborg Christian Center"
          }
        }
      ],
      "totalCount": 2
    }
  }
}
```
