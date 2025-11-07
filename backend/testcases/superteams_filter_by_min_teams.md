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
          "cursor": "U1QwMUs5REdTNjhEMjYwRjZaRDFBVkY0REhRMw==",
          "node": {
            "description": "voluptas molestiae consequuntur autem totam et sed impedit.",
            "id": "ST01K9DGS68D260F6ZD1AVF4DHQ3",
            "name": "Hills-Hills Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5REdTRlg5M1haOUVHSlQ2TlgxOVNUVw==",
          "node": {
            "description": "quis exercitationem enim unde ut occaecati pariatur non.",
            "id": "ST01K9DGSFX93XZ9EGJT6NX19STW",
            "name": "Brakus, Brakus and Brakus Alliance"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "U1QwMUs5REdTRlg5M1haOUVHSlQ2TlgxOVNUVw==",
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs5REdTNjhEMjYwRjZaRDFBVkY0REhRMw=="
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
