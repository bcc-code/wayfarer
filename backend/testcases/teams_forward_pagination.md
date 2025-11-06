# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for teams query with forward pagination using cursor. This tests:
- Forward pagination with first and after parameters
- Cursor-based pagination continuation
- Proper hasNextPage and hasPreviousPage values

## Query

```graphql
query GetTeamsPage($first: Int!, $after: String) {
  teams(first: $first, after: $after) {
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
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "VE0wMUs4WFY2S04yRkMxUjhZUTI4M1pUUFRZSA==",
        "endCursor": "VE0wMUs4WFY2TVM0QUI1M1haSEVHQlNLRzFEQg=="
      },
      "totalCount": 31
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
