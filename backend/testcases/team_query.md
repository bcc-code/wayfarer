# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for querying a team by ID using the team(id) query. This tests:
- team(id) to fetch a specific team
- Team fields including id, name, description, and parentProject

## Query

```graphql
query {
  team(id: "TM01K9DGS69XTYTDE9NW5CZ222WB") {
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
      "description": "vel voluptas commodi non qui consequatur voluptate libero illo asperiores.",
      "id": "TM01K9DGS69XTYTDE9NW5CZ222WB",
      "name": "Coremaking Machine Operator Team",
      "parentProject": {
        "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
        "name": "Summer Bible Camp 2025"
      }
    }
  }
}
```
