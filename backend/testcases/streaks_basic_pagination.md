# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for basic relay pagination on the `streaks` query. This tests:
- Forward pagination with first and after
- Correct edge structure with cursor and node
- PageInfo with hasNextPage and hasPreviousPage
- TotalCount field

## Query

```graphql
query {
  streaks(first: 2) {
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
    "streaks": {
      "edges": [
        {
          "cursor": "U0swMUs5UTNUVDJHMVYzU042NDdYN1FHSjdEOQ==",
          "node": {
            "description": "Maintain your streak by completing activities daily!",
            "id": "SK01K9Q3TT2G1V3SN647X7QGJ7D9",
            "name": "voluptas Streak"
          }
        },
        {
          "cursor": "U0swMUs5UTNUVDQ2VzU3RFg5NkFSVFBFVEVBQg==",
          "node": {
            "description": "Maintain your streak by completing activities daily!",
            "id": "SK01K9Q3TT46W57DX96ARTPETEAB",
            "name": "et Streak"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "U0swMUs5UTNUVDQ2VzU3RFg5NkFSVFBFVEVBQg==",
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "U0swMUs5UTNUVDJHMVYzU042NDdYN1FHSjdEOQ=="
      },
      "totalCount": 5
    }
  }
}
```
