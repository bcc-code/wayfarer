# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for the `church(id)` query. This should return a church by its ID.

## Query

```graphql
query {
  church(id: "CH01K9Q3TQQJSXD62D5KZ3XYMAB4") {
    id
    name
    country
    category
  }
}
```

## Expected

```json
{
  "data": {
    "church": {
      "category": "S",
      "country": "Norway",
      "id": "CH01K9Q3TQQJSXD62D5KZ3XYMAB4",
      "name": "Stavanger Baptist Church"
    }
  }
}
```
