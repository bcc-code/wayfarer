# Wayfarer Mock GraphQL Server

A mock GraphQL server for frontend development based on the Wayfarer schema.

## Quick Start

```bash
# Install dependencies (first time only)
npm install

# Start the server
npm start

# Or use watch mode (auto-restart on changes)
npm run dev
```

The server will start at `http://localhost:4000/`

## Using with Nuxt Frontend

### Option 1: Configure Apollo/GraphQL Client

In your Nuxt app, configure your GraphQL client to point to the mock server:

```typescript
// nuxt.config.ts or plugin
export default defineNuxtConfig({
  runtimeConfig: {
    public: {
      graphqlUrl: 'http://localhost:4000/'
    }
  }
})
```

### Option 2: Use fetch/ofetch

```typescript
const { data } = await $fetch('http://localhost:4000/', {
  method: 'POST',
  body: {
    query: `
      query {
        me {
          id
          name
          email
        }
      }
    `
  }
})
```

## Example Queries

### Get Current User
```graphql
query {
  me {
    id
    name
    email
    age
    gender
    church {
      name
      country
      category
    }
  }
}
```

### Get Current Project
```graphql
query {
  currentProject {
    id
    name
    description
    startDate
    endDate
    branding {
      logo
      colors {
        primary
        secondary
        tertiary
      }
    }
  }
}
```

### Get Challenges
```graphql
query {
  challenges {
    id
    name
    description
    image
    buttonText
    publishedAt
    endTime
    userCompletedAt
  }
}
```

### Get Project with Achievements
```graphql
query {
  currentProject {
    id
    name
    achievements {
      id
      name
      description
      image
      points
      achievedAt
    }
  }
}
```

### Get Leaderboard
```graphql
query {
  currentProject {
    leaderboard(type: TOTAL) {
      name
      score
      description
      image
    }
  }
}
```

## Mock Data

The server generates realistic mock data for all types defined in the schema:

- **Users**: Random names, emails, ages, and genders
- **Projects**: Various project names with branding and colors
- **Events**: Event names with date ranges
- **Achievements**: Different types (Simple, Reading, Listening, Streak)
- **Teams & SuperTeams**: Team names and descriptions
- **Leaderboards**: Score entries with names and points

## Schema Notes

The mock server automatically fixes some schema issues:
- Converts `LeaderboardFilter` from type to input type
- Converts `AgeRange` from type to input type

These are temporary fixes for the mock server. Your actual backend should define these as input types in the schema.

## Testing

You can test the server with curl:

```bash
curl -X POST http://localhost:4000 \
  -H "Content-Type: application/json" \
  -d '{"query":"{ me { id name email } }"}'
```

## Apollo Studio Sandbox

Open http://localhost:4000/ in your browser to access the Apollo Studio Sandbox for interactive query testing.
