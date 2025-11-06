# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for the `project` query with variables. This demonstrates how to pass
variables to GraphQL queries in test cases.

This query requires admin role to access a specific project by ID.

## Query

```graphql
query GetProject($id: ID!) {
  project(id: $id) {
    id
    name
    description
    startDate
    endDate
  }
}
```

## Variables

```json
{
  "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K"
}
```

## Expected

```json
{
  "data": {
    "project": {
      "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
      "name": "Summer Bible Camp 2025",
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "startDate": "2025-10-01T20:16:36+02:00",
      "endDate": "2025-12-30T20:16:36+01:00"
    }
  }
}
```

## Notes

This query requires the user to have admin privileges. The test tool automatically
generates tokens with both "user" and "admin" roles for maximum compatibility.
