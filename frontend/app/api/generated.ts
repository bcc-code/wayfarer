import gql from 'graphql-tag';
import * as Urql from '@urql/vue';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
export type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  DateTime: { input: any; output: any; }
  HTML: { input: any; output: any; }
};

export type Achievement = {
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  points: Scalars['Int']['output'];
  project: Project;
};

export type AgeRange = {
  max: Scalars['Int']['input'];
  min: Scalars['Int']['input'];
};

export type Article = {
  __typename?: 'Article';
  author: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  title: Scalars['String']['output'];
  url: Scalars['String']['output'];
};

export type Branding = {
  __typename?: 'Branding';
  colors: Colors;
  logo: Scalars['String']['output'];
};

export type Challenge = {
  buttonText: Scalars['String']['output'];
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  project: Project;
  publishedAt: Scalars['DateTime']['output'];
  url: Scalars['String']['output'];
  userCompletedAt?: Maybe<Scalars['DateTime']['output']>;
};

export type Church = {
  __typename?: 'Church';
  category: ChurchCategory;
  country: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export enum ChurchCategory {
  L = 'L',
  S = 'S',
  Xl = 'XL'
}

export type Colors = {
  __typename?: 'Colors';
  primary: Scalars['String']['output'];
  secondary: Scalars['String']['output'];
  tertiary: Scalars['String']['output'];
};

export type Event = {
  __typename?: 'Event';
  challenges: Array<Challenge>;
  description: Scalars['String']['output'];
  endDate: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  leaderboard: Array<LeaderboardEntry>;
  name: Scalars['String']['output'];
  parentProject: Project;
  startDate: Scalars['DateTime']['output'];
};


export type EventLeaderboardArgs = {
  filter?: InputMaybe<LeaderboardFilter>;
  type: LeaderboardType;
};

export enum Gender {
  Female = 'FEMALE',
  Male = 'MALE'
}

export type LeaderboardEntry = {
  __typename?: 'LeaderboardEntry';
  description?: Maybe<Scalars['String']['output']>;
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  score: Scalars['Int']['output'];
};

export type LeaderboardFilter = {
  ageRange?: InputMaybe<AgeRange>;
  churchCategory?: InputMaybe<ChurchCategory>;
  churchId?: InputMaybe<Scalars['ID']['input']>;
  country?: InputMaybe<Scalars['String']['input']>;
  gender?: InputMaybe<Gender>;
  maxScore?: InputMaybe<Scalars['Int']['input']>;
  minScore?: InputMaybe<Scalars['Int']['input']>;
  superTeamId?: InputMaybe<Scalars['ID']['input']>;
  teamId?: InputMaybe<Scalars['ID']['input']>;
};

export enum LeaderboardType {
  Monthly = 'MONTHLY',
  Total = 'TOTAL',
  Weekly = 'WEEKLY'
}

export type ListeningAchievement = Achievement & {
  __typename?: 'ListeningAchievement';
  achievedAt: Scalars['DateTime']['output'];
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  nextTrack: Track;
  points: Scalars['Int']['output'];
  project: Project;
  tracks: Array<Track>;
  userHasListened: Array<Track>;
};

export type M2MQueryRoot = {
  __typename?: 'M2MQueryRoot';
  achievement: Achievement;
  challenges: Array<Challenge>;
  currentEvent: Event;
  currentProject: Project;
  event: Event;
  project: Project;
  superteam: SuperTeam;
  team: Team;
  user: User;
};


export type M2MQueryRootAchievementArgs = {
  id: Scalars['ID']['input'];
};


export type M2MQueryRootChallengesArgs = {
  ids: Array<Scalars['ID']['input']>;
};


export type M2MQueryRootEventArgs = {
  id: Scalars['ID']['input'];
};


export type M2MQueryRootProjectArgs = {
  id: Scalars['ID']['input'];
};


export type M2MQueryRootSuperteamArgs = {
  id: Scalars['ID']['input'];
};


export type M2MQueryRootTeamArgs = {
  id: Scalars['ID']['input'];
};


export type M2MQueryRootUserArgs = {
  id: Scalars['ID']['input'];
};

export type Project = {
  __typename?: 'Project';
  achievements: Array<Achievement>;
  branding: Branding;
  challenges: Array<Challenge>;
  description: Scalars['String']['output'];
  endDate: Scalars['DateTime']['output'];
  events: Array<Event>;
  id: Scalars['ID']['output'];
  leaderboard: Array<LeaderboardEntry>;
  myTeam: Team;
  name: Scalars['String']['output'];
  startDate: Scalars['DateTime']['output'];
  teams: Array<Team>;
};


export type ProjectLeaderboardArgs = {
  filter?: InputMaybe<LeaderboardFilter>;
  type: LeaderboardType;
};

export type ReadingAchievement = Achievement & {
  __typename?: 'ReadingAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  articles: Array<Article>;
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  nextArticle: Article;
  points: Scalars['Int']['output'];
  project: Project;
  userHasRead: Array<Article>;
};

export type SimpleAchievement = Achievement & {
  __typename?: 'SimpleAchievement';
  achievedAt: Scalars['DateTime']['output'];
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  points: Scalars['Int']['output'];
  project: Project;
};

export type StreakAchievement = Achievement & {
  __typename?: 'StreakAchievement';
  achievedAt: Scalars['DateTime']['output'];
  challenge?: Maybe<Challenge>;
  currentStreak: Scalars['Int']['output'];
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  neededStreak: Scalars['Int']['output'];
  points: Scalars['Int']['output'];
  project: Project;
};

export type SuperTeam = {
  __typename?: 'SuperTeam';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  leaderboard: Array<LeaderboardEntry>;
  members: Array<User>;
  name: Scalars['String']['output'];
  parentProject: Project;
  teams: Array<Team>;
};


export type SuperTeamLeaderboardArgs = {
  filter?: InputMaybe<LeaderboardFilter>;
  type: LeaderboardType;
};

export type Team = {
  __typename?: 'Team';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  leaderboard: Array<LeaderboardEntry>;
  members: Array<User>;
  name: Scalars['String']['output'];
  parentProject: Project;
  superTeam?: Maybe<SuperTeam>;
};


export type TeamLeaderboardArgs = {
  filter?: InputMaybe<LeaderboardFilter>;
  type: LeaderboardType;
};

export type Track = {
  __typename?: 'Track';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
};

export type User = {
  __typename?: 'User';
  age: Scalars['Int']['output'];
  church: Church;
  email: Scalars['String']['output'];
  events: Array<Event>;
  gender: Gender;
  id: Scalars['ID']['output'];
  membersId: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  projects: Array<Project>;
  superTeams: Array<SuperTeam>;
  teams: Array<Team>;
};

export type UserQueryRoot = {
  __typename?: 'UserQueryRoot';
  challenges: Array<Challenge>;
  currentEvent: Event;
  currentProject: Project;
  events: Array<Event>;
  me: User;
  projects: Array<Project>;
};


export type UserQueryRootEventsArgs = {
  ids: Array<Scalars['ID']['input']>;
};


export type UserQueryRootProjectsArgs = {
  ids: Array<Scalars['ID']['input']>;
};

export type GetStandingsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetStandingsQuery = { __typename?: 'UserQueryRoot', me: { __typename?: 'User', id: string, name: string, email: string }, currentProject: { __typename?: 'Project', name: string, description: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', primary: string, secondary: string } } } };


export const GetStandingsDocument = gql`
    query GetStandings {
  me {
    id
    name
    email
  }
  currentProject {
    name
    description
    branding {
      colors {
        primary
        secondary
      }
    }
  }
}
    `;

export function useGetStandingsQuery(options?: Omit<Urql.UseQueryArgs<never, GetStandingsQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<GetStandingsQuery, GetStandingsQueryVariables | undefined>({ query: GetStandingsDocument, variables: undefined, ...options });
};