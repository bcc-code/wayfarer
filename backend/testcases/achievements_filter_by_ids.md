# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for achievements query filtering by specific IDs. This tests:
- Filtering achievements by specific IDs
- Pagination with IDs filter
- Proper handling of non-existent IDs

## Query

```graphql
query GetAchievementsByIds($filter: AchievementFilter!, $first: Int) {
  achievements(filter: $filter, first: $first) {
    edges {
      cursor
      node {
        id
        name
        points
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
    "ids": ["AC01K8XV6KN2FC1R8YQ283ZTPTYH", "AC01K8XV6MS4AB53XZHEGBSKG1DB"]
  },
  "first": 10
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
- Filtering by specific achievement IDs works
- Pagination works with IDs filter
- Returns empty result when specified IDs don't exist
- totalCount reflects actual number of matching achievements
