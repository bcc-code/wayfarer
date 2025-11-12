# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for superteams query filtered by projectId. This tests:
- SuperTeamFilter input with projectId
- Filtering superteams belonging to a specific project
- Proper cursor pagination with filters applied

## Query

```graphql
query GetSuperTeamsByProject($projectId: ID!, $first: Int!) {
  superteams(filter: { projectId: $projectId }, first: $first) {
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
    "superteams": {
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
- projectId filter correctly limits results to superteams in the specified project
- totalCount reflects filtered count
- Filtered results still maintain proper pagination structure
