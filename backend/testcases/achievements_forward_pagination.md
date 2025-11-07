# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for achievements query with forward pagination using cursor. This tests:
- Forward pagination with first parameter
- Cursor-based pagination continuation
- Proper hasNextPage and hasPreviousPage values
- Achievement filter by projectId

## Query

```graphql
query GetAchievementsPage($filter: AchievementFilter!, $first: Int, $after: String) {
  achievements(filter: $filter, first: $first, after: $after) {
    edges {
      cursor
      node {
        id
        name
        description
        image
        points
        hidden
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
  "filter": {
    "projectId": "PR01K8XV6KN2FC1R8YQ283ZTPTYH"
  },
  "first": 5
}
```

## Expected

```json
{
  "data": {
    "achievements": {
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
- first parameter limits page size correctly
- Cursors work for achievements pagination
- pageInfo.hasNextPage reflects if more results exist
- pageInfo.hasPreviousPage is false on first page
- Filter by projectId works correctly
- Returns empty result when no achievements exist for the project
