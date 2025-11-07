# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for basic teams pagination without filters. This tests:
- teams query with default pagination (first 5)
- TeamConnection type with edges, pageInfo, and totalCount
- Cursor-based pagination structure

## Query

```graphql
query {
  teams(first: 5) {
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
    "teams": {
      "edges": [
        {
          "cursor": "VE0wMUs5REdTNjlYVFlUREU5Tlc1Q1oyMjJXQg==",
          "node": {
            "description": "vel voluptas commodi non qui consequatur voluptate libero illo asperiores.",
            "id": "TM01K9DGS69XTYTDE9NW5CZ222WB",
            "name": "Coremaking Machine Operator Team"
          }
        },
        {
          "cursor": "VE0wMUs5REdTNzhXNzhRWDhZQTVEVFRQNkdKVw==",
          "node": {
            "description": "omnis sit tempore quis eligendi aut minima omnis tenetur qui.",
            "id": "TM01K9DGS78W78QX8YA5DTTP6GJW",
            "name": "Accountant Team"
          }
        },
        {
          "cursor": "VE0wMUs5REdTOFNXRjYzQjczUE1ZODg5N0FNRw==",
          "node": {
            "description": "voluptatem soluta harum quae voluptatem qui et nihil debitis illo.",
            "id": "TM01K9DGS8SWF63B73PMY8897AMG",
            "name": "Job Printer Team"
          }
        },
        {
          "cursor": "VE0wMUs5REdTOUg3TUNWQlRDWFRRTjJGQUVHVw==",
          "node": {
            "description": "deleniti sint exercitationem ab voluptates ad quia eveniet velit animi.",
            "id": "TM01K9DGS9H7MCVBTCXTQN2FAEGW",
            "name": "Home Appliance Repairer Team"
          }
        },
        {
          "cursor": "VE0wMUs5REdTQU1GVjlWSlNaUkVQRU05UTZUNQ==",
          "node": {
            "description": "occaecati tenetur deleniti aliquam nulla molestiae quis maxime quia voluptatem.",
            "id": "TM01K9DGSAMFV9VJSZREPEM9Q6T5",
            "name": "Plating Operator OR Coating Machine Operator Team"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "VE0wMUs5REdTQU1GVjlWSlNaUkVQRU05UTZUNQ==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "VE0wMUs5REdTNjlYVFlUREU5Tlc1Q1oyMjJXQg=="
      },
      "totalCount": 31
    }
  }
}
```

## Notes

This test verifies:
- TeamConnection structure matches the schema
- Cursors are properly encoded
- pageInfo fields correctly indicate pagination state
- totalCount reflects total number of teams
