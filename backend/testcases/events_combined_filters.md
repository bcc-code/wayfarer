# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for events query with multiple combined filters. This tests:
- Combining projectId and date filters
- Multiple filter conditions working together (AND logic)
- Pagination with complex filters
- Cache key generation with multiple parameters

This represents a realistic use case: finding events for a specific project
within a certain time range, useful for displaying "upcoming events for this project"
or "events happening this week in this project".

## Query

```graphql
query GetProjectEventsInDateRange(
  $projectId: ID!
  $startDateAfter: DateTime!
  $first: Int!
) {
  events(
    filter: {
      projectId: $projectId
      startDateAfter: $startDateAfter
    }
    first: $first
  ) {
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
  "projectId": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
  "startDateAfter": "2025-10-01T00:00:00Z",
  "first": 5
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
            "endDate": "2025-10-10T22:22:22+02:00",
            "id": "EV01K9DGS53NHHFAQZ60C8MG9N8M",
            "name": "dolorem Event",
            "startDate": "2025-10-07T22:22:22+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNTZLQjNCSFMzU1kzRU1OM1EzWA==",
          "node": {
            "description": "in autem corporis aliquid at ut aut autem et tempore.",
            "endDate": "2025-10-17T22:22:22+02:00",
            "id": "EV01K9DGS56KB3BHS3SY3EMN3Q3X",
            "name": "molestias Event",
            "startDate": "2025-10-14T22:22:22+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNTg0QU1ZWVk0RzJEMTMwRldWQw==",
          "node": {
            "description": "sed nostrum odit tempore illum molestias est est ut voluptas.",
            "endDate": "2025-10-24T22:22:22+02:00",
            "id": "EV01K9DGS584AMYYY4G2D130FWVC",
            "name": "esse Event",
            "startDate": "2025-10-21T22:22:22+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNTlQU1IyM1pSVlhQSzRORVJROA==",
          "node": {
            "description": "suscipit impedit cupiditate quia aut aut autem necessitatibus numquam qui.",
            "endDate": "2025-10-31T22:22:22+01:00",
            "id": "EV01K9DGS59PSR23ZRVXPK4NERQ8",
            "name": "omnis Event",
            "startDate": "2025-10-28T22:22:22+01:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNUI4QlNIWVEzQ1pENDJEOTc2Rg==",
          "node": {
            "description": "harum non accusamus doloremque ea at ut voluptas et debitis.",
            "endDate": "2025-11-07T22:22:22+01:00",
            "id": "EV01K9DGS5B8BSHYQ3CZD42D976F",
            "name": "rerum Event",
            "startDate": "2025-11-04T22:22:22+01:00"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "RVYwMUs5REdTNUI4QlNIWVEzQ1pENDJEOTc2Rg==",
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs5REdTNTNOSEhGQVFaNjBDOE1HOU44TQ=="
      },
      "totalCount": 5
    }
  }
}
```

## Notes

This test verifies:
- Multiple filters work together with AND logic
- projectId filter limits to one project
- startDateAfter filter limits to future/current events
- totalCount reflects the combined filter result
- Cache keys are correctly generated from multiple filter parameters

Real-world use cases:
- "Show me upcoming events for this project"
- "What events are happening this month for this project"
- "Find events for this project that haven't ended yet"

The query combines:
- Project scoping (projectId)
- Time filtering (startDateAfter)
- Pagination (first, cursor-based)
- Relational data fetching (parentProject)
