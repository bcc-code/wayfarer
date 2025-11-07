# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for challenges query filtering by projectId. This tests:
- Filter challenges by projectId
- Basic pagination without cursors
- Proper totalCount calculation

## Query

```graphql
query GetChallengesByProject($filter: ChallengeFilter, $first: Int) {
  challenges(filter: $filter, first: $first) {
    edges {
      cursor
      node {
        id
        name
        description
        image
        url
        buttonText
        publishedAt
        endTime
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
  "first": 10
}
```

## Expected

```json
{
  "data": {
    "challenges": {
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
- Filter by projectId returns only challenges for that project
- totalCount accurately reflects filtered results
- Empty result when no challenges exist for the project
- Basic pagination structure works correctly
