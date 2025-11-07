# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description

Test for filtering streaks by project ID. This tests:
- Filter parameter working correctly
- Only returns streaks for the specified project

## Query

```graphql
query {
  streaks(filter: { projectId: "PR01K9DGS5HA9SWE7T8FCJ975JZT" }, first: 10) {
    edges {
      node {
        id
        name
        project {
          id
          name
        }
      }
    }
    totalCount
  }
}
```

## Expected

```json
{
  "data": {
    "streaks": {
      "edges": [
        {
          "node": {
            "id": "SK01K9DGS5S4CEFDC3XQC2HQ7HGW",
            "name": "similique Streak",
            "project": {
              "id": "PR01K9DGS5HA9SWE7T8FCJ975JZT",
              "name": "Youth Winter Retreat 2024"
            }
          }
        },
        {
          "node": {
            "id": "SK01K9DGS5TP6WWQGZ3CVMAY5WDQ",
            "name": "accusantium Streak",
            "project": {
              "id": "PR01K9DGS5HA9SWE7T8FCJ975JZT",
              "name": "Youth Winter Retreat 2024"
            }
          }
        }
      ],
      "totalCount": 2
    }
  }
}
```
