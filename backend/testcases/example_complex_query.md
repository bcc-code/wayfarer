# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Complex test case that combines multiple queries in a single request:
- `me` - Current user information
- `project` - Specific project by ID
- `projects` - Filtered list of projects with pagination
- `user` - Another user's information with their projects

This test verifies that multiple queries can be executed in a single GraphQL request
and that the server handles complex filters and nested fields correctly.

## Query

```graphql
query {
  me {
    id
    name
  }
  project(id: "PR01K8XV6J9H7BAEV49ZFVYS8R1K") {
    name
    description
    startDate
    endDate
  }

  projects(filter:{endDateAfter:"0020-01-02T15:04:05Z"}, first:100) {
    edges{
      node{
        name
      }
    }
    totalCount
  }

  user(id: "US01K8XV6J7WAC9WSFTPP376NPSD") {
    id
    name
    projects{
      id
    }
  }
}
```

## Expected

```json
{
  "data": {
    "me": {
      "id": "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
      "name": "Matjaz"
    },
    "project": {
      "name": "Summer Bible Camp 2025",
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "startDate": "2025-10-01T20:16:36+02:00",
      "endDate": "2025-12-30T20:16:36+01:00"
    },
    "projects": {
      "edges": [
        {
          "node": {
            "name": "Summer Bible Camp 2025"
          }
        },
        {
          "node": {
            "name": "Youth Winter Retreat 2024"
          }
        }
      ],
      "totalCount": 2
    },
    "user": {
      "id": "US01K8XV6J7WAC9WSFTPP376NPSD",
      "name": "Austin Parisian",
      "projects": [
        {
          "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K"
        },
        {
          "id": "PR01K8XV6JTX5GYJH9WNP4BN2P2A"
        }
      ]
    }
  }
}
```

## Notes

This test requires admin privileges to access the `user` query and `projects` with filters.
The test demonstrates:
- Multiple top-level queries in a single request
- Pagination with edges/nodes pattern
- Nested field resolution (user.projects)
- Date filtering with endDateAfter
