# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "VE0wMUs4WFY2S04yRkMxUjhZUTI4M1pUUFRZSA==",
        "endCursor": "VE0wMUs4WFY2UkdTRUJBVlpIUDY3M1FaM1FDNg=="
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
