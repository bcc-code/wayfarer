# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Basic test for the `me` query. This should return the current user's information
including their ID, name, and email address.

## Query

```graphql
query {
  me {
    id
    name
    email
  }
}
```

## Expected

```json
{
  "data": {
    "me": {
      "email": "hackett.clemmie@gmail.com",
      "id": "US01K9DGS18D92WBMV3X7ETHNPMN",
      "name": "Alaina King"
    }
  }
}
```
