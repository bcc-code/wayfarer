# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for teams query with forward pagination using cursor. This tests:
- Forward pagination with first and after parameters
- Cursor-based pagination continuation
- Proper hasNextPage and hasPreviousPage values

## Query

```graphql
query GetTeamsPage($first: Int!, $after: String) {
  teams(first: $first, after: $after) {
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

## Variables

```json
{
  "first": 2
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
        }
      ],
      "pageInfo": {
        "endCursor": "VE0wMUs5UTNUVjVCUkhBQTMxN0M1NzI4UzZUVg==",
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
- first parameter limits page size correctly
- Cursors can be used to navigate through pages
- pageInfo.hasNextPage is true when more results exist
- pageInfo.hasPreviousPage is false on first page
