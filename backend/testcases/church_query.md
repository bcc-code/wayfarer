# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for the `church(id)` query. This should return a church by its ID.

## Query

```graphql
query {
  church(id: "CH01K9DGS10Y2KXG2CNWT4DR1ZCG") {
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
      "id": "CH01K9DGS10Y2KXG2CNWT4DR1ZCG",
      "name": "Stockholm Evangelical Church",
      "country": "Sweden",
      "category": "XL"
    }
  }
}
```
