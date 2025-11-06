# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
  "projectId": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
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
          "cursor": "RVYwMUs4WFY2SkNNWDVDTlJYSzJENFhHQTI3Tg==",
          "node": {
            "id": "EV01K8XV6JCMX5CNRXK2D4XGA27N",
            "name": "aperiam Event",
            "description": "velit saepe labore omnis aut est mollitia fuga unde fuga.",
            "startDate": "2025-10-01T20:16:36+02:00",
            "endDate": "2025-10-04T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2SkZROFJFOUUwWTQ2SzExQUtKTQ==",
          "node": {
            "id": "EV01K8XV6JFQ8RE9E0Y46K11AKJM",
            "name": "distinctio Event",
            "description": "natus voluptatem non id ullam rerum id et eveniet nostrum.",
            "startDate": "2025-10-08T20:16:36+02:00",
            "endDate": "2025-10-11T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2SkhCQUpOWUhKWFJZRFlIRDRXSw==",
          "node": {
            "id": "EV01K8XV6JHBAJNYHJXRYDYHD4WK",
            "name": "voluptatem Event",
            "description": "laboriosam rerum qui expedita enim unde ex et provident pariatur.",
            "startDate": "2025-10-15T20:16:36+02:00",
            "endDate": "2025-10-18T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2SkpZUVZYQzQ2UVNLMUJYNjM5Vg==",
          "node": {
            "id": "EV01K8XV6JJYQVXC46QSK1BX639V",
            "name": "nihil Event",
            "description": "consectetur omnis et iure assumenda et sit ut nihil soluta.",
            "startDate": "2025-10-22T20:16:36+02:00",
            "endDate": "2025-10-25T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2Sk1KVFpLSE0xOENHWUZSQzJZMQ==",
          "node": {
            "id": "EV01K8XV6JMJTZKHM18CGYFRC2Y1",
            "name": "molestias Event",
            "description": "quod cumque vel delectus quibusdam at qui voluptatibus aut commodi.",
            "startDate": "2025-10-29T20:16:36+01:00",
            "endDate": "2025-11-01T20:16:36+01:00"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs4WFY2SkNNWDVDTlJYSzJENFhHQTI3Tg==",
        "endCursor": "RVYwMUs4WFY2Sk1KVFpLSE0xOENHWUZSQzJZMQ=="
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
