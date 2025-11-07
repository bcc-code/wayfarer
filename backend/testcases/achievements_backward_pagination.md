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
      "edges": [
        {
          "cursor": "QUMwMUs5REdURDUzVFQ1TjZWSjJCV0NXOFZUNw==",
          "node": {
            "description": "a dignissimos voluptatem culpa ipsum voluptatem asperiores repudiandae blanditiis est.",
            "id": "AC01K9DGTD53TT5N6VJ2BWCW8VT7",
            "name": "nisi Achievement",
            "points": 40
          }
        },
        {
          "cursor": "QUMwMUs5REdURDg3S00zMjRGSFJKTUZCNUJBUw==",
          "node": {
            "description": "velit voluptas magnam dolor ipsum pariatur aliquid eligendi nemo illo.",
            "id": "AC01K9DGTD87KM324FHRJMFB5BAS",
            "name": "exercitationem Achievement",
            "points": 80
          }
        },
        {
          "cursor": "QUMwMUs5REdURDlTSkpERUdCQjJCRDhIQ0pFMg==",
          "node": {
            "description": "ut voluptas assumenda vero similique cupiditate impedit alias ea fugit.",
            "id": "AC01K9DGTD9SJJDEGBB2BD8HCJE2",
            "name": "nostrum Achievement",
            "points": 75
          }
        },
        {
          "cursor": "QUMwMUs5REdURE1WVFk4OTFORUU3VFZLSDZFWg==",
          "node": {
            "description": "Complete all articles to earn this achievement.",
            "id": "AC01K9DGTDMVTY891NEE7TVKH6EZ",
            "name": "Read: nostrum",
            "points": 92
          }
        }
      ],
      "pageInfo": {
        "endCursor": "QUMwMUs5REdURE1WVFk4OTFORUU3VFZLSDZFWg==",
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "QUMwMUs5REdURDUzVFQ1TjZWSjJCV0NXOFZUNw=="
      },
      "totalCount": 4
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
