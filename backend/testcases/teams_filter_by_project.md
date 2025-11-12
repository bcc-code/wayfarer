# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for teams query filtered by projectId. This tests:
- TeamFilter input with projectId
- Filtering teams belonging to a specific project
- Proper cursor pagination with filters applied

## Query

```graphql
query GetTeamsByProject($projectId: ID!, $first: Int!) {
  teams(filter: { projectId: $projectId }, first: $first) {
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
  "projectId": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
  "first": 10
}
```

## Expected

```json
{
  "data": {
    "teams": {
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
- projectId filter correctly limits results to teams in the specified project
- totalCount reflects filtered count
- Filtered results still maintain proper pagination structure
