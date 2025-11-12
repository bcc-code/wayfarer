# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for basic superteams pagination without filters. This tests:
- superteams query with default pagination (first 5)
- SuperTeamConnection type with edges, pageInfo, and totalCount
- Cursor-based pagination structure

## Query

```graphql
query {
  superteams(first: 5) {
    edges {
      cursor
      node {
        id
        name
        description
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
    "superteams": {
      "edges": [
        {
          "cursor": "U1QwMUs5UTNUVEZWVFlFTTVXQkZaTTk4RzM0OQ==",
          "node": {
            "description": "doloremque totam optio et saepe sed omnis voluptas.",
            "id": "ST01K9Q3TTFVTYEM5WBFZM98G349",
            "name": "Kuvalis, Kuvalis and Kuvalis Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5UTNUVEhIV1pYSldQU1hWQlE5TUVNUA==",
          "node": {
            "description": "quae aliquid sequi fugiat est sequi voluptas ab.",
            "id": "ST01K9Q3TTHHWZXJWPSXVBQ9MEMP",
            "name": "DuBuque, DuBuque and DuBuque Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5UTNUVEpCR0RTSDlUUE5FRUgzM05TSg==",
          "node": {
            "description": "vero fuga et inventore ullam qui optio odit.",
            "id": "ST01K9Q3TTJBGDSH9TPNEEH33NSJ",
            "name": "Kovacek Group Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5UTNWMEpUSjhGMEI0M1QyQVdXRUNGUw==",
          "node": {
            "description": "ut eaque cum eligendi et aut corrupti ea.",
            "id": "ST01K9Q3V0JTJ8F0B43T2AWWECFS",
            "name": "Ernser and Sons Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5UTNWMEtLRFpIVFJWODdDOVJCV1oxUg==",
          "node": {
            "description": "non cupiditate consectetur perspiciatis dolores quibusdam dicta ea.",
            "id": "ST01K9Q3V0KKDZHTRV87C9RBWZ1R",
            "name": "Cole, Cole and Cole Alliance"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "U1QwMUs5UTNWMEtLRFpIVFJWODdDOVJCV1oxUg==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs5UTNUVEZWVFlFTTVXQkZaTTk4RzM0OQ=="
      },
      "totalCount": 8
    }
  }
}
```

## Notes

This test verifies:
- SuperTeamConnection structure matches the schema
- Cursors are properly encoded
- pageInfo fields correctly indicate pagination state
- totalCount reflects total number of super teams
