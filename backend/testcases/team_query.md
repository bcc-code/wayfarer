# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for querying a team by ID using the team(id) query. This tests:
- team(id) to fetch a specific team
- Team fields including id, name, description, and parentProject

## Query

```graphql
query {
  team(id: "TM01K8XV6VK9ED2GBZSQ2VDTAT8T") {
    id
    name
    description
    parentProject {
      id
      name
    }
  }
}
```

## Expected

```json
{
  "data": {
    "team": {
      "id": "TM01K8XV6VK9ED2GBZSQ2VDTAT8T",
      "name": "Butcher Team",
      "description": "voluptatem quos similique odit accusantium velit unde praesentium consequuntur ut.",
      "parentProject": {
        "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
