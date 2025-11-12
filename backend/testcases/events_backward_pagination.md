# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

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
          "cursor": "RVYwMUs5UTNUVENFUDNBQkU0Qk1UTUJFTjc0MQ==",
          "node": {
            "description": "quas hic fugit qui ad ad veritatis ex quis repellendus.",
            "endDate": "2025-05-11T15:48:29+02:00",
            "id": "EV01K9Q3TTCEP3ABE4BMTMBEN741",
            "name": "nesciunt Event",
            "startDate": "2025-05-08T15:48:29+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5UTNUVEQ5WkVDTUpLUzZGSFNIWkZWVg==",
          "node": {
            "description": "tenetur aspernatur sed ullam facere est odit et officia est.",
            "endDate": "2025-05-18T15:48:29+02:00",
            "id": "EV01K9Q3TTD9ZECMJKS6FHSHZFVV",
            "name": "architecto Event",
            "startDate": "2025-05-15T15:48:29+02:00"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "RVYwMUs5UTNUVEQ5WkVDTUpLUzZGSFNIWkZWVg==",
        "hasNextPage": false,
        "hasPreviousPage": true,
        "startCursor": "RVYwMUs5UTNUVENFUDNBQkU0Qk1UTUJFTjc0MQ=="
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
