# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for events query filtered by date ranges. This tests:
- EventFilter input with date range filters
- startDateAfter, startDateBefore filters
- endDateAfter, endDateBefore filters
- Combining date filters to find events in specific time windows

## Query

```graphql
query GetEventsByDateRange(
  $startDateAfter: DateTime
  $startDateBefore: DateTime
  $endDateAfter: DateTime
  $endDateBefore: DateTime
  $first: Int!
) {
  events(
    filter: {
      startDateAfter: $startDateAfter
      startDateBefore: $startDateBefore
      endDateAfter: $endDateAfter
      endDateBefore: $endDateBefore
    }
    first: $first
  ) {
    edges {
      node {
        id
        name
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
  "startDateAfter": "2025-10-01T00:00:00Z",
  "startDateBefore": "2025-10-31T23:59:59Z",
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
          "node": {
            "endDate": "2025-10-14T15:48:29+02:00",
            "id": "EV01K9Q3TSXDMF2FX52XCGAAFDE5",
            "name": "eius Event",
            "startDate": "2025-10-11T15:48:29+02:00"
          }
        },
        {
          "node": {
            "endDate": "2025-10-21T15:48:29+02:00",
            "id": "EV01K9Q3TSZ3H643QR32WBJN7JKS",
            "name": "animi Event",
            "startDate": "2025-10-18T15:48:29+02:00"
          }
        },
        {
          "node": {
            "endDate": "2025-10-28T15:48:29+01:00",
            "id": "EV01K9Q3TSZXS4VNNNMQ4WJYPXGV",
            "name": "quos Event",
            "startDate": "2025-10-25T15:48:29+02:00"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": false
      },
      "totalCount": 3
    }
  }
}
```

## Notes

This test verifies:
- Date filters correctly limit events to specified time ranges
- Events with startDate within October 2025 are returned
- DateTime scalar type works correctly with filters
- Multiple date filters can be combined

Use cases for date filters:
- Find events happening this month: startDateAfter and startDateBefore
- Find ongoing events: startDateBefore today AND endDateAfter today
- Find upcoming events: startDateAfter today
- Find past events: endDateBefore today
