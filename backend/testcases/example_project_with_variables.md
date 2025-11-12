# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

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
  "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP"
}
```

## Expected

```json
{
  "data": {
    "project": {
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "endDate": "2026-01-09T15:48:29+01:00",
      "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP",
      "name": "Summer Bible Camp 2025",
      "startDate": "2025-10-11T15:48:29+02:00"
    }
  }
}
```

## Notes

This query requires the user to have admin privileges. The test tool automatically
generates tokens with both "user" and "admin" roles for maximum compatibility.
