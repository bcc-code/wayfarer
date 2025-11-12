# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for querying a challenge by ID using the challenge(id) query. This tests:
- challenge(id) to fetch a specific challenge
- Challenge fields including id, name, description, image, url, buttonText, publishedAt, endTime
- Nested project resolution via dataloader

## Query

```graphql
query {
  challenge(id: "CL01K9Q3VD3ZSAN16YFGKWDNEH8D") {
    id
    name
    description
    image
    url
    buttonText
    publishedAt
    endTime
    project {
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
    "challenge": {
      "buttonText": "Accept",
      "description": "sed ut qui harum tenetur necessitatibus odio consequuntur deserunt velit quas id.",
      "endTime": "2026-01-07T15:48:49+01:00",
      "id": "CL01K9Q3VD3ZSAN16YFGKWDNEH8D",
      "image": "https://placecats.com/412/330",
      "name": "assumenda Challenge",
      "project": {
        "id": "PR01K9Q3TSVQ05C0E9GJ6QPXHFXP",
        "name": "Summer Bible Camp 2025"
      },
      "publishedAt": "2025-11-04T15:48:49+01:00",
      "url": "http://www.kdj.com/"
    }
  }
}
```
