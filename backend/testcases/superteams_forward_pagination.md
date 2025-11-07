# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

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
          "cursor": "U1QwMUs5REdTNjNaU0MzUk4yMFIyOE1QTVhBNg==",
          "node": {
            "description": "et tenetur ad et repellat omnis expedita aperiam.",
            "id": "ST01K9DGS63ZSC3RN20R28MPMXA6",
            "name": "Mertz, Mertz and Mertz Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5REdTNjZYQ1dOMTM1VjZLUjc1MDhCNw==",
          "node": {
            "description": "tempore id minus nihil omnis doloremque dolores ut.",
            "id": "ST01K9DGS66XCWN135V6KR7508B7",
            "name": "Koelpin and Sons Alliance"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "U1QwMUs5REdTNjZYQ1dOMTM1VjZLUjc1MDhCNw==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs5REdTNjNaU0MzUk4yMFIyOE1QTVhBNg=="
      },
      "totalCount": 8
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
