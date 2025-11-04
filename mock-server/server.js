import { createYoga } from 'graphql-yoga';
import { createServer } from 'node:http';
import { addMocksToSchema } from '@graphql-tools/mock';
import { makeExecutableSchema } from '@graphql-tools/schema';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { faker } from '@faker-js/faker';
import { ulid } from 'ulid';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Load seeded data
const seedData = JSON.parse(readFileSync(join(__dirname, 'data.json'), 'utf-8'));

// Initialize seeded projects and events with proper IDs
const seededProjects = seedData.projects.map(project => {
  const projectId = 'PR' + ulid();
  const events = project.events.map(event => {
    const eventId = 'EV' + ulid();
    const eventObj = {
      __typename: 'Event',
      id: eventId,
      name: event.name,
      description: faker.lorem.sentence(),
      startDate: project.startDate,
      endDate: project.endDate,
      __seeded: true, // Mark as seeded
    };
    return eventObj;
  });

  return {
    __typename: 'Project',
    id: projectId,
    name: project.name,
    description: faker.lorem.paragraph(),
    startDate: project.startDate,
    endDate: project.endDate,
    events: events,
    __seeded: true, // Mark as seeded
  };
});

// Create a map for quick lookups
const projectsMap = new Map(seededProjects.map(p => [p.id, p]));
const eventsMap = new Map();
seededProjects.forEach(project => {
  project.events.forEach(event => {
    eventsMap.set(event.id, event);
  });
});

// Track current context for resolvers
let eventIndex = 0;
let currentProjectEvents = [];

// Helper function to remove internal fields
const cleanSeededData = (obj) => {
  if (!obj) return obj;
  const { __seeded, __typename, ...clean } = obj;

  // Also clean nested event arrays
  if (clean.events && Array.isArray(clean.events)) {
    clean.events = clean.events.map(e => {
      const { __seeded: _, __typename: __, ...cleanEvent } = e;
      return cleanEvent;
    });
  }

  return clean;
};

// Read all schema files
const sharedSchema = readFileSync(join(__dirname, '../gql/shared.graphqls'), 'utf-8');
const mainSchema = readFileSync(join(__dirname, '../gql/schema.graphqls'), 'utf-8');

// Combine schemas
const typeDefs = `
${sharedSchema}
${mainSchema}
`;

