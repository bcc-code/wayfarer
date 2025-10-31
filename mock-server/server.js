import { createYoga } from 'graphql-yoga';
import { createServer } from 'node:http';
import { addMocksToSchema } from '@graphql-tools/mock';
import { makeExecutableSchema } from '@graphql-tools/schema';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Read all schema files
const sharedSchema = readFileSync(join(__dirname, '../gql/shared.graphqls'), 'utf-8');
const userSchema = readFileSync(join(__dirname, '../gql/user.graphqls'), 'utf-8');
const adminSchema = readFileSync(join(__dirname, '../gql/admin.graphqls'), 'utf-8');
const m2mSchema = readFileSync(join(__dirname, '../gql/m2m.graphqls'), 'utf-8');

// Remove individual schema definitions and combine
const removeSchemaDefinition = (schema) => {
  return schema.replace(/schema\s*{[^}]*}/g, '');
};

// Combine all schemas with a unified schema definition that exposes all APIs
const typeDefs = `
${sharedSchema}
${removeSchemaDefinition(userSchema)}
${removeSchemaDefinition(adminSchema)}
${removeSchemaDefinition(m2mSchema)}

schema {
  query: CombinedQuery
  mutation: CombinedMutation
}

type CombinedQuery {
  user: UserQueryRoot!
  admin: AdminQueryRoot!
  m2m: M2MQueryRoot!
}

type CombinedMutation {
  user: UserMutationRoot!
  admin: AdminMutationRoot!
  m2m: M2MMutationRoot!
}
`;

