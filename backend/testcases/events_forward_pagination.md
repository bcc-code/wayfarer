# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for events query with forward pagination using cursor. This tests:
- Forward pagination with first and after parameters
- Cursor-based pagination continuation
- Proper hasNextPage and hasPreviousPage values

This is a multi-step test that should be split into two requests in practice:
1. First request: Get first page
2. Second request: Use endCursor from first page as after parameter

## Query

```graphql
query GetEventsPage($first: Int!, $after: String) {
  events(first: $first, after: $after) {
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
    "events": {
      "edges": [
        {
          "cursor": "RVYwMUs5REdTNTNOSEhGQVFaNjBDOE1HOU44TQ==",
          "node": {
            "description": "rerum eveniet id et ducimus est et aliquid ducimus nesciunt.",
            "id": "EV01K9DGS53NHHFAQZ60C8MG9N8M",
            "name": "dolorem Event"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNTZLQjNCSFMzU1kzRU1OM1EzWA==",
          "node": {
            "description": "in autem corporis aliquid at ut aut autem et tempore.",
            "id": "EV01K9DGS56KB3BHS3SY3EMN3Q3X",
            "name": "molestias Event"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "RVYwMUs5REdTNTZLQjNCSFMzU1kzRU1OM1EzWA==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs5REdTNTNOSEhGQVFaNjBDOE1HOU44TQ=="
      },
      "totalCount": 12
    }
  }
}
```

## Notes

This test verifies:
- first parameter limits page size correctly
- Cursors can be used to navigate through pages
- pageInfo.hasNextPage is true when more results exist
- pageInfo.hasPreviousPage is false on first page, true after using cursor

To test actual pagination:
1. Run with first: 2
2. Note the endCursor value
3. Run again with first: 2, after: <endCursor>
4. Verify hasNextPage and hasPreviousPage update correctly