// Create mock data generators
const mocks = {
  ID: () => faker.string.uuid(),
  DateTime: () => faker.date.recent().toISOString(),
  HTML: () => `<p>${faker.lorem.paragraph()}</p><p><strong>${faker.lorem.sentence()}</strong></p>`,

  User: () => ({
    id: () => faker.string.uuid(),
    membersId: () => 'MEM-' + faker.number.int({ min: 1000, max: 99999 }),
    name: () => faker.person.fullName(),
    email: () => faker.internet.email().toLowerCase(),
    age: () => faker.number.int({ min: 15, max: 65 }),
    gender: () => faker.helpers.arrayElement(['MALE', 'FEMALE']),
    image: () => faker.image.personPortrait()
  }),

  Church: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Church', 'Chapel', 'Community', 'Fellowship', 'Parish'])}`,
      `${faker.person.lastName()} ${faker.helpers.arrayElement(['Church', 'Chapel', 'Community'])}`,
      `${faker.location.city()} ${faker.helpers.arrayElement(['Church', 'Chapel', 'Community'])}`
    ]),
    country: () => faker.location.country(),
    category: () => faker.helpers.arrayElement(['S', 'L', 'XL']),
  }),

  Project: (parent) => {
    // If this is a seeded project, preserve its data
    if (parent && parent.__seeded) {
      return parent;
    }

    // Otherwise generate mock data
    const timeState = faker.helpers.arrayElement(['past', 'active', 'future']);

    let startDate, endDate;
    if (timeState === 'past') {
      // Past project: both dates in the past
      startDate = faker.date.past({ years: 1 });
      endDate = faker.date.between({ from: startDate, to: new Date() });
    } else if (timeState === 'active') {
      // Active project: started in past, ends in future
      startDate = faker.date.recent({ days: 30 });
      endDate = faker.date.soon({ days: 60 });
    } else {
      // Future project: both dates in the future
      startDate = faker.date.soon({ days: 30 });
      endDate = faker.date.soon({ days: 90 });
    }

    return {
      id: () => faker.string.uuid(),
      name: () => faker.helpers.arrayElement([
        `${faker.date.month()} ${faker.helpers.arrayElement(['Camp', 'Retreat', 'Study', 'Challenge', 'Journey'])} ${new Date().getFullYear()}`,
        `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Bible Study', 'Discipleship', 'Youth Camp', 'Faith Journey'])}`,
        `The ${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Path', 'Way', 'Journey', 'Quest'])}`
      ]),
      description: () => faker.lorem.paragraph(),
      startDate: () => startDate.toISOString(),
      endDate: () => endDate.toISOString(),
    };
  },

  Event: (parent) => {
    // If this is a seeded event, preserve its data
    if (parent && parent.__seeded) {
      return parent;
    }
    // Otherwise generate mock data
    return {
      id: () => faker.string.uuid(),
      name: () => faker.helpers.arrayElement([
        `Week ${faker.number.int({ min: 1, max: 8 })}`,
        `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Ceremony', 'Challenge', 'Session', 'Gathering'])}`,
        `${faker.date.weekday()} ${faker.helpers.arrayElement(['Meeting', 'Service', 'Event'])}`
      ]),
      description: () => faker.lorem.sentence(),
      startDate: () => faker.date.recent({ days: 7 }).toISOString(),
      endDate: () => faker.date.soon({ days: 14 }).toISOString(),
    };
  },

  Team: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `Team ${faker.color.human()}`,
      `The ${faker.animal.type()}s`,
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Warriors', 'Champions', 'Heroes', 'Squad'])}`,
      faker.location.city()
    ]),
    description: () => faker.lorem.sentence(),
  }),

  SuperTeam: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.color.human()} Division`,
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Alliance', 'Coalition', 'League'])}`,
      `The ${faker.animal.type()} Group`
    ]),
    description: () => faker.lorem.sentence(),
  }),

  Challenge: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Reading', 'Prayer', 'Service', 'Study', 'Memory'])} Challenge`,
      `${faker.number.int({ min: 7, max: 30 })} Day ${faker.helpers.arrayElement(['Devotion', 'Scripture', 'Prayer'])}`,
      `The ${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Quest', 'Journey', 'Path'])}`
    ]),
    description: () => `<p>${faker.lorem.paragraph()}</p><p><strong>${faker.lorem.sentence()}</strong></p>`,
    image: () => faker.image.url(),
    url: () => faker.internet.url(),
    buttonText: () => faker.helpers.arrayElement(['Start Challenge', 'Begin', 'Take Challenge', 'Accept', 'Join Now']),
    publishedAt: () => faker.date.recent({ days: 7 }).toISOString(),
    endTime: () => faker.datatype.boolean() ? faker.date.soon({ days: 14 }).toISOString() : null,
    userCompletedAt: () => faker.datatype.boolean(0.4) ? faker.date.recent().toISOString() : null,
  }),

  Article: () => ({
    id: () => faker.string.uuid(),
    title: () => faker.lorem.sentence({ min: 3, max: 8 }).replace('.', ''),
    author: () => faker.person.fullName(),
    url: () => faker.internet.url(),
  }),

  Track: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.music.genre()} ${faker.helpers.arrayElement(['Worship', 'Meditation', 'Study'])}`,
      `${faker.lorem.words(3)}`,
      `Episode ${faker.number.int({ min: 1, max: 100 })}: ${faker.lorem.words(3)}`
    ]),
    description: () => faker.lorem.sentence(),
    image: () => faker.image.url(),
  }),

  SimpleAchievement: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Starter', 'Beginner', 'Achiever', 'Champion'])}`,
      `${faker.number.int({ min: 1, max: 10 })} ${faker.helpers.arrayElement(['Steps', 'Milestones', 'Days', 'Weeks'])}`,
      `The ${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Path', 'Journey', 'Way'])}`
    ]),
    description: () => faker.lorem.sentence(),
    image: () => faker.image.url(),
    points: () => faker.number.int({ min: 10, max: 100 }),
    hidden: () => faker.datatype.boolean(0.3),
    achievedAt: () => faker.datatype.boolean() ? faker.date.recent().toISOString() : null,
  }),

  ReadingAchievement: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Reader', 'Scholar', 'Bookworm', 'Student'])}`,
      `${faker.number.int({ min: 5, max: 100 })} ${faker.helpers.arrayElement(['Articles', 'Pages', 'Books', 'Chapters'])}`,
      `The ${faker.word.adjective({ capitalize: true })} Mind`
    ]),
    description: () => faker.lorem.sentence(),
    image: () => faker.image.url(),
    points: () => faker.number.int({ min: 20, max: 150 }),
    hidden: () => faker.datatype.boolean(0.2),
    achievedAt: () => faker.datatype.boolean(0.6) ? faker.date.recent().toISOString() : null,
  }),

  ListeningAchievement: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Listener', 'Ear', 'Devotee', 'Enthusiast'])}`,
      `${faker.number.int({ min: 5, max: 50 })} ${faker.helpers.arrayElement(['Tracks', 'Hours', 'Sessions'])}`,
      `Audio ${faker.word.adjective({ capitalize: true })}`
    ]),
    description: () => faker.lorem.sentence(),
    image: () => faker.image.url(),
    points: () => faker.number.int({ min: 30, max: 120 }),
    hidden: () => faker.datatype.boolean(0.2),
    achievedAt: () => faker.date.recent().toISOString(),
  }),

  StreakAchievement: () => {
    const neededStreak = faker.number.int({ min: 5, max: 25 });
    const currentStreak = faker.number.int({ min: 0, max: neededStreak + 5 });
    return {
      id: () => faker.string.uuid(),
      name: () => faker.helpers.arrayElement([
        `${neededStreak} Day Streak`,
        `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Consistency', 'Dedication', 'Commitment'])}`,
        `Streak ${faker.word.adjective({ capitalize: true })}`
      ]),
      description: () => faker.lorem.sentence(),
      image: () => faker.image.url(),
      points: () => faker.number.int({ min: 50, max: 200 }),
      hidden: () => faker.datatype.boolean(0.2),
      neededStreak: () => neededStreak,
      currentStreak: () => currentStreak,
      achievedAt: () => currentStreak >= neededStreak ? faker.date.recent().toISOString() : null,
    };
  },

  LeaderboardEntry: () => ({
    name: () => faker.datatype.boolean()
      ? faker.person.fullName()
      : `Team ${faker.color.human()}`,
    score: () => faker.number.int({ min: 100, max: 10000 }),
    description: () => faker.datatype.boolean() ? faker.lorem.sentence({ min: 1, max: 3 }) : null,
    image: () => faker.datatype.boolean() ? faker.image.url() : null,
  }),

  Branding: () => ({
    logo: () => faker.image.avatar(),
    colors: () => ({}),
    rounding: () => faker.helpers.arrayElement([0, 4, 8, 12, 16, 20, 24]),
  }),

  Colors: () => ({
    primary: () => faker.color.rgb(),
    secondary: () => faker.color.rgb(),
    tertiary: () => faker.color.rgb(),
  }),

  AgeRange: () => ({
    min: () => faker.number.int({ min: 12, max: 18 }),
    max: () => faker.number.int({ min: 20, max: 35 }),
  }),

  Streak: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `${faker.date.weekday()} ${faker.helpers.arrayElement(['Devotion', 'Reading', 'Prayer'])}`,
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Streak', 'Challenge', 'Journey'])}`,
      `${faker.number.int({ min: 7, max: 30 })} Day Challenge`
    ]),
    description: () => faker.lorem.sentence(),
    status: () => faker.number.int({ min: 0, max: 30 }),
  }),

  DateRange: () => ({
    start: () => faker.date.recent({ days: 30 }).toISOString().split('T')[0],
    end: () => faker.date.soon({ days: 30 }).toISOString().split('T')[0],
  }),

  Date: () => faker.date.recent().toISOString().split('T')[0],
  Upload: () => null,

  // Query root resolvers
  Query: () => ({
    // User API queries
    me: () => ({}),
    myProjects: () => seededProjects.map(p => cleanSeededData(p)),
    myEvents: () => seededProjects.flatMap(p => p.events).map(e => cleanSeededData(e)),
    myCurrentProject: () => cleanSeededData(seededProjects[0]) || {},
    myCurrentEvent: () => cleanSeededData(seededProjects[0]?.events[0]) || {},

    // Admin/M2M API queries
    user: () => ({}),
    users: () => Array(100).fill(null).map(() => ({})),
    project: (root, args) => {
      if (args && args.id) {
        return cleanSeededData(projectsMap.get(args.id)) || {};
      }
      return cleanSeededData(seededProjects[0]) || {};
    },
    projects: () => seededProjects.map(p => cleanSeededData(p)),
    event: (root, args) => {
      if (args && args.id) {
        return cleanSeededData(eventsMap.get(args.id)) || {};
      }
      return cleanSeededData(seededProjects[0]?.events[0]) || {};
    },
    events: () => seededProjects.flatMap(p => p.events).map(e => cleanSeededData(e)),
    team: () => ({}),
    teams: () => Array(20).fill(null).map(() => ({})),
    superteam: () => ({}),
    superteams: () => Array(4).fill(null).map(() => ({})),
    achievement: () => ({}),
    achievements: () => Array(5).fill(null).map(() => ({})),
    challenge: () => ({}),
    challenges: () => Array(3).fill(null).map(() => ({})),
    church: () => ({}),
    churches: () => Array(10).fill(null).map(() => ({})),
    streak: () => ({}),
    streaks: () => Array(2).fill(null).map(() => ({})),
    currentProject: () => cleanSeededData(seededProjects[0]) || {},
    currentEvent: () => cleanSeededData(seededProjects[0]?.events[0]) || {},
  }),

  // Mutation root resolvers
  Mutation: () => ({}),
};

