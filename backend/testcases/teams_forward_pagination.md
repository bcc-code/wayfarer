# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

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
        }
      ],
      "pageInfo": {
        "endCursor": "VE0wMUs5REdTNzhXNzhRWDhZQTVEVFRQNkdKVw==",
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
- first parameter limits page size correctly
- Cursors can be used to navigate through pages
- pageInfo.hasNextPage is true when more results exist
- pageInfo.hasPreviousPage is false on first page
