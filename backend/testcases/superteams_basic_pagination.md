# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
        },
        {
          "cursor": "U1QwMUs4WFY3MDdYUlBTMTY2RDBLRzAzQjUwQw==",
          "node": {
            "id": "ST01K8XV707XRPS166D0KG03B50C",
            "name": "Robel, Robel and Robel Alliance",
            "description": "asperiores et quo eum sed cumque et facilis."
          }
        },
        {
          "cursor": "U1QwMUs4WFY3MDlIVk5HRjA0N0FUNEhQR0VEWg==",
          "node": {
            "id": "ST01K8XV709HVNGF047AT4HPGEDZ",
            "name": "Durgan-Durgan Alliance",
            "description": "quam rerum magnam explicabo eveniet qui eum aut."
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs4WFY2S0VQNkpRTUFLVlI1OFhBRlcwVA==",
        "endCursor": "U1QwMUs4WFY3MDlIVk5HRjA0N0FUNEhQR0VEWg=="
      },
      "totalCount": 9
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
