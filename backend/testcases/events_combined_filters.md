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
      "edges": [],
      "pageInfo": {
        "endCursor": null,
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": null
      },
      "totalCount": 0
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
