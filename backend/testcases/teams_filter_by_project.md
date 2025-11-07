# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for teams query filtered by projectId. This tests:
- TeamFilter input with projectId
- Filtering teams belonging to a specific project
- Proper cursor pagination with filters applied

## Query

```graphql
query GetTeamsByProject($projectId: ID!, $first: Int!) {
  teams(filter: { projectId: $projectId }, first: $first) {
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
  "first": 10
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
        },
        {
          "cursor": "VE0wMUs5REdTQjhLUlE5UzdaMFk3RTc3RTNNQw==",
          "node": {
            "description": "eum tenetur quos animi aut et qui totam sed voluptatibus.",
            "id": "TM01K9DGSB8KRQ9S7Z0Y7E77E3MC",
            "name": "Telecommunications Line Installer Team"
          }
        },
        {
          "cursor": "VE0wMUs5REdTQzY3MldZU0NEUjhGVDZOTk1GQw==",
          "node": {
            "description": "voluptas veritatis quos quia asperiores quia culpa eum temporibus architecto.",
            "id": "TM01K9DGSC672WYSCDR8FT6NNMFC",
            "name": "Home Team"
          }
        },
        {
          "cursor": "VE0wMUs5REdTREY3RDY3RzhCU0IwMTlCMkJaSw==",
          "node": {
            "description": "veritatis debitis odio nostrum quod rerum ipsam sequi sit dolorem.",
            "id": "TM01K9DGSDF7D67G8BSB019B2BZK",
            "name": "Metal Fabricator Team"
          }
        },
        {
          "cursor": "VE0wMUs5REdTRUhNN1dRMDNZNUtKVjAyVFZYOQ==",
          "node": {
            "description": "consequatur cum sunt perferendis doloremque quidem porro earum facilis quam.",
            "id": "TM01K9DGSEHM7WQ03Y5KJV02TVX9",
            "name": "Buffing and Polishing Operator Team"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "VE0wMUs5REdTRUhNN1dRMDNZNUtKVjAyVFZYOQ==",
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "VE0wMUs5REdTNjlYVFlUREU5Tlc1Q1oyMjJXQg=="
      },
      "totalCount": 9
    }
  }
}
```

## Notes

This test verifies:
- projectId filter correctly limits results to teams in the specified project
- totalCount reflects filtered count
- Filtered results still maintain proper pagination structure
