# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for challenges query with backward pagination using cursor. This tests:
- Backward pagination with last and before parameters
- Cursor-based backward navigation
- Proper hasNextPage and hasPreviousPage values
- Challenge filter by eventId

## Query

```graphql
query GetChallengesPageBackward($filter: ChallengeFilter, $last: Int, $before: String) {
  challenges(filter: $filter, last: $last, before: $before) {
    edges {
      cursor
      node {
        id
        name
        description
        url
        buttonText
        publishedAt
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
    "eventId": "EV01K9DGS5QHK10Y0KYNT06C95TJ"
  },
  "last": 5
}
```

## Expected

```json
{
  "data": {
    "challenges": {
      "edges": [
        {
          "cursor": "Q0wwMUs5REdUODJXMVhBVDc5MkdNMU5CSkVBUA==",
          "node": {
            "buttonText": "Complete Challenge",
            "description": "voluptas expedita nostrum ab sit architecto quis ea odit totam quos modi.",
            "id": "CL01K9DGT82W1XAT792GM1NBJEAP",
            "name": "provident Challenge",
            "publishedAt": "2025-10-22T22:22:58+02:00",
            "url": "http://qsh.com/qlij-jz.html"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "Q0wwMUs5REdUODJXMVhBVDc5MkdNMU5CSkVBUA==",
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "Q0wwMUs5REdUODJXMVhBVDc5MkdNMU5CSkVBUA=="
      },
      "totalCount": 1
    }
  }
}
```

## Notes

This test verifies:
- last parameter limits page size correctly for backward pagination
- Backward pagination works with before cursor
- pageInfo reflects correct pagination state
- Filter by eventId works correctly
- Returns empty result when no challenges exist for the event
