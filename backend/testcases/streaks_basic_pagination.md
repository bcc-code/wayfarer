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
          "cursor": "U0swMUs5REdTNUNTRTBTU0MwSFM1OTMyNVhQTg==",
          "node": {
            "id": "SK01K9DGS5CSE0SSC0HS59325XPN",
            "name": "nulla Streak",
            "description": "Maintain your streak by completing activities daily!"
          }
        },
        {
          "cursor": "U0swMUs5REdTNUZUNllEOUpUSlAxSFFYOUNHMg==",
          "node": {
            "id": "SK01K9DGS5FT6YD9JTJP1HQX9CG2",
            "name": "dolorum Streak",
            "description": "Maintain your streak by completing activities daily!"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "hasPreviousPage": false,
        "startCursor": "U0swMUs5REdTNUNTRTBTU0MwSFM1OTMyNVhQTg==",
        "endCursor": "U0swMUs5REdTNUZUNllEOUpUSlAxSFFYOUNHMg=="
      },
      "totalCount": 5
    }
  }
}
```
