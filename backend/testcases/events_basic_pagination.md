# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
          "cursor": "RVYwMUs4WFY2SkNNWDVDTlJYSzJENFhHQTI3Tg==",
          "node": {
            "id": "EV01K8XV6JCMX5CNRXK2D4XGA27N",
            "name": "aperiam Event",
            "description": "velit saepe labore omnis aut est mollitia fuga unde fuga.",
            "startDate": "2025-10-01T20:16:36+02:00",
            "endDate": "2025-10-04T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2SkZROFJFOUUwWTQ2SzExQUtKTQ==",
          "node": {
            "id": "EV01K8XV6JFQ8RE9E0Y46K11AKJM",
            "name": "distinctio Event",
            "description": "natus voluptatem non id ullam rerum id et eveniet nostrum.",
            "startDate": "2025-10-08T20:16:36+02:00",
            "endDate": "2025-10-11T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2SkhCQUpOWUhKWFJZRFlIRDRXSw==",
          "node": {
            "id": "EV01K8XV6JHBAJNYHJXRYDYHD4WK",
            "name": "voluptatem Event",
            "description": "laboriosam rerum qui expedita enim unde ex et provident pariatur.",
            "startDate": "2025-10-15T20:16:36+02:00",
            "endDate": "2025-10-18T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2SkpZUVZYQzQ2UVNLMUJYNjM5Vg==",
          "node": {
            "id": "EV01K8XV6JJYQVXC46QSK1BX639V",
            "name": "nihil Event",
            "description": "consectetur omnis et iure assumenda et sit ut nihil soluta.",
            "startDate": "2025-10-22T20:16:36+02:00",
            "endDate": "2025-10-25T20:16:36+02:00"
          }
        },
        {
          "cursor": "RVYwMUs4WFY2Sk1KVFpLSE0xOENHWUZSQzJZMQ==",
          "node": {
            "id": "EV01K8XV6JMJTZKHM18CGYFRC2Y1",
            "name": "molestias Event",
            "description": "quod cumque vel delectus quibusdam at qui voluptatibus aut commodi.",
            "startDate": "2025-10-29T20:16:36+01:00",
            "endDate": "2025-11-01T20:16:36+01:00"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "RVYwMUs4WFY2SkNNWDVDTlJYSzJENFhHQTI3Tg==",
        "endCursor": "RVYwMUs4WFY2Sk1KVFpLSE0xOENHWUZSQzJZMQ=="
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
