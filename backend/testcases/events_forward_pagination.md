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
          "cursor": "RVYwMUs5UTNUU1hETUYyRlg1MlhDR0FBRkRFNQ==",
          "node": {
            "description": "accusantium modi neque ex itaque est eligendi deleniti a dolore.",
            "id": "EV01K9Q3TSXDMF2FX52XCGAAFDE5",
            "name": "eius Event"
          }
        },
        {
          "cursor": "RVYwMUs5UTNUU1ozSDY0M1FSMzJXQkpON0pLUw==",
          "node": {
            "description": "temporibus qui maiores et atque ipsum sunt ea maxime omnis.",
            "id": "EV01K9Q3TSZ3H643QR32WBJN7JKS",
            "name": "animi Event"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "RVYwMUs5UTNUU1ozSDY0M1FSMzJXQkpON0pLUw==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs5UTNUU1hETUYyRlg1MlhDR0FBRkRFNQ=="
      },
      "totalCount": 13
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
