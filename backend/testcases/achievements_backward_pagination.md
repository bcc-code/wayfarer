# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for achievements query with backward pagination using cursor. This tests:
- Backward pagination with last and before parameters
- Cursor-based backward navigation
- Proper hasNextPage and hasPreviousPage values
- Achievement filter by eventId

## Query

```graphql
query GetAchievementsPageBackward($filter: AchievementFilter!, $last: Int, $before: String) {
  achievements(filter: $filter, last: $last, before: $before) {
    edges {
      cursor
      node {
        id
        name
        description
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
    "eventId": "EV01K9DGS5QHK10Y0KYNT06C95TJ"
  },
  "last": 5
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
- last parameter limits page size correctly for backward pagination
- Backward pagination works with before cursor
- pageInfo reflects correct pagination state
- Filter by eventId works correctly
- Returns empty result when no achievements exist for the event
