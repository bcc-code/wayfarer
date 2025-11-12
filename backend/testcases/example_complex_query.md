# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

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
  project(id: "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP") {
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

  user(id: "US01K9Q3TQYGR8W5JHW4GMVPWS44") {
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
      "id": "US01K9Q3TQYGR8W5JHW4GMVPWS44",
      "name": "Art Fritsch"
    },
    "project": {
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "endDate": "2026-01-09T15:48:29+01:00",
      "name": "Summer Bible Camp 2025",
      "startDate": "2025-10-11T15:48:29+02:00"
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
      "id": "US01K9Q3TQYGR8W5JHW4GMVPWS44",
      "name": "Art Fritsch",
      "projects": [
        {
          "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP"
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
