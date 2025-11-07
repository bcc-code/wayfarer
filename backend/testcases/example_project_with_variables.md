# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

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
  "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC"
}
```

## Expected

```json
{
  "data": {
    "project": {
      "description": "Join us for an amazing summer adventure exploring God's word!",
      "endDate": "2026-01-05T22:22:22+01:00",
      "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
      "name": "Summer Bible Camp 2025",
      "startDate": "2025-10-07T22:22:22+02:00"
    }
  }
}
```

## Notes

This query requires the user to have admin privileges. The test tool automatically
generates tokens with both "user" and "admin" roles for maximum compatibility.
