# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
          "cursor": "RVYwMUs4WFY2SkNNWDVDTlJYSzJENFhHQTI3Tg==",
          "node": {
            "id": "EV01K8XV6JCMX5CNRXK2D4XGA27N",
            "name": "aperiam Event",
            "description": "velit saepe labore omnis aut est mollitia fuga unde fuga."
          }
        },
        {
          "cursor": "RVYwMUs4WFY2SkZROFJFOUUwWTQ2SzExQUtKTQ==",
          "node": {
            "id": "EV01K8XV6JFQ8RE9E0Y46K11AKJM",
            "name": "distinctio Event",
            "description": "natus voluptatem non id ullam rerum id et eveniet nostrum."
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs4WFY2SkNNWDVDTlJYSzJENFhHQTI3Tg==",
        "endCursor": "RVYwMUs4WFY2SkZROFJFOUUwWTQ2SzExQUtKTQ=="
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
