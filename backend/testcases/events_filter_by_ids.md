# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for events query filtered by multiple IDs (bulk lookup). This tests:
- EventFilter input with ids array
- Bulk ID lookup for M2M API use cases
- Efficient fetching of multiple specific events
- Pagination with ID filter

This is particularly useful for M2M (machine-to-machine) APIs where external
systems need to fetch multiple events by their IDs in a single request.

## Query

```graphql
query GetEventsByIds($ids: [ID!]!, $first: Int!) {
  events(filter: { ids: $ids }, first: $first) {
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
    }
    totalCount
  }
}
```

## Variables

```json
{
  "ids": ["EV01K8XV6JHBAJNYHJXRYDYHD4WK"],
  "first": 10
}
```

## Expected

```json
{
  "data": {
    "events": {
      "edges": [
        {
          "cursor": "RVYwMUs4WFY2SkhCQUpOWUhKWFJZRFlIRDRXSw==",
          "node": {
            "id": "EV01K8XV6JHBAJNYHJXRYDYHD4WK",
            "name": "voluptatem Event",
            "description": "laboriosam rerum qui expedita enim unde ex et provident pariatur.",
            "startDate": "2025-10-15T20:16:36+02:00",
            "endDate": "2025-10-18T20:16:36+02:00"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": false
      },
      "totalCount": 1
    }
  }
}
```

## Notes

This test verifies:
- ids filter correctly limits results to specified event IDs
- Multiple IDs can be provided in array format
- Non-existent IDs are silently ignored (no error thrown)
- Results maintain pagination structure even with ID filter

Use cases for bulk ID lookup:
1. M2M API: External service has list of event IDs to fetch details for
2. Bookmarks/Favorites: User has saved event IDs, fetch all at once
3. Sync operations: Fetch specific events that changed since last sync
4. Batch operations: Process multiple events in single GraphQL query

Performance benefits:
- Single query instead of N individual queries
- Better caching with filter-based cache keys
- Reduced network overhead
- Can still use pagination if result set is large

Example with multiple IDs:
```json
{
  "ids": [
    "EV01K8XV6JHBAJNYHJXRYDYHD4WK",
    "EV01K8XV6JTYPEZV123456789ABC",
    "EV01K8XV6JXAMPLE123456789DEF"
  ],
  "first": 10
}
```
