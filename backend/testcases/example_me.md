# UserID: US01K9Q3TQYGR8W5JHW4GMVPWS44

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
      "email": "paula.durgan@hth.biz",
      "id": "US01K9Q3TQYGR8W5JHW4GMVPWS44",
      "name": "Art Fritsch"
    }
  }
}
```
