# UserID: US01ARZ3NDEKTSV4RRFFQ69G5FAV

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
      "id": "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
      "name": "Matjaz",
      "email": "matjaz@example.com"
    }
  }
}
```
