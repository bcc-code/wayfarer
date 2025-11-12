# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for challenges query filtering by published dates. This tests:
- Filter challenges by publishedAfter and publishedBefore
- Date range filtering
- Proper totalCount calculation with date filters

## Query

```graphql
query GetChallengesByDates($filter: ChallengeFilter, $first: Int) {
  challenges(filter: $filter, first: $first) {
    edges {
      cursor
      node {
        id
        name
        description
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
    "eventId": "EV01K9DGS5QHK10Y0KYNT06C95TJ",
    "publishedAfter": "2020-01-01T00:00:00Z",
    "publishedBefore": "2030-12-31T23:59:59Z"
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
- Filter by publishedAfter returns only challenges published after that date
- Filter by publishedBefore returns only challenges published before that date
- Date range filtering works correctly when both filters are combined
- totalCount accurately reflects filtered results
- Combines eventId filter with date range filters
- Returns challenges within the specified date range
