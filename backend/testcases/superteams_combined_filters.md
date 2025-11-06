# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for superteams query with multiple filters combined. This tests:
- SuperTeamFilter with multiple filter parameters
- Combining projectId, minTeams, and maxMembers filters
- Complex filtering logic with proper cursor pagination

## Query

```graphql
query GetSuperTeamsWithCombinedFilters(
  $projectId: ID!,
  $minTeams: Int!,
  $maxMembers: Int!,
  $first: Int!
) {
  superteams(
    filter: {
      projectId: $projectId,
      minTeams: $minTeams,
      maxMembers: $maxMembers
    },
    first: $first
  ) {
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
  "projectId": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
  "minTeams": 2,
  "maxMembers": 50,
  "first": 10
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
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs4WFY2S0hTMThEMEtZQVM1NjE4UTE5SA==",
        "endCursor": "U1QwMUs4WFY2S0hTMThEMEtZQVM1NjE4UTE5SA=="
      },
      "totalCount": 1
    }
  }
}
```

## Notes

This test verifies:
- Multiple filters can be combined in a single query
- All filter conditions are applied correctly with AND logic
- totalCount reflects the combined filtered count
- Complex aggregations (team count and member count) work with filters
