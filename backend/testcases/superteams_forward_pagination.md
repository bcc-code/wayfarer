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
          "cursor": "U1QwMUs5UTNUVEZWVFlFTTVXQkZaTTk4RzM0OQ==",
          "node": {
            "description": "doloremque totam optio et saepe sed omnis voluptas.",
            "id": "ST01K9Q3TTFVTYEM5WBFZM98G349",
            "name": "Kuvalis, Kuvalis and Kuvalis Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5UTNUVEhIV1pYSldQU1hWQlE5TUVNUA==",
          "node": {
            "description": "quae aliquid sequi fugiat est sequi voluptas ab.",
            "id": "ST01K9Q3TTHHWZXJWPSXVBQ9MEMP",
            "name": "DuBuque, DuBuque and DuBuque Alliance"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "U1QwMUs5UTNUVEhIV1pYSldQU1hWQlE5TUVNUA==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs5UTNUVEZWVFlFTTVXQkZaTTk4RzM0OQ=="
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
