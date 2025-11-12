# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for events query filtered by projectId. This tests:
- EventFilter input with projectId
- Filtering events belonging to a specific project
- Proper cursor pagination with filters applied

## Query

```graphql
query GetEventsByProject($projectId: ID!, $first: Int!) {
  events(filter: { projectId: $projectId }, first: $first) {
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
  "first": 10
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
- projectId filter correctly limits results to events in the specified project
- totalCount reflects filtered count, not all events
- Filtered results still maintain proper pagination structure