// Create executable schema with custom resolvers for seeded data
const resolvers = {
  Project: {
    events: (parent) => {
      // If this is a seeded project, return its events without internal fields
      if (parent.__seeded && parent.events) {
        return parent.events.map(e => cleanSeededData(e));
      }
      // Otherwise return empty to let mocks handle it
      return undefined;
    },
  },
};

const schema = makeExecutableSchema({ typeDefs, resolvers });

// Add mocks to schema
const schemaWithMocks = addMocksToSchema({
  schema,
  mocks,
  preserveResolvers: true, // Preserve our custom resolvers
});

// Create Yoga instance with GraphiQL
const yoga = createYoga({
  schema: schemaWithMocks,
  graphiql: true,
});

// Create and start the server
const server = createServer(yoga);
const port = 8080;

server.listen(port, () => {
  console.log(`🚀 Mock GraphQL server ready at: http://localhost:${port}/graphql`);
  console.log(`📝 GraphiQL UI available at: http://localhost:${port}/graphql`);
  console.log('');
  console.log('The schema exposes a unified GraphQL API with role-based access control:');
  console.log('  - User queries: me, myProjects, myEvents, myCurrentProject, myCurrentEvent');
  console.log('  - Admin/M2M queries: user, users, project, projects, event, events, etc.');
  console.log('  - Mutations use @requireRole directives for access control');
  console.log('');
  console.log('Example User API query:');
  console.log(`
  query {
    me {
      id
      name
      email
      gender
      church {
        name
        country
      }
    }
    myCurrentProject {
      id
      name
      description
      startDate
      endDate
    }
    myProjects {
      id
      name
      events {
        id
        name
      }
    }
  }
`);
  console.log('');
  console.log('Example Admin API query:');
  console.log(`
  query {
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
`);
  console.log('');
  console.log('Example M2M API mutation:');
  console.log(`
  mutation {
    awardAchievement(userId: "user123", achievementId: "ach456") {
      id
      name
      points
    }
  }
`);
});
