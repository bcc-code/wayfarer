# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for basic teams pagination without filters. This tests:
- teams query with default pagination (first 5)
- TeamConnection type with edges, pageInfo, and totalCount
- Cursor-based pagination structure

## Query

```graphql
query {
  teams(first: 5) {
    edges {
      cursor
      node {
        id
        name
        description
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
    "teams": {
      "edges": [
        {
          "cursor": "VE0wMUs5UTNUVEs1WDJOUFc4QUJRV1JHTTlWSw==",
          "node": {
            "description": "amet sint ab molestiae velit ex aut impedit doloribus ea.",
            "id": "TM01K9Q3TTK5X2NPW8ABQWRGM9VK",
            "name": "Signal Repairer OR Track Switch Repairer Team"
          }
        },
        {
          "cursor": "VE0wMUs5UTNUVjVCUkhBQTMxN0M1NzI4UzZUVg==",
          "node": {
            "description": "veritatis et rem error qui deleniti aut vel rerum natus.",
            "id": "TM01K9Q3TV5BRHAA317C5728S6TV",
            "name": "Hand Presser Team"
          }
        },
        {
          "cursor": "VE0wMUs5UTNUVlZNUURRU1hHWU5OMzVCWU5OUA==",
          "node": {
            "description": "omnis a debitis ipsum architecto sed dolores maiores quis qui.",
            "id": "TM01K9Q3TVVMQDQSXGYNN35BYNNP",
            "name": "Well and Core Drill Operator Team"
          }
        },
        {
          "cursor": "VE0wMUs5UTNUV0NSU0c4WUhQUUhEVkU4M1kyMQ==",
          "node": {
            "description": "ab nam occaecati eaque alias id quasi provident dolor quasi.",
            "id": "TM01K9Q3TWCRSG8YHPQHDVE83Y21",
            "name": "Communication Equipment Repairer Team"
          }
        },
        {
          "cursor": "VE0wMUs5UTNUWDE3OUdNN1IyMFlBOUZLTldOSA==",
          "node": {
            "description": "culpa doloribus sit sed facere quasi accusantium et voluptatem velit.",
            "id": "TM01K9Q3TX179GM7R20YA9FKNWNH",
            "name": "Forester Team"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "VE0wMUs5UTNUWDE3OUdNN1IyMFlBOUZLTldOSA==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "VE0wMUs5UTNUVEs1WDJOUFc4QUJRV1JHTTlWSw=="
      },
      "totalCount": 33
    }
  }
}
```

## Notes

This test verifies:
- TeamConnection structure matches the schema
- Cursors are properly encoded
- pageInfo fields correctly indicate pagination state
- totalCount reflects total number of teams
