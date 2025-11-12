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
          "cursor": "RVYwMUs5UTNUU1hETUYyRlg1MlhDR0FBRkRFNQ==",
          "node": {
            "description": "accusantium modi neque ex itaque est eligendi deleniti a dolore.",
            "endDate": "2025-10-14T15:48:29+02:00",
            "id": "EV01K9Q3TSXDMF2FX52XCGAAFDE5",
            "name": "eius Event",
            "startDate": "2025-10-11T15:48:29+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5UTNUU1ozSDY0M1FSMzJXQkpON0pLUw==",
          "node": {
            "description": "temporibus qui maiores et atque ipsum sunt ea maxime omnis.",
            "endDate": "2025-10-21T15:48:29+02:00",
            "id": "EV01K9Q3TSZ3H643QR32WBJN7JKS",
            "name": "animi Event",
            "startDate": "2025-10-18T15:48:29+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5UTNUU1pYUzRWTk5OTVE0V0pZUFhHVg==",
          "node": {
            "description": "est consectetur asperiores harum et quia quia provident fuga animi.",
            "endDate": "2025-10-28T15:48:29+01:00",
            "id": "EV01K9Q3TSZXS4VNNNMQ4WJYPXGV",
            "name": "quos Event",
            "startDate": "2025-10-25T15:48:29+02:00"
          }
        },
        {
          "cursor": "RVYwMUs5UTNUVDBROU1aQUZYOUc0WFo4R05SQw==",
          "node": {
            "description": "magni quaerat officiis sunt est modi nesciunt sed minima quidem.",
            "endDate": "2025-11-04T15:48:29+01:00",
            "id": "EV01K9Q3TT0Q9MZAFX9G4XZ8GNRC",
            "name": "id Event",
            "startDate": "2025-11-01T15:48:29+01:00"
          }
        },
        {
          "cursor": "RVYwMUs5UTNUVDFOWVpWSk4wWlBKRFZTVDFFTQ==",
          "node": {
            "description": "debitis impedit explicabo velit rerum facilis quia hic cupiditate eius.",
            "endDate": "2025-11-11T15:48:29+01:00",
            "id": "EV01K9Q3TT1NYZVJN0ZPJDVST1EM",
            "name": "nihil Event",
            "startDate": "2025-11-08T15:48:29+01:00"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "RVYwMUs5UTNUVDFOWVpWSk4wWlBKRFZTVDFFTQ==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs5UTNUU1hETUYyRlg1MlhDR0FBRkRFNQ=="
      },
      "totalCount": 13
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
