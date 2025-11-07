# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for basic superteams pagination without filters. This tests:
- superteams query with default pagination (first 5)
- SuperTeamConnection type with edges, pageInfo, and totalCount
- Cursor-based pagination structure

## Query

```graphql
query {
  superteams(first: 5) {
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
        },
        {
          "cursor": "U1QwMUs5REdTNjhEMjYwRjZaRDFBVkY0REhRMw==",
          "node": {
            "description": "voluptas molestiae consequuntur autem totam et sed impedit.",
            "id": "ST01K9DGS68D260F6ZD1AVF4DHQ3",
            "name": "Hills-Hills Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5REdTRlg5M1haOUVHSlQ2TlgxOVNUVw==",
          "node": {
            "description": "quis exercitationem enim unde ut occaecati pariatur non.",
            "id": "ST01K9DGSFX93XZ9EGJT6NX19STW",
            "name": "Brakus, Brakus and Brakus Alliance"
          }
        },
        {
          "cursor": "U1QwMUs5REdTRllSUEJSWUdQWEc3S0hHR0E2SA==",
          "node": {
            "description": "nam iste et modi ducimus iusto qui molestias.",
            "id": "ST01K9DGSFYRPBRYGPXG7KHGGA6H",
            "name": "Miller, Miller and Miller Alliance"
          }
        }
      ],
      "pageInfo": {
        "endCursor": "U1QwMUs5REdTRllSUEJSWUdQWEc3S0hHR0E2SA==",
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
- SuperTeamConnection structure matches the schema
- Cursors are properly encoded
- pageInfo fields correctly indicate pagination state
- totalCount reflects total number of super teams
