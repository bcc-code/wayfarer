# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for superteams query filtered by projectId. This tests:
- SuperTeamFilter input with projectId
- Filtering superteams belonging to a specific project
- Proper cursor pagination with filters applied

## Query

```graphql
query GetSuperTeamsByProject($projectId: ID!, $first: Int!) {
  superteams(filter: { projectId: $projectId }, first: $first) {
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
        }
      ],
      "pageInfo": {
        "endCursor": "U1QwMUs5REdTNjhEMjYwRjZaRDFBVkY0REhRMw==",
        "hasNextPage": false,
        "hasPreviousPage": false,
        "startCursor": "U1QwMUs5REdTNjNaU0MzUk4yMFIyOE1QTVhBNg=="
      },
      "totalCount": 3
    }
  }
}
```

## Notes

This test verifies:
- projectId filter correctly limits results to superteams in the specified project
- totalCount reflects filtered count
- Filtered results still maintain proper pagination structure
