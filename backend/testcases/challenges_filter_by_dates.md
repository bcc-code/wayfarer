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
      "edges": [
        {
          "cursor": "Q0wwMUs5REdUODJXMVhBVDc5MkdNMU5CSkVBUA==",
          "node": {
            "buttonText": "Accept",
            "description": "voluptas expedita nostrum ab sit architecto quis ea odit totam quos modi.",
            "endTime": null,
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
- Filter by publishedAfter returns only challenges published after that date
- Filter by publishedBefore returns only challenges published before that date
- Date range filtering works correctly when both filters are combined
- totalCount accurately reflects filtered results
- Combines eventId filter with date range filters
- Returns challenges within the specified date range
