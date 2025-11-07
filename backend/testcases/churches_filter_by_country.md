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
            "id": "CH01K9DGS10Y2KXG2CNWT4DR1ZCG",
            "name": "Stockholm Evangelical Church"
          }
        },
        {
          "node": {
            "category": "L",
            "country": "Sweden",
            "id": "CH01K9DGS12D87K30Q6XEXVZ19JC",
            "name": "Göteborg Christian Center"
          }
        }
      ],
      "totalCount": 2
    }
  }
}
```
