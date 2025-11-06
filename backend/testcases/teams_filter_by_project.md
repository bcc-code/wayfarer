# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
  "projectId": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
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
          "cursor": "VE0wMUs4WFY2S04yRkMxUjhZUTI4M1pUUFRZSA==",
          "node": {
            "id": "TM01K8XV6KN2FC1R8YQ283ZTPTYH",
            "name": "Claims Adjuster Team",
            "description": "aut nihil officiis eaque vel animi quia adipisci est quasi."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2TVM0QUI1M1haSEVHQlNLRzFEQg==",
          "node": {
            "id": "TM01K8XV6MS4AB53XZHEGBSKG1DB",
            "name": "Project Manager Team",
            "description": "quibusdam voluptas quia officia voluptatum est qui rerum ducimus ut."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2UDhSOEZCR0VGOUFRWTFQSEQ2Nw==",
          "node": {
            "id": "TM01K8XV6P8R8FBGEF9AQY1PHD67",
            "name": "Sheriff Team",
            "description": "debitis corporis est nostrum eius rerum impedit facilis est ullam."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2UUVEUkhUWU43WlE1NVYyWlRXSw==",
          "node": {
            "id": "TM01K8XV6QEDRHTYN7ZQ55V2ZTWK",
            "name": "Librarian Team",
            "description": "odio neque ut aut quae voluptas et numquam quaerat velit."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2UkdTRUJBVlpIUDY3M1FaM1FDNg==",
          "node": {
            "id": "TM01K8XV6RGSEBAVZHP673QZ3QC6",
            "name": "Administrative Support Supervisors Team",
            "description": "quis recusandae aspernatur optio delectus quasi natus explicabo qui adipisci."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2U0Y5R1ZRSjlWRUhBNDdIU1kwUw==",
          "node": {
            "id": "TM01K8XV6SF9GVQJ9VEHA47HSY0S",
            "name": "Electronics Engineer Team",
            "description": "iusto dolor aspernatur placeat cumque voluptatem modi earum culpa dolore."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2VDc2WkgwRUpLMU0wTlQwQjQxWA==",
          "node": {
            "id": "TM01K8XV6T76ZH0EJK1M0NT0B41X",
            "name": "Mechanical Equipment Sales Representative Team",
            "description": "consequatur possimus veritatis dolores incidunt consequuntur amet ea officiis dignissimos."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2Vks5RUQyR0JaU1EyVkRUQVQ4VA==",
          "node": {
            "id": "TM01K8XV6VK9ED2GBZSQ2VDTAT8T",
            "name": "Butcher Team",
            "description": "voluptatem quos similique odit accusantium velit unde praesentium consequuntur ut."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2V1FORjFGRzdSSjVQRDc1NkIyRw==",
          "node": {
            "id": "TM01K8XV6WQNF1FG7RJ5PD756B2G",
            "name": "Well and Core Drill Operator Team",
            "description": "ut ut in non labore veritatis consequatur dignissimos corporis nisi."
          }
        },
        {
          "cursor": "VE0wMUs4WFY2WTVNSlIzMjg4SDYyNUMzOEFIMA==",
          "node": {
            "id": "TM01K8XV6Y5MJR3288H625C38AH0",
            "name": "Central Office and PBX Installers Team",
            "description": "itaque eos aut consectetur sed voluptatem nisi delectus ut cum."
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "VE0wMUs4WFY2S04yRkMxUjhZUTI4M1pUUFRZSA==",
        "endCursor": "VE0wMUs4WFY2WTVNSlIzMjg4SDYyNUMzOEFIMA=="
      },
      "totalCount": 11
    }
  }
}
```

## Notes

This test verifies:
- projectId filter correctly limits results to teams in the specified project
- totalCount reflects filtered count
- Filtered results still maintain proper pagination structure
