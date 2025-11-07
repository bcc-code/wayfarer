# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for querying a specific superteam using superteam(id). This tests:
- superteam(id) to fetch a specific super team
- Basic fields: id, name, description
- parentProject relation

## Query

```graphql
query {
  superteam(id: "ST01K9DGS63ZSC3RN20R28MPMXA6") {
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
      "description": "et tenetur ad et repellat omnis expedita aperiam.",
      "id": "ST01K9DGS63ZSC3RN20R28MPMXA6",
      "name": "Mertz, Mertz and Mertz Alliance",
      "parentProject": {
        "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
