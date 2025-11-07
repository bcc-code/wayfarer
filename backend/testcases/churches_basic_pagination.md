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
          "cursor": "Q0gwMUs5REdTMFYxQkJSR1lKWkRYUFpRM1hZWg==",
          "node": {
            "category": "L",
            "country": "Norway",
            "id": "CH01K9DGS0V1BBRGYJZDXPZQ3XYZ",
            "name": "Oslo Community Church"
          }
        },
        {
          "cursor": "Q0gwMUs5REdTMFhaUVFRQlFaUDNHMzVOMTlWMQ==",
          "node": {
            "category": "S",
            "country": "Norway",
            "id": "CH01K9DGS0XZQQQBQZP3G35N19V1",
            "name": "Bergen Fellowship"
          }
        },
        {
          "cursor": "Q0gwMUs5REdTMFpEMU02VzBHUDVRWUVKNUVSQQ==",
          "node": {
            "category": "S",
            "country": "Norway",
            "id": "CH01K9DGS0ZD1M6W0GP5QYEJ5ERA",
            "name": "Stavanger Baptist Church"
          }
        },
        {
          "cursor": "Q0gwMUs5REdTMTBZMktYRzJDTldUNERSMVpDRw==",
          "node": {
            "category": "XL",
            "country": "Sweden",
            "id": "CH01K9DGS10Y2KXG2CNWT4DR1ZCG",
            "name": "Stockholm Evangelical Church"
          }
        },
        {
          "cursor": "Q0gwMUs5REdTMTJEODdLMzBRNlhFWFZaMTlKQw==",
          "node": {
            "category": "L",
            "country": "Sweden",
            "id": "CH01K9DGS12D87K30Q6XEXVZ19JC",
            "name": "Göteborg Christian Center"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "Q0gwMUs5REdTMTJEODdLMzBRNlhFWFZaMTlKQw==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "Q0gwMUs5REdTMFYxQkJSR1lKWkRYUFpRM1hZWg=="
      },
      "totalCount": 9
    }
  }
}
```
