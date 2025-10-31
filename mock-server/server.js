import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { addMocksToSchema } from '@graphql-tools/mock';
import { makeExecutableSchema } from '@graphql-tools/schema';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Read the schema and fix type/input issues
let typeDefs = readFileSync(join(__dirname, '../gql/wayfarer.graphqls'), 'utf-8');

// Fix LeaderboardFilter and AgeRange - they should be input types, not types
typeDefs = typeDefs.replace('type LeaderboardFilter {', 'input LeaderboardFilter {');
typeDefs = typeDefs.replace('type AgeRange {', 'input AgeRange {');

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
    achievedAt: () => new Date().toISOString(),
  }),

  LeaderboardEntry: () => ({
    name: () => ['Team Alpha', 'Team Beta', 'Jane Doe', 'John Smith'][Math.floor(Math.random() * 4)],
    score: () => Math.floor(Math.random() * 1000) + 100,
    description: () => Math.random() > 0.5 ? 'Doing great!' : null,
    image: () => Math.random() > 0.5 ? `https://picsum.photos/seed/${Math.random()}/100/100` : null,
  }),

  Branding: () => ({
    logo: () => `https://picsum.photos/seed/${Math.random()}/200/100`,
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

  // Query resolvers with more control
  UserQueryRoot: () => ({
    currentProject: () => ({}),
    currentEvent: () => ({}),
    me: () => ({}),
    challenges: () => [],
    projects: (_root, args) => {
      return args.ids.map(() => ({}));
    },
    events: (_root, args) => {
      return args.ids.map(() => ({}));
    },
  }),
};

// Create executable schema
const schema = makeExecutableSchema({ typeDefs });

// Add mocks to schema
const schemaWithMocks = addMocksToSchema({
  schema,
  mocks,
  preserveResolvers: false,
});

// Create Apollo Server
const server = new ApolloServer({
  schema: schemaWithMocks,
});

// Start the server
const { url } = await startStandaloneServer(server, {
  listen: { port: 4000 },
});

console.log(`🚀 Mock GraphQL server ready at: ${url}`);
console.log(`📝 GraphQL Playground available at: ${url}`);
console.log('');
console.log('Example query:');
console.log(`
  query {
    me {
      id
      name
      email
      age
    }
    currentProject {
      id
      name
      description
    }
  }
`);
