import { createYoga } from 'graphql-yoga';
import { createServer } from 'node:http';
import { addMocksToSchema } from '@graphql-tools/mock';
import { makeExecutableSchema } from '@graphql-tools/schema';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { faker } from '@faker-js/faker';

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

  Project: () => {
    // Randomly choose project time state: past, active, or future
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

  Event: () => ({
    id: () => faker.string.uuid(),
    name: () => faker.helpers.arrayElement([
      `Week ${faker.number.int({ min: 1, max: 8 })}`,
      `${faker.word.adjective({ capitalize: true })} ${faker.helpers.arrayElement(['Ceremony', 'Challenge', 'Session', 'Gathering'])}`,
      `${faker.date.weekday()} ${faker.helpers.arrayElement(['Meeting', 'Service', 'Event'])}`
    ]),
    description: () => faker.lorem.sentence(),
    startDate: () => faker.date.recent({ days: 7 }).toISOString(),
    endDate: () => faker.date.soon({ days: 14 }).toISOString(),
  }),

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
    users: () => Array(100).fill(null).map(() => ({})),
    project: () => ({}),
    projects: () => Array(6).fill(null).map(() => ({})),
    event: () => ({}),
    events: () => Array(4).fill(null).map(() => ({})),
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
