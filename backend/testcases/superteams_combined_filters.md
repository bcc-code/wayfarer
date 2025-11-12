# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

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
  "projectId": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
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
- Multiple filters can be combined in a single query
- All filter conditions are applied correctly with AND logic
- totalCount reflects the combined filtered count
- Complex aggregations (team count and member count) work with filters