// Create mock data generators
const mocks = {
  ID: () => Math.random().toString(36).substring(2, 15),
  DateTime: () => new Date().toISOString(),
  HTML: () => '<p>This is <strong>HTML</strong> content with formatting.</p>',

  User: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    membersId: () => 'MEM-' + Math.floor(Math.random() * 10000),
    name: () => ['John Doe', 'Jane Smith', 'Bob Johnson', 'Alice Williams'][Math.floor(Math.random() * 4)],
    email: () => `user${Math.floor(Math.random() * 1000)}@example.com`,
    age: () => Math.floor(Math.random() * 50) + 15,
    gender: () => Math.random() > 0.5 ? 'MALE' : 'FEMALE',
  }),

  Church: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Grace Church', 'Faith Community', 'Hope Chapel', 'Trinity Church'][Math.floor(Math.random() * 4)],
    country: () => ['Norway', 'Sweden', 'Denmark', 'Finland'][Math.floor(Math.random() * 4)],
    category: () => ['S', 'L', 'XL'][Math.floor(Math.random() * 3)],
  }),

  Project: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Summer Camp 2024', 'Winter Retreat', 'Bible Study Challenge'][Math.floor(Math.random() * 3)],
    description: () => 'An exciting project to engage youth in biblical learning and community building.',
    startDate: () => new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
    endDate: () => new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
  }),

  Event: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Week 1', 'Week 2', 'Opening Ceremony', 'Final Challenge'][Math.floor(Math.random() * 4)],
    description: () => 'A special event as part of the project timeline.',
    startDate: () => new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    endDate: () => new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
  }),

  Team: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Team Alpha', 'Team Beta', 'Team Gamma', 'Team Delta'][Math.floor(Math.random() * 4)],
    description: () => 'A dedicated team working together towards common goals.',
  }),

  SuperTeam: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Red Division', 'Blue Division', 'Green Division'][Math.floor(Math.random() * 3)],
    description: () => 'A super team consisting of multiple teams.',
  }),

  Challenge: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Daily Reading', 'Prayer Challenge', 'Scripture Memory'][Math.floor(Math.random() * 3)],
    description: () => '<p>Complete this <strong>exciting challenge</strong> to earn points!</p>',
    image: () => `https://picsum.photos/seed/${Math.random()}/400/300`,
    url: () => `https://example.com/challenges/${Math.random().toString(36).substring(2, 15)}`,
    buttonText: () => ['Start Challenge', 'Begin', 'Take Challenge'][Math.floor(Math.random() * 3)],
    publishedAt: () => new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    endTime: () => Math.random() > 0.5 ? new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString() : null,
    userCompletedAt: () => Math.random() > 0.6 ? new Date().toISOString() : null,
  }),

  Article: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    title: () => ['Understanding Grace', 'Faith in Action', 'The Power of Prayer'][Math.floor(Math.random() * 3)],
    author: () => ['C.S. Lewis', 'John Piper', 'Tim Keller'][Math.floor(Math.random() * 3)],
    url: () => `https://example.com/articles/${Math.random().toString(36).substring(2, 15)}`,
  }),

  Track: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Sermon on the Mount', 'Psalms Reading', 'Gospel Reflection'][Math.floor(Math.random() * 3)],
    description: () => 'An audio track for listening and reflection.',
    image: () => `https://picsum.photos/seed/${Math.random()}/400/300`,
  }),

  SimpleAchievement: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['First Steps', 'Early Bird', 'Consistent Learner'][Math.floor(Math.random() * 3)],
    description: () => 'Complete your first challenge.',
    image: () => `https://picsum.photos/seed/${Math.random()}/200/200`,
    points: () => Math.floor(Math.random() * 100) + 10,
    hidden: () => Math.random() > 0.7,
    achievedAt: () => Math.random() > 0.5 ? new Date().toISOString() : null,
  }),

  ReadingAchievement: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Bookworm', 'Bible Scholar', 'Knowledge Seeker'][Math.floor(Math.random() * 3)],
    description: () => 'Read all the assigned articles.',
    image: () => `https://picsum.photos/seed/${Math.random()}/200/200`,
    points: () => Math.floor(Math.random() * 100) + 10,
    hidden: () => Math.random() > 0.8,
    achievedAt: () => Math.random() > 0.6 ? new Date().toISOString() : null,
  }),

  ListeningAchievement: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Attentive Listener', 'Audio Devotee', 'Sound Student'][Math.floor(Math.random() * 3)],
    description: () => 'Listen to all the audio tracks.',
    image: () => `https://picsum.photos/seed/${Math.random()}/200/200`,
    points: () => Math.floor(Math.random() * 100) + 10,
    hidden: () => Math.random() > 0.8,
    achievedAt: () => new Date().toISOString(),
  }),

  StreakAchievement: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Streak Master', 'Consistency King', 'Daily Dedication'][Math.floor(Math.random() * 3)],
    description: () => 'Maintain a daily streak.',
    image: () => `https://picsum.photos/seed/${Math.random()}/200/200`,
    points: () => Math.floor(Math.random() * 100) + 10,
    hidden: () => Math.random() > 0.8,
    neededStreak: () => Math.floor(Math.random() * 20) + 5,
    currentStreak: () => Math.floor(Math.random() * 15),
    achievedAt: () => Math.random() > 0.5 ? new Date().toISOString() : null,
  }),

  LeaderboardEntry: () => ({
    name: () => ['Team Alpha', 'Team Beta', 'Jane Doe', 'John Smith'][Math.floor(Math.random() * 4)],
    score: () => Math.floor(Math.random() * 1000) + 100,
    description: () => Math.random() > 0.5 ? 'Doing great!' : null,
    image: () => Math.random() > 0.5 ? `https://picsum.photos/seed/${Math.random()}/100/100` : null,
  }),

  Branding: () => ({
    logo: () => `https://picsum.photos/seed/${Math.random()}/200/100`,
    rounding: () => [0, 4, 8, 12, 16][Math.floor(Math.random() * 5)],
  }),

  Colors: () => ({
    primary: () => ['#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A'][Math.floor(Math.random() * 4)],
    secondary: () => ['#95E1D3', '#F38181', '#AA96DA', '#FCBAD3'][Math.floor(Math.random() * 4)],
    tertiary: () => ['#C7CEEA', '#FFDAC1', '#E2F0CB', '#B5EAD7'][Math.floor(Math.random() * 4)],
  }),

  AgeRange: () => ({
    min: () => 15,
    max: () => 25,
  }),

  Streak: () => ({
    id: () => Math.random().toString(36).substring(2, 15),
    name: () => ['Daily Devotion', 'Weekly Reading', 'Monthly Challenge'][Math.floor(Math.random() * 3)],
    description: () => 'Keep up your streak by completing daily activities.',
    status: () => Math.floor(Math.random() * 30),
  }),

  DateRange: () => ({
    start: () => new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    end: () => new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
  }),

  Date: () => new Date().toISOString().split('T')[0],
  Upload: () => null,

  // Combined root resolvers
  CombinedQuery: () => ({
    user: () => ({}),
    admin: () => ({}),
    m2m: () => ({}),
  }),

  CombinedMutation: () => ({
    user: () => ({}),
    admin: () => ({}),
    m2m: () => ({}),
  }),

  // User API Query resolvers
  UserQueryRoot: () => ({
    currentProject: () => ({}),
    currentEvent: () => ({}),
    me: () => ({}),
    projects: () => Array(3).fill(null).map(() => ({})),
    events: () => Array(2).fill(null).map(() => ({})),
  }),

  UserMutationRoot: () => ({}),

  // Admin API Query resolvers
  AdminQueryRoot: () => ({
    user: () => ({}),
    users: () => Array(5).fill(null).map(() => ({})),
    project: () => ({}),
    projects: () => Array(3).fill(null).map(() => ({})),
    event: () => ({}),
    events: () => Array(2).fill(null).map(() => ({})),
    team: () => ({}),
    teams: () => Array(4).fill(null).map(() => ({})),
    superteam: () => ({}),
    superteams: () => Array(2).fill(null).map(() => ({})),
    achievement: () => ({}),
    achievements: () => Array(5).fill(null).map(() => ({})),
    challenge: () => ({}),
    challenges: () => Array(3).fill(null).map(() => ({})),
    church: () => ({}),
    churches: () => Array(10).fill(null).map(() => ({})),
    streak: () => ({}),
    streaks: () => Array(2).fill(null).map(() => ({})),
    currentProject: () => ({}),
    currentEvent: () => ({}),
  }),

  AdminMutationRoot: () => ({}),

  // M2M API Query resolvers
  M2MQueryRoot: () => ({
    user: () => ({}),
    project: () => ({}),
    event: () => ({}),
    team: () => ({}),
    superteam: () => ({}),
    achievement: () => ({}),
    challenge: () => ({}),
    users: () => Array(5).fill(null).map(() => ({})),
    challenges: () => Array(3).fill(null).map(() => ({})),
    achievements: () => Array(5).fill(null).map(() => ({})),
    currentProject: () => ({}),
    currentEvent: () => ({}),
  }),

  M2MMutationRoot: () => ({}),
};

