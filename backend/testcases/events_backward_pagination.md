# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for events query with backward pagination using cursor. This tests:
- Backward pagination with last and before parameters
- Cursor-based reverse navigation
- Proper hasNextPage and hasPreviousPage values in backward mode
- Results are returned in correct order (not reversed)

Backward pagination is useful for:
- "Load previous page" buttons
- Navigating backwards through results
- Loading the last N items

## Query

```graphql
query GetEventsPageBackward($last: Int!, $before: String) {
  events(last: $last, before: $before) {
    edges {
      cursor
      node {
        id
        name
        description
        startDate
        endDate
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
  "last": 2
}
```

## Expected

```json
{
  "data": {
    "events": {
      "edges": [
        {
          "cursor": "RVYwMUs4WFY2SzlNQUFQMDlYR05TRDQzQ1hKOA==",
          "node": {
            "id": "EV01K8XV6K9MAAP09XGNSD43CXJ8",
            "name": "atque Event",
            "description": "et sunt culpa modi earum iusto est maxime reiciendis voluptatibus.",
            "startDate": "2025-05-05T20:16:37+02:00",
            "endDate": "2025-05-08T20:16:37+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2S0JBNURSREhXQzk2RjlaR1AxUw==",
          "node": {
            "id": "EV01K8XV6KBA5DRDHWC96F9ZGP1S",
            "name": "voluptas Event",
            "description": "unde molestias non id numquam quas eaque accusamus id asperiores.",
            "startDate": "2025-05-12T20:16:37+02:00",
            "endDate": "2025-05-15T20:16:37+02:00"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": true,
        "startCursor": "RVYwMUs4WFY2SzlNQUFQMDlYR05TRDQzQ1hKOA==",
        "endCursor": "RVYwMUs4WFY2S0JBNURSREhXQzk2RjlaR1AxUw=="
      },
      "totalCount": 13
    }
  }
}
```

## Notes

This test verifies:
- last parameter limits page size correctly
- Results are returned in forward order (not reversed)
- pageInfo.hasPreviousPage is true when more results exist before cursor
- pageInfo.hasNextPage is false on last page, true when using before cursor

Pagination flow with backward navigation:
1. User is on page N with cursor C
2. User clicks "Previous Page"
3. Query with last: 10, before: C
4. Get previous 10 items before cursor C
5. Check hasPreviousPage to enable/disable "Previous" button
6. Check hasNextPage (should be true since we came from a later page)

Important notes:
- Cannot use both first and last in same query
- Results are always in forward order, even with backward pagination
- The implementation fetches in reverse, then reverses the results
