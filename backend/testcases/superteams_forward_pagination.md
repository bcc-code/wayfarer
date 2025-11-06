# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for superteams query with forward pagination using cursor. This tests:
- Forward pagination with first and after parameters
- Cursor-based pagination continuation
- Proper hasNextPage and hasPreviousPage values

## Query

```graphql
query GetSuperTeamsPage($first: Int!, $after: String) {
  superteams(first: $first, after: $after) {
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
  "first": 2
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
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs4WFY2S0VQNkpRTUFLVlI1OFhBRlcwVA==",
        "endCursor": "U1QwMUs4WFY2S0hTMThEMEtZQVM1NjE4UTE5SA=="
      },
      "totalCount": 9
    }
  }
}
```

## Notes

This test verifies:
- first parameter limits page size correctly
- Cursors can be used to navigate through pages
- pageInfo.hasNextPage is true when more results exist
- pageInfo.hasPreviousPage is false on first page
