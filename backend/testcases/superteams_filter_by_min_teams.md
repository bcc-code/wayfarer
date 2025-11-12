# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for superteams query filtered by minTeams. This tests:
- SuperTeamFilter input with minTeams
- Filtering superteams with minimum number of teams
- Proper cursor pagination with filters applied

## Query

```graphql
query GetSuperTeamsByMinTeams($minTeams: Int!, $first: Int!) {
  superteams(filter: { minTeams: $minTeams }, first: $first) {
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

## Variables

```json
{
  "minTeams": 3,
  "first": 5
}
```

## Expected

```json
{
  "data": {
    "superteams": {
      "edges": [
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
          "cursor": "U1QwMUs5UTNWNjk4RTNFNlczWjdYWTNCMVBSQg==",
          "node": {
            "description": "illo ut voluptates dignissimos voluptates nemo nihil rerum.",
            "id": "ST01K9Q3V698E3E6W3Z7XY3B1PRB",
            "name": "Gulgowski-Gulgowski Alliance"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "U1QwMUs5UTNWNjk4RTNFNlczWjdYWTNCMVBSQg==",
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs5UTNUVEpCR0RTSDlUUE5FRUgzM05TSg=="
      },
      "totalCount": 3
    }
  }
}
```

## Notes

This test verifies:
- minTeams filter correctly limits results to superteams with at least N teams
- totalCount reflects filtered count
- Aggregate counting of teams works correctly
