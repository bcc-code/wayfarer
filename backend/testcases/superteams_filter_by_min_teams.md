# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
          "cursor": "U1QwMUs4WFY2S0hTMThEMEtZQVM1NjE4UTE5SA==",
          "node": {
            "id": "ST01K8XV6KHS18D0KYAS5618Q19H",
            "name": "Cassin PLC Alliance",
            "description": "et ut nihil quia dignissimos temporibus ad voluptates."
          }
        },
        {
          "cursor": "U1QwMUs4WFY3QUpKV1ZINjkzTUFZV1gzMUpZMA==",
          "node": {
            "id": "ST01K8XV7AJJWVH693MAYWX31JY0",
            "name": "Friesen-Friesen Alliance",
            "description": "ut voluptate quos magnam libero ullam voluptates eaque."
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs4WFY2S0hTMThEMEtZQVM1NjE4UTE5SA==",
        "endCursor": "U1QwMUs4WFY3QUpKV1ZINjkzTUFZV1gzMUpZMA=="
      },
      "totalCount": 2
    }
  }
}
```

## Notes

This test verifies:
- minTeams filter correctly limits results to superteams with at least N teams
- totalCount reflects filtered count
- Aggregate counting of teams works correctly
