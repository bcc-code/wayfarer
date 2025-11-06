# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

## Description

Test for querying a specific superteam using superteam(id). This tests:
- superteam(id) to fetch a specific super team
- Basic fields: id, name, description
- parentProject relation

## Query

```graphql
query {
  superteam(id: "ST01K8XV6KHS18D0KYAS5618Q19H") {
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
    "superteam": {
      "id": "ST01K8XV6KHS18D0KYAS5618Q19H",
      "name": "Cassin PLC Alliance",
      "description": "et ut nihil quia dignissimos temporibus ad voluptates.",
      "parentProject": {
        "id": "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