// Create executable schema
const schema = makeExecutableSchema({ typeDefs });

// Add mocks to schema
const schemaWithMocks = addMocksToSchema({
  schema,
  mocks,
  preserveResolvers: false,
});

// Create Yoga instance with GraphiQL
const yoga = createYoga({
  schema: schemaWithMocks,
  graphiql: true,
});

// Create and start the server
const server = createServer(yoga);
const port = 4000;

server.listen(port, () => {
  console.log(`🚀 Mock GraphQL server ready at: http://localhost:${port}/graphql`);
  console.log(`📝 GraphiQL UI available at: http://localhost:${port}/graphql`);
console.log('');
console.log('The schema now exposes three separate APIs:');
console.log('  - user: End user API (mobile/web apps)');
console.log('  - admin: Admin API (management interface)');
console.log('  - m2m: Machine-to-Machine API (external integrations)');
console.log('');
console.log('Example User API query:');
console.log(`
  query {
    user {
      me {
        id
        name
        email
        age
        gender
        church {
          name
          country
        }
      }
      currentProject {
        id
        name
        description
        startDate
        endDate
      }
      projects {
        id
        name
        events {
          id
          name
        }
      }
    }
  }
`);
console.log('');
console.log('Example Admin API query:');
console.log(`
  query {
    admin {
      projects {
        id
        name
        description
      }
      users(filter: { gender: MALE }) {
        id
        name
        email
      }
    }
  }
`);
console.log('');
console.log('Example M2M API mutation:');
console.log(`
  mutation {
    m2m {
      awardAchievement(userId: "user123", achievementId: "ach456") {
        id
        name
        points
      }
    }
  }
`);
});
