# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for querying a challenge by ID using the challenge(id) query. This tests:
- challenge(id) to fetch a specific challenge
- Challenge fields including id, name, description, image, url, buttonText, publishedAt, endTime
- Nested project resolution via dataloader

## Query

```graphql
query {
  challenge(id: "CL01K9DGT73BTRFRZ1JQCQ8CX8XS") {
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
      "description": "voluptatum rem excepturi et enim voluptates nemo amet consectetur voluptatem ut fuga.",
      "endTime": null,
      "id": "CL01K9DGT73BTRFRZ1JQCQ8CX8XS",
      "image": "https://placecats.com/499/319",
      "name": "corporis Challenge",
      "project": {
        "id": "PR01K9DGS50S1RZSE5HGN8JQ1XDC",
        "name": "Summer Bible Camp 2025"
      },
      "publishedAt": "2025-10-20T22:22:57+02:00",
      "url": "http://wgd.com/elwlx-l"
    }
  }
}
```
