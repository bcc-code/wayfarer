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
          "cursor": "RVYwMUs5REdTNVo4Njg0RDU5MzZCUzlHRDFNQQ==",
          "node": {
            "description": "velit voluptas sint voluptatibus et est et maiores dolorem atque.",
            "endDate": "2025-04-30T22:22:23+02:00",
            "id": "EV01K9DGS5Z8684D5936BS9GD1MA",
            "name": "repellat Event",
            "startDate": "2025-04-27T22:22:23+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNjBXR1g2WkcyUzVCMkU1VkVIMA==",
          "node": {
            "description": "quod possimus accusamus non alias praesentium ut odio nesciunt est.",
            "endDate": "2025-05-07T22:22:23+02:00",
            "id": "EV01K9DGS60WGX6ZG2S5B2E5VEH0",
            "name": "eum Event",
            "startDate": "2025-05-04T22:22:23+02:00"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "RVYwMUs5REdTNjBXR1g2WkcyUzVCMkU1VkVIMA==",
        "hasNextPage": false,
        "hasPreviousPage": true,
        "startCursor": "RVYwMUs5REdTNVo4Njg0RDU5MzZCUzlHRDFNQQ=="
      },
      "totalCount": 12
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
