# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
            "id": "EV01K8XV6JCMX5CNRXK2D4XGA27N",
            "name": "aperiam Event",
            "startDate": "2025-10-01T20:16:36+02:00",
            "endDate": "2025-10-04T20:16:36+02:00"
          }
        },
        {
          "node": {
            "id": "EV01K8XV6JFQ8RE9E0Y46K11AKJM",
            "name": "distinctio Event",
            "startDate": "2025-10-08T20:16:36+02:00",
            "endDate": "2025-10-11T20:16:36+02:00"
          }
        },
        {
          "node": {
            "id": "EV01K8XV6JHBAJNYHJXRYDYHD4WK",
            "name": "voluptatem Event",
            "startDate": "2025-10-15T20:16:36+02:00",
            "endDate": "2025-10-18T20:16:36+02:00"
          }
        },
        {
          "node": {
            "id": "EV01K8XV6JJYQVXC46QSK1BX639V",
            "name": "nihil Event",
            "startDate": "2025-10-22T20:16:36+02:00",
            "endDate": "2025-10-25T20:16:36+02:00"
          }
        },
        {
          "node": {
            "id": "EV01K8XV6JMJTZKHM18CGYFRC2Y1",
            "name": "molestias Event",
            "startDate": "2025-10-29T20:16:36+01:00",
            "endDate": "2025-11-01T20:16:36+01:00"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": false
      },
      "totalCount": 5
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
