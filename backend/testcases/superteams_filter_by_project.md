# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for superteams query filtered by projectId. This tests:
- SuperTeamFilter input with projectId
- Filtering superteams belonging to a specific project
- Proper cursor pagination with filters applied

## Query

```graphql
query GetSuperTeamsByProject($projectId: ID!, $first: Int!) {
  superteams(filter: { projectId: $projectId }, first: $first) {
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
          "cursor": "U1QwMUs4WFY2S0VQNkpRTUFLVlI1OFhBRlcwVA==",
          "node": {
            "id": "ST01K8XV6KEP6JQMAKVR58XAFW0T",
            "name": "Pfeffer-Pfeffer Alliance",
            "description": "enim repudiandae perferendis veniam adipisci sint facilis velit."
          }
        },
        {
          "cursor": "U1QwMUs4WFY2S0hTMThEMEtZQVM1NjE4UTE5SA==",
          "node": {
            "id": "ST01K8XV6KHS18D0KYAS5618Q19H",
            "name": "Cassin PLC Alliance",
            "description": "et ut nihil quia dignissimos temporibus ad voluptates."
          }
        },
        {
          "cursor": "U1QwMUs4WFY2S0tFUUJZUDBEOTg2UEY5NURDVg==",
          "node": {
            "id": "ST01K8XV6KKEQBYP0D986PF95DCV",
            "name": "Bartoletti, Bartoletti and Bartoletti Alliance",
            "description": "et iste placeat aut sequi et quas voluptatibus."
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs4WFY2S0VQNkpRTUFLVlI1OFhBRlcwVA==",
        "endCursor": "U1QwMUs4WFY2S0tFUUJZUDBEOTg2UEY5NURDVg=="
      },
      "totalCount": 3
    }
  }
}
```

## Notes

This test verifies:
- projectId filter correctly limits results to superteams in the specified project
- totalCount reflects filtered count
- Filtered results still maintain proper pagination structure
