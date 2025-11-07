# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for basic events pagination without filters. This tests:
- events query with default pagination (first 10)
- EventConnection type with edges, pageInfo, and totalCount
- Cursor-based pagination structure

## Query

```graphql
query {
  events(first: 5) {
    edges {
      cursor
      node {
        id
        name
        description
        startDate
        endDate
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

## Expected

```json
{
  "data": {
    "events": {
      "edges": [
        {
          "cursor": "RVYwMUs5REdTNTNOSEhGQVFaNjBDOE1HOU44TQ==",
          "node": {
            "description": "rerum eveniet id et ducimus est et aliquid ducimus nesciunt.",
            "endDate": "2025-10-10T22:22:22+02:00",
            "id": "EV01K9DGS53NHHFAQZ60C8MG9N8M",
            "name": "dolorem Event",
            "startDate": "2025-10-07T22:22:22+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNTZLQjNCSFMzU1kzRU1OM1EzWA==",
          "node": {
            "description": "in autem corporis aliquid at ut aut autem et tempore.",
            "endDate": "2025-10-17T22:22:22+02:00",
            "id": "EV01K9DGS56KB3BHS3SY3EMN3Q3X",
            "name": "molestias Event",
            "startDate": "2025-10-14T22:22:22+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNTg0QU1ZWVk0RzJEMTMwRldWQw==",
          "node": {
            "description": "sed nostrum odit tempore illum molestias est est ut voluptas.",
            "endDate": "2025-10-24T22:22:22+02:00",
            "id": "EV01K9DGS584AMYYY4G2D130FWVC",
            "name": "esse Event",
            "startDate": "2025-10-21T22:22:22+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNTlQU1IyM1pSVlhQSzRORVJROA==",
          "node": {
            "description": "suscipit impedit cupiditate quia aut aut autem necessitatibus numquam qui.",
            "endDate": "2025-10-31T22:22:22+01:00",
            "id": "EV01K9DGS59PSR23ZRVXPK4NERQ8",
            "name": "omnis Event",
            "startDate": "2025-10-28T22:22:22+01:00"
          }
        },
        {
          "cursor": "RVYwMUs5REdTNUI4QlNIWVEzQ1pENDJEOTc2Rg==",
          "node": {
            "description": "harum non accusamus doloremque ea at ut voluptas et debitis.",
            "endDate": "2025-11-07T22:22:22+01:00",
            "id": "EV01K9DGS5B8BSHYQ3CZD42D976F",
            "name": "rerum Event",
            "startDate": "2025-11-04T22:22:22+01:00"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "RVYwMUs5REdTNUI4QlNIWVEzQ1pENDJEOTc2Rg==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs5REdTNTNOSEhGQVFaNjBDOE1HOU44TQ=="
      },
      "totalCount": 12
    }
  }
}
```

## Notes

This test verifies:
- EventConnection structure matches the schema
- Cursors are properly encoded
- pageInfo fields correctly indicate pagination state
- totalCount reflects total number of events (not just page size)
