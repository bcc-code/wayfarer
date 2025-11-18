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
  Date: { input: any; output: any; }
  DateTime: { input: any; output: any; }
  HTML: { input: any; output: any; }
  Upload: { input: any; output: any; }
};

export type Achievement = {
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  points: Scalars['Int']['output'];
  project: Project;
};

export type AchievementConnection = {
  __typename?: 'AchievementConnection';
  edges: Array<AchievementEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type AchievementEdge = {
  __typename?: 'AchievementEdge';
  cursor: Scalars['String']['output'];
  node: Achievement;
};

export type AchievementFilter = {
  eventId?: InputMaybe<Scalars['ID']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
};

export type AgeRange = {
  __typename?: 'AgeRange';
  max: Scalars['Int']['output'];
  min: Scalars['Int']['output'];
};

export type AgeRangeInput = {
  max: Scalars['Int']['input'];
  min: Scalars['Int']['input'];
};

export type Article = {
  __typename?: 'Article';
  author: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  title: Scalars['String']['output'];
  url?: Maybe<Scalars['String']['output']>;
};

export type ArticleInput = {
  author: Scalars['String']['input'];
  title: Scalars['String']['input'];
  url?: InputMaybe<Scalars['String']['input']>;
};

export type AssignRoleInput = {
  role: RoleType;
  scopeId?: InputMaybe<Scalars['ID']['input']>;
  scopeType?: InputMaybe<ScopeType>;
  userId: Scalars['ID']['input'];
};

export type Branding = {
  __typename?: 'Branding';
  colors: Colors;
  logo?: Maybe<Scalars['String']['output']>;
  rounding: Scalars['Int']['output'];
};

export type BrandingInput = {
  colors: ColorsInput;
  logo?: InputMaybe<Scalars['String']['input']>;
  rounding: Scalars['Int']['input'];
};

export type Challenge = {
  __typename?: 'Challenge';
  buttonText: Scalars['String']['output'];
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  project: Project;
  publishedAt: Scalars['DateTime']['output'];
  url?: Maybe<Scalars['String']['output']>;
  userCompletedAt?: Maybe<Scalars['DateTime']['output']>;
};

export type ChallengeConnection = {
  __typename?: 'ChallengeConnection';
  edges: Array<ChallengeEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ChallengeEdge = {
  __typename?: 'ChallengeEdge';
  cursor: Scalars['String']['output'];
  node: Challenge;
};

export type ChallengeFilter = {
  eventId?: InputMaybe<Scalars['ID']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  publishedAfter?: InputMaybe<Scalars['DateTime']['input']>;
  publishedBefore?: InputMaybe<Scalars['DateTime']['input']>;
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

export type ChurchConnection = {
  __typename?: 'ChurchConnection';
  edges: Array<ChurchEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ChurchEdge = {
  __typename?: 'ChurchEdge';
  cursor: Scalars['String']['output'];
  node: Church;
};

export type ChurchFilter = {
  category?: InputMaybe<ChurchCategory>;
  country?: InputMaybe<Scalars['String']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type Colors = {
  __typename?: 'Colors';
  primary: Scalars['String']['output'];
  secondary: Scalars['String']['output'];
  tertiary: Scalars['String']['output'];
};

export type ColorsInput = {
  primary: Scalars['String']['input'];
  secondary: Scalars['String']['input'];
  tertiary: Scalars['String']['input'];
};

export type CreateChallengeInput = {
  buttonText: Scalars['String']['input'];
  description?: InputMaybe<Scalars['HTML']['input']>;
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  url?: InputMaybe<Scalars['String']['input']>;
};

export type CreateChurchInput = {
  category: ChurchCategory;
  country: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type CreateEventInput = {
  description: Scalars['String']['input'];
  endDate: Scalars['DateTime']['input'];
  name: Scalars['String']['input'];
  startDate: Scalars['DateTime']['input'];
};

export type CreateListeningAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  tracks: Array<TrackInput>;
};

export type CreateProjectInput = {
  branding: BrandingInput;
  description?: InputMaybe<Scalars['String']['input']>;
  endDate: Scalars['DateTime']['input'];
  name: Scalars['String']['input'];
  startDate: Scalars['DateTime']['input'];
};

export type CreateReadingAchievementInput = {
  articles: Array<ArticleInput>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateSimpleAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateStreakAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  neededStreak: Scalars['Int']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateStreakInput = {
  description: Scalars['String']['input'];
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  relevantDays: Array<DateRangeInput>;
};

export type CreateSuperTeamInput = {
  description: Scalars['String']['input'];
  name: Scalars['String']['input'];
  teamIds?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type CreateTeamInput = {
  description: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type CreateUserInput = {
  age: Scalars['Int']['input'];
  churchId: Scalars['ID']['input'];
  email: Scalars['String']['input'];
  gender: Gender;
  membersId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type DateRange = {
  __typename?: 'DateRange';
  end: Scalars['Date']['output'];
  start: Scalars['Date']['output'];
};

export type DateRangeInput = {
  end: Scalars['Date']['input'];
  start: Scalars['Date']['input'];
};

export type Event = {
  __typename?: 'Event';
  challenges: Array<Challenge>;
  description: Scalars['String']['output'];
  endDate: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  leaderboard: LeaderboardConnection;
  name: Scalars['String']['output'];
  parentProject: Project;
  startDate: Scalars['DateTime']['output'];
};


export type EventLeaderboardArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  entityType: LeaderboardEntityType;
  filter?: InputMaybe<LeaderboardFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};

export type EventConnection = {
  __typename?: 'EventConnection';
  edges: Array<EventEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type EventEdge = {
  __typename?: 'EventEdge';
  cursor: Scalars['String']['output'];
  node: Event;
};

export type EventFilter = {
  endDateAfter?: InputMaybe<Scalars['DateTime']['input']>;
  endDateBefore?: InputMaybe<Scalars['DateTime']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  startDateAfter?: InputMaybe<Scalars['DateTime']['input']>;
  startDateBefore?: InputMaybe<Scalars['DateTime']['input']>;
};

export enum ExportFormat {
  Csv = 'CSV',
  Excel = 'EXCEL',
  Json = 'JSON'
}

export enum Gender {
  Female = 'FEMALE',
  Male = 'MALE'
}

export type LeaderboardConnection = {
  __typename?: 'LeaderboardConnection';
  edges: Array<LeaderboardEdge>;
  me?: Maybe<LeaderboardEntry>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type LeaderboardEdge = {
  __typename?: 'LeaderboardEdge';
  cursor: Scalars['String']['output'];
  node: LeaderboardEntry;
};

export enum LeaderboardEntityType {
  Churches = 'CHURCHES',
  Persons = 'PERSONS',
  Superteams = 'SUPERTEAMS',
  Teams = 'TEAMS'
}

export type LeaderboardEntry = {
  __typename?: 'LeaderboardEntry';
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  isMe: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  rank: Scalars['Int']['output'];
  score: Scalars['Int']['output'];
};

export type LeaderboardFilter = {
  ageRange?: InputMaybe<AgeRangeInput>;
  churchCategory?: InputMaybe<ChurchCategory>;
  churchId?: InputMaybe<Scalars['ID']['input']>;
  country?: InputMaybe<Scalars['String']['input']>;
  gender?: InputMaybe<Gender>;
  maxScore?: InputMaybe<Scalars['Int']['input']>;
  minScore?: InputMaybe<Scalars['Int']['input']>;
  superTeamId?: InputMaybe<Scalars['ID']['input']>;
  teamId?: InputMaybe<Scalars['ID']['input']>;
};

export type ListeningAchievement = Achievement & {
  __typename?: 'ListeningAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  nextTrack: Track;
  points: Scalars['Int']['output'];
  project: Project;
  tracks: Array<Track>;
  userHasListened: Array<Track>;
};

export type Mutation = {
  __typename?: 'Mutation';
  addTeamMembers: Team;
  adjustSuperTeamScore: Scalars['Boolean']['output'];
  adjustTeamScore: Scalars['Boolean']['output'];
  adjustUserScore: Scalars['Boolean']['output'];
  archiveProject: Scalars['Boolean']['output'];
  assignChallengeToEvent: Challenge;
  assignRole: UserRole;
  assignTeamLead: Team;
  assignTeamsToSuperTeam: SuperTeam;
  assignUserToEvent: User;
  assignUserToProject: User;
  awardAchievement: Achievement;
  awardSuperTeamAchievement: Achievement;
  awardTeamAchievement: Achievement;
  bulkAwardAchievements: Array<Achievement>;
  bulkAwardSuperTeamAchievements: Array<Achievement>;
  bulkAwardTeamAchievements: Array<Achievement>;
  bulkCompleteChallenges: Array<Challenge>;
  bulkCreateChallenges: Array<Challenge>;
  bulkPublishChallenges: Array<Challenge>;
  completeChallenge: Challenge;
  createChallenge: Challenge;
  createEvent: Event;
  createListeningAchievement: ListeningAchievement;
  createProject: Project;
  createReadingAchievement: ReadingAchievement;
  createSimpleAchievement: SimpleAchievement;
  createStreak: Streak;
  createStreakAchievement: StreakAchievement;
  createSuperTeam: SuperTeam;
  createTeam: Team;
  deleteAchievement: Scalars['Boolean']['output'];
  deleteChallenge: Scalars['Boolean']['output'];
  deleteEvent: Scalars['Boolean']['output'];
  deleteProject: Scalars['Boolean']['output'];
  deleteStreak: Scalars['Boolean']['output'];
  deleteSuperTeam: Scalars['Boolean']['output'];
  deleteTeam: Scalars['Boolean']['output'];
  joinEvent: Event;
  joinProject: Project;
  joinTeam: Team;
  linkAchievementToChallenge: Achievement;
  markArticleAsRead: ReadingAchievement;
  markTrackAsListened: ListeningAchievement;
  moveEvent: Event;
  publishChallenge: Challenge;
  recordStreakActivity: StreakAchievement;
  regenerateJoinCode: Team;
  removeTeamMembers: Team;
  removeUserFromProject: User;
  revokeAchievement: Scalars['Boolean']['output'];
  revokeRole: Scalars['Boolean']['output'];
  revokeSuperTeamAchievement: Scalars['Boolean']['output'];
  revokeTeamAchievement: Scalars['Boolean']['output'];
  uncompleteChallenge: Scalars['Boolean']['output'];
  unmarkArticleAsRead: ReadingAchievement;
  unmarkTrackAsListened: ListeningAchievement;
  updateAchievement: Achievement;
  updateAvatar: User;
  updateChallenge: Challenge;
  updateEvent: Event;
  updateProject: Project;
  updateStreak: Streak;
  updateSuperTeam: SuperTeam;
  updateTeam: Team;
};


export type MutationAddTeamMembersArgs = {
  force?: InputMaybe<Scalars['Boolean']['input']>;
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type MutationAdjustSuperTeamScoreArgs = {
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  superTeamId: Scalars['ID']['input'];
};


export type MutationAdjustTeamScoreArgs = {
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  teamId: Scalars['ID']['input'];
};


export type MutationAdjustUserScoreArgs = {
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  userId: Scalars['ID']['input'];
};


export type MutationArchiveProjectArgs = {
  id: Scalars['ID']['input'];
};


export type MutationAssignChallengeToEventArgs = {
  challengeId: Scalars['ID']['input'];
  eventId: Scalars['ID']['input'];
};


export type MutationAssignRoleArgs = {
  input: AssignRoleInput;
};


export type MutationAssignTeamLeadArgs = {
  teamId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationAssignTeamsToSuperTeamArgs = {
  superTeamId: Scalars['ID']['input'];
  teamIds: Array<Scalars['ID']['input']>;
};


export type MutationAssignUserToEventArgs = {
  eventId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationAssignUserToProjectArgs = {
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationAwardAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationAwardSuperTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  superTeamId: Scalars['ID']['input'];
};


export type MutationAwardTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  teamId: Scalars['ID']['input'];
};


export type MutationBulkAwardAchievementsArgs = {
  achievementId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type MutationBulkAwardSuperTeamAchievementsArgs = {
  achievementId: Scalars['ID']['input'];
  superTeamIds: Array<Scalars['ID']['input']>;
};


export type MutationBulkAwardTeamAchievementsArgs = {
  achievementId: Scalars['ID']['input'];
  teamIds: Array<Scalars['ID']['input']>;
};


export type MutationBulkCompleteChallengesArgs = {
  challengeId: Scalars['ID']['input'];
  completedAt?: InputMaybe<Scalars['DateTime']['input']>;
  userIds: Array<Scalars['ID']['input']>;
};


export type MutationBulkCreateChallengesArgs = {
  eventId: Scalars['ID']['input'];
  inputs: Array<CreateChallengeInput>;
  projectId: Scalars['ID']['input'];
};


export type MutationBulkPublishChallengesArgs = {
  ids: Array<Scalars['ID']['input']>;
  publishedAt: Scalars['DateTime']['input'];
};


export type MutationCompleteChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  completedAt?: InputMaybe<Scalars['DateTime']['input']>;
  userId: Scalars['ID']['input'];
};


export type MutationCreateChallengeArgs = {
  eventId: Scalars['ID']['input'];
  input: CreateChallengeInput;
  projectId: Scalars['ID']['input'];
};


export type MutationCreateEventArgs = {
  input: CreateEventInput;
  projectId: Scalars['ID']['input'];
};


export type MutationCreateListeningAchievementArgs = {
  input: CreateListeningAchievementInput;
};


export type MutationCreateProjectArgs = {
  input: CreateProjectInput;
};


export type MutationCreateReadingAchievementArgs = {
  input: CreateReadingAchievementInput;
};


export type MutationCreateSimpleAchievementArgs = {
  input: CreateSimpleAchievementInput;
};


export type MutationCreateStreakArgs = {
  input: CreateStreakInput;
};


export type MutationCreateStreakAchievementArgs = {
  input: CreateStreakAchievementInput;
};


export type MutationCreateSuperTeamArgs = {
  input: CreateSuperTeamInput;
  projectId: Scalars['ID']['input'];
};


export type MutationCreateTeamArgs = {
  input: CreateTeamInput;
  projectId: Scalars['ID']['input'];
};


export type MutationDeleteAchievementArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteChallengeArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteEventArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteProjectArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteStreakArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteSuperTeamArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteTeamArgs = {
  id: Scalars['ID']['input'];
};


export type MutationJoinEventArgs = {
  eventId: Scalars['ID']['input'];
};


export type MutationJoinProjectArgs = {
  projectId: Scalars['ID']['input'];
};


export type MutationJoinTeamArgs = {
  code: Scalars['ID']['input'];
};


export type MutationLinkAchievementToChallengeArgs = {
  achievementId: Scalars['ID']['input'];
  challengeId: Scalars['ID']['input'];
};


export type MutationMarkArticleAsReadArgs = {
  achievementId: Scalars['ID']['input'];
  articleId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationMarkTrackAsListenedArgs = {
  achievementId: Scalars['ID']['input'];
  trackId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationMoveEventArgs = {
  id: Scalars['ID']['input'];
  newProjectId: Scalars['ID']['input'];
};


export type MutationPublishChallengeArgs = {
  id: Scalars['ID']['input'];
  publishedAt: Scalars['DateTime']['input'];
};


export type MutationRecordStreakActivityArgs = {
  achievementId: Scalars['ID']['input'];
  currentStreak: Scalars['Int']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationRegenerateJoinCodeArgs = {
  teamId: Scalars['ID']['input'];
};


export type MutationRemoveTeamMembersArgs = {
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type MutationRemoveUserFromProjectArgs = {
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationRevokeAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationRevokeRoleArgs = {
  input: RevokeRoleInput;
};


export type MutationRevokeSuperTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  superTeamId: Scalars['ID']['input'];
};


export type MutationRevokeTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  teamId: Scalars['ID']['input'];
};


export type MutationUncompleteChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationUnmarkArticleAsReadArgs = {
  achievementId: Scalars['ID']['input'];
  articleId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationUnmarkTrackAsListenedArgs = {
  achievementId: Scalars['ID']['input'];
  trackId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationUpdateAchievementArgs = {
  id: Scalars['ID']['input'];
  input: UpdateAchievementInput;
};


export type MutationUpdateAvatarArgs = {
  file: Scalars['Upload']['input'];
};


export type MutationUpdateChallengeArgs = {
  id: Scalars['ID']['input'];
  input: UpdateChallengeInput;
};


export type MutationUpdateEventArgs = {
  id: Scalars['ID']['input'];
  input: UpdateEventInput;
};


export type MutationUpdateProjectArgs = {
  id: Scalars['ID']['input'];
  input: UpdateProjectInput;
};


export type MutationUpdateStreakArgs = {
  id: Scalars['ID']['input'];
  input: UpdateStreakInput;
};


export type MutationUpdateSuperTeamArgs = {
  id: Scalars['ID']['input'];
  input: UpdateSuperTeamInput;
};


export type MutationUpdateTeamArgs = {
  id: Scalars['ID']['input'];
  input: UpdateTeamInput;
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor?: Maybe<Scalars['String']['output']>;
};

export type Project = {
  __typename?: 'Project';
  achievements: Array<Achievement>;
  archivedAt?: Maybe<Scalars['Boolean']['output']>;
  branding: Branding;
  challenges: Array<Challenge>;
  description: Scalars['String']['output'];
  endDate: Scalars['DateTime']['output'];
  events: Array<Event>;
  id: Scalars['ID']['output'];
  leaderboard: LeaderboardConnection;
  myTeam?: Maybe<Team>;
  name: Scalars['String']['output'];
  startDate: Scalars['DateTime']['output'];
  streaks: Array<Streak>;
  teams: Array<Team>;
};


export type ProjectLeaderboardArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  entityType: LeaderboardEntityType;
  filter?: InputMaybe<LeaderboardFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};

export type ProjectConnection = {
  __typename?: 'ProjectConnection';
  edges: Array<ProjectEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ProjectEdge = {
  __typename?: 'ProjectEdge';
  cursor: Scalars['String']['output'];
  node: Project;
};

export type ProjectFilter = {
  archived?: InputMaybe<Scalars['Boolean']['input']>;
  endDateAfter?: InputMaybe<Scalars['DateTime']['input']>;
  endDateBefore?: InputMaybe<Scalars['DateTime']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  startDateAfter?: InputMaybe<Scalars['DateTime']['input']>;
  startDateBefore?: InputMaybe<Scalars['DateTime']['input']>;
};

export type Query = {
  __typename?: 'Query';
  achievement: Achievement;
  achievements: AchievementConnection;
  challenge: Challenge;
  challenges: ChallengeConnection;
  church: Church;
  churches: ChurchConnection;
  currentEvent: Event;
  currentProject: Project;
  event: Event;
  events: EventConnection;
  me: User;
  myCurrentEvent: Event;
  myCurrentProject: Project;
  myEvents: Array<Event>;
  myProjects: Array<Project>;
  project: Project;
  projects: ProjectConnection;
  streak: Streak;
  streaks: StreakConnection;
  superteam: SuperTeam;
  superteams: SuperTeamConnection;
  team: Team;
  teams: TeamConnection;
  user: User;
  userRoles: Array<UserRole>;
  users: UserConnection;
  usersWithRole: Array<User>;
};


export type QueryAchievementArgs = {
  id: Scalars['ID']['input'];
};


export type QueryAchievementsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter: AchievementFilter;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryChallengeArgs = {
  id: Scalars['ID']['input'];
};


export type QueryChallengesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<ChallengeFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryChurchArgs = {
  id: Scalars['ID']['input'];
};


export type QueryChurchesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<ChurchFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryEventArgs = {
  id: Scalars['ID']['input'];
};


export type QueryEventsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<EventFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryMyEventsArgs = {
  project?: InputMaybe<Scalars['ID']['input']>;
};


export type QueryProjectArgs = {
  id: Scalars['ID']['input'];
};


export type QueryProjectsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<ProjectFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryStreakArgs = {
  id: Scalars['ID']['input'];
};


export type QueryStreaksArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<StreakFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QuerySuperteamArgs = {
  id: Scalars['ID']['input'];
};


export type QuerySuperteamsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<SuperTeamFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryTeamArgs = {
  id: Scalars['ID']['input'];
};


export type QueryTeamsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<TeamFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryUserArgs = {
  id: Scalars['ID']['input'];
};


export type QueryUserRolesArgs = {
  userId: Scalars['ID']['input'];
};


export type QueryUsersArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<UserFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryUsersWithRoleArgs = {
  role: RoleType;
  scopeId?: InputMaybe<Scalars['ID']['input']>;
  scopeType?: InputMaybe<ScopeType>;
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
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  nextArticle: Article;
  points: Scalars['Int']['output'];
  project: Project;
  userHasRead: Array<Article>;
};

export type RevokeRoleInput = {
  role: RoleType;
  scopeId?: InputMaybe<Scalars['ID']['input']>;
  scopeType?: InputMaybe<ScopeType>;
  userId: Scalars['ID']['input'];
};

export type RoleScope = {
  __typename?: 'RoleScope';
  church?: Maybe<Church>;
  id: Scalars['ID']['output'];
  project?: Maybe<Project>;
  team?: Maybe<Team>;
  type: ScopeType;
};

export enum RoleType {
  Admin = 'ADMIN',
  ChurchAdmin = 'CHURCH_ADMIN',
  M2M = 'M2M',
  ProjectAdmin = 'PROJECT_ADMIN',
  Superadmin = 'SUPERADMIN',
  TeamLead = 'TEAM_LEAD',
  User = 'USER'
}

export enum ScopeType {
  Church = 'CHURCH',
  Project = 'PROJECT',
  Team = 'TEAM'
}

export type SimpleAchievement = Achievement & {
  __typename?: 'SimpleAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  points: Scalars['Int']['output'];
  project: Project;
};

export type Streak = {
  __typename?: 'Streak';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  listenedDays: Array<StreakDay>;
  name: Scalars['String']['output'];
  project: Project;
  relevantDays: Array<DateRange>;
  status: Scalars['Int']['output'];
};


export type StreakListenedDaysArgs = {
  last: Scalars['Int']['input'];
};

export type StreakAchievement = Achievement & {
  __typename?: 'StreakAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  neededStreak: Scalars['Int']['output'];
  points: Scalars['Int']['output'];
  project: Project;
  streak: Streak;
};

export type StreakConnection = {
  __typename?: 'StreakConnection';
  edges: Array<StreakEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type StreakDay = {
  __typename?: 'StreakDay';
  active: Scalars['Boolean']['output'];
  date: Scalars['Date']['output'];
};

export type StreakEdge = {
  __typename?: 'StreakEdge';
  cursor: Scalars['String']['output'];
  node: Streak;
};

export type StreakFilter = {
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
};

export type SuperTeam = {
  __typename?: 'SuperTeam';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  members: UserConnection;
  name: Scalars['String']['output'];
  parentProject: Project;
  teams: Array<Team>;
};


export type SuperTeamMembersArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};

export type SuperTeamConnection = {
  __typename?: 'SuperTeamConnection';
  edges: Array<SuperTeamEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type SuperTeamEdge = {
  __typename?: 'SuperTeamEdge';
  cursor: Scalars['String']['output'];
  node: SuperTeam;
};

export type SuperTeamFilter = {
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  maxMembers?: InputMaybe<Scalars['Int']['input']>;
  maxTeams?: InputMaybe<Scalars['Int']['input']>;
  minMembers?: InputMaybe<Scalars['Int']['input']>;
  minTeams?: InputMaybe<Scalars['Int']['input']>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
};

export type Team = {
  __typename?: 'Team';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  joinCode: Scalars['String']['output'];
  members: Array<User>;
  name: Scalars['String']['output'];
  parentProject: Project;
  superTeam?: Maybe<SuperTeam>;
};

export type TeamConnection = {
  __typename?: 'TeamConnection';
  edges: Array<TeamEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type TeamEdge = {
  __typename?: 'TeamEdge';
  cursor: Scalars['String']['output'];
  node: Team;
};

export type TeamFilter = {
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  maxMembers?: InputMaybe<Scalars['Int']['input']>;
  minMembers?: InputMaybe<Scalars['Int']['input']>;
  noSuperTeam?: InputMaybe<Scalars['Boolean']['input']>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  superTeamId?: InputMaybe<Scalars['ID']['input']>;
};

export type Track = {
  __typename?: 'Track';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
};

export type TrackInput = {
  description: Scalars['String']['input'];
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
};

export type UpdateAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden?: InputMaybe<Scalars['Boolean']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  points?: InputMaybe<Scalars['Int']['input']>;
};

export type UpdateChallengeInput = {
  buttonText?: InputMaybe<Scalars['String']['input']>;
  description?: InputMaybe<Scalars['HTML']['input']>;
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  url?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateChurchInput = {
  category?: InputMaybe<ChurchCategory>;
  country?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateEventInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  endDate?: InputMaybe<Scalars['DateTime']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  startDate?: InputMaybe<Scalars['DateTime']['input']>;
};

export type UpdateProjectInput = {
  branding?: InputMaybe<BrandingInput>;
  description?: InputMaybe<Scalars['String']['input']>;
  endDate?: InputMaybe<Scalars['DateTime']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  startDate?: InputMaybe<Scalars['DateTime']['input']>;
};

export type UpdateStreakInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  relevantDays?: InputMaybe<Array<DateRangeInput>>;
};

export type UpdateSuperTeamInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateTeamInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type User = {
  __typename?: 'User';
  age?: Maybe<Scalars['Int']['output']>;
  birthdate: Scalars['String']['output'];
  church: Church;
  churchId: Scalars['ID']['output'];
  email: Scalars['String']['output'];
  events: Array<Event>;
  gender: Gender;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  membersId: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  projects: Array<Project>;
  roles: Array<UserRole>;
  superTeams: Array<SuperTeam>;
  teams: Array<Team>;
};

export type UserConnection = {
  __typename?: 'UserConnection';
  edges: Array<UserEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type UserEdge = {
  __typename?: 'UserEdge';
  cursor: Scalars['String']['output'];
  node: User;
};

export type UserFilter = {
  churchId?: InputMaybe<Scalars['ID']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  gender?: InputMaybe<Gender>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  maxAge?: InputMaybe<Scalars['Int']['input']>;
  minAge?: InputMaybe<Scalars['Int']['input']>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  teamId?: InputMaybe<Scalars['ID']['input']>;
};

export type UserRole = {
  __typename?: 'UserRole';
  id: Scalars['ID']['output'];
  role: RoleType;
  scope?: Maybe<RoleScope>;
  user: User;
};

export type GetMeQueryVariables = Exact<{ [key: string]: never; }>;


export type GetMeQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string, email: string, image?: string | null, membersId: string, gender: Gender, birthdate: string, church: { __typename?: 'Church', id: string, name: string, country: string, category: ChurchCategory }, roles: Array<{ __typename?: 'UserRole', id: string, role: RoleType, scope?: { __typename?: 'RoleScope', id: string, type: ScopeType, church?: { __typename?: 'Church', id: string } | null, team?: { __typename?: 'Team', id: string } | null, project?: { __typename?: 'Project', id: string } | null } | null }> } };

export type AdminSidebarQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminSidebarQuery = { __typename?: 'Query', projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, endDate: any, startDate: any } }> } };

export type CurrentProjectQueryVariables = Exact<{ [key: string]: never; }>;


export type CurrentProjectQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', primary: string } } } };

export type AdminHomePageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminHomePageQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string }, projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, description: string, endDate: any, startDate: any, branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', primary: string, secondary: string, tertiary: string } } } }> } };

export type AdminProjectEditPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectEditPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, description: string, startDate: any, endDate: any, archivedAt?: boolean | null, branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', primary: string } } } };

export type UpdateProjectMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateProjectInput;
}>;


export type UpdateProjectMutation = { __typename?: 'Mutation', updateProject: { __typename?: 'Project', id: string } };

export type AdminProjectPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, description: string, startDate: any, endDate: any, branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', primary: string } } }, achievements: { __typename?: 'AchievementConnection', edges: Array<{ __typename?: 'AchievementEdge', node:
        | { __typename?: 'ListeningAchievement', id: string, name: string }
        | { __typename?: 'ReadingAchievement', id: string, name: string }
        | { __typename?: 'SimpleAchievement', id: string, name: string }
        | { __typename?: 'StreakAchievement', id: string, name: string }
       }> }, events: { __typename?: 'EventConnection', edges: Array<{ __typename?: 'EventEdge', node: { __typename?: 'Event', id: string, name: string } }> }, challenges: { __typename?: 'ChallengeConnection', edges: Array<{ __typename?: 'ChallengeEdge', node: { __typename?: 'Challenge', id: string, name: string } }> }, streaks: { __typename?: 'StreakConnection', edges: Array<{ __typename?: 'StreakEdge', node: { __typename?: 'Streak', id: string, name: string } }> } };

export type AdminProjectsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminProjectsPageQuery = { __typename?: 'Query', projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, description: string, endDate: any, startDate: any, branding: { __typename?: 'Branding', logo?: string | null, colors: { __typename?: 'Colors', primary: string } } } }> } };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'Project', id: string } };

export type AdminUserPageQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type AdminUserPageQuery = { __typename?: 'Query', user: { __typename?: 'User', id: string, name: string, email: string, membersId: string, gender: Gender, birthdate: string, age?: number | null, image?: string | null, church: { __typename?: 'Church', id: string, name: string }, roles: Array<{ __typename?: 'UserRole', id: string, role: RoleType, scope?: { __typename?: 'RoleScope', id: string, type: ScopeType } | null }> } };

export type AdminUsersPageQueryVariables = Exact<{
  filter?: InputMaybe<UserFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
}>;


export type AdminUsersPageQuery = { __typename?: 'Query', users: { __typename?: 'UserConnection', totalCount: number, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null }, edges: Array<{ __typename?: 'UserEdge', cursor: string, node: { __typename?: 'User', id: string, name: string, email: string, image?: string | null, church: { __typename?: 'Church', name: string }, roles: Array<{ __typename?: 'UserRole', id: string, role: RoleType }> } }> } };

export type ChallengesPageQueryVariables = Exact<{ [key: string]: never; }>;


export type ChallengesPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', challenges: Array<{ __typename?: 'Challenge', id: string, name: string, description: any, userCompletedAt?: any | null, image?: string | null, url?: string | null, buttonText: string, publishedAt: any, endTime?: any | null }> } };

export type StandingsPageQueryVariables = Exact<{
  entityType: LeaderboardEntityType;
  filter?: InputMaybe<LeaderboardFilter>;
}>;


export type StandingsPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', id: string, leaderboard: { __typename?: 'LeaderboardConnection', edges: Array<{ __typename?: 'LeaderboardEdge', node: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, image?: string | null, rank: number, isMe: boolean } }>, me?: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank: number, isMe: boolean, image?: string | null } | null } } };

export type ProfilePageQueryVariables = Exact<{ [key: string]: never; }>;


export type ProfilePageQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string, image?: string | null, projects: Array<{ __typename?: 'Project', id: string, achievements: Array<
        | { __typename?: 'ListeningAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
        | { __typename?: 'ReadingAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
        | { __typename?: 'SimpleAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
        | { __typename?: 'StreakAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
      > }> } };

export type UnitPageQueryVariables = Exact<{ [key: string]: never; }>;


export type UnitPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', id: string, myTeam?: { __typename?: 'Team', id: string, name: string, superTeam?: { __typename?: 'SuperTeam', id: string, name: string } | null } | null } };


export const GetMeDocument = gql`
    query GetMe {
  me {
    id
    name
    email
    image
    membersId
    church {
      id
      name
      country
      category
    }
    gender
    birthdate
    roles {
      id
      role
      scope {
        id
        type
        church {
          id
        }
        team {
          id
        }
        project {
          id
        }
      }
    }
  }
}
    `;

export function useGetMeQuery(options?: Omit<Urql.UseQueryArgs<never, GetMeQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<GetMeQuery, GetMeQueryVariables | undefined>({ query: GetMeDocument, variables: undefined, ...options });
};
export const AdminSidebarDocument = gql`
    query AdminSidebar {
  projects {
    edges {
      node {
        id
        name
        endDate
        startDate
      }
    }
  }
}
    `;

export function useAdminSidebarQuery(options?: Omit<Urql.UseQueryArgs<never, AdminSidebarQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminSidebarQuery, AdminSidebarQueryVariables | undefined>({ query: AdminSidebarDocument, variables: undefined, ...options });
};
export const CurrentProjectDocument = gql`
    query CurrentProject {
  myCurrentProject {
    branding {
      logo
      colors {
        primary
      }
      rounding
    }
  }
}
    `;

export function useCurrentProjectQuery(options?: Omit<Urql.UseQueryArgs<never, CurrentProjectQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<CurrentProjectQuery, CurrentProjectQueryVariables | undefined>({ query: CurrentProjectDocument, variables: undefined, ...options });
};
export const AdminHomePageDocument = gql`
    query AdminHomePage {
  me {
    id
    name
  }
  projects {
    edges {
      node {
        id
        name
        description
        endDate
        startDate
        branding {
          logo
          rounding
          colors {
            primary
            secondary
            tertiary
          }
        }
      }
    }
  }
}
    `;

export function useAdminHomePageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminHomePageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminHomePageQuery, AdminHomePageQueryVariables | undefined>({ query: AdminHomePageDocument, variables: undefined, ...options });
};
export const AdminProjectEditPageDocument = gql`
    query AdminProjectEditPage($projectId: ID!) {
  project(id: $projectId) {
    id
    name
    description
    startDate
    endDate
    archivedAt
    branding {
      logo
      rounding
      colors {
        primary
      }
    }
  }
}
    `;

export function useAdminProjectEditPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectEditPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectEditPageQuery, AdminProjectEditPageQueryVariables | undefined>({ query: AdminProjectEditPageDocument, variables: undefined, ...options });
};
export const UpdateProjectDocument = gql`
    mutation UpdateProject($id: ID!, $input: UpdateProjectInput!) {
  updateProject(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateProjectMutation() {
  return Urql.useMutation<UpdateProjectMutation, UpdateProjectMutationVariables>(UpdateProjectDocument);
};
export const AdminProjectPageDocument = gql`
    query AdminProjectPage($projectId: ID!) {
  project(id: $projectId) {
    id
    name
    description
    startDate
    endDate
    branding {
      logo
      rounding
      colors {
        primary
      }
    }
  }
  achievements(filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
      }
    }
  }
  events(filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
      }
    }
  }
  challenges(filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
      }
    }
  }
  streaks(filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
      }
    }
  }
}
    `;

export function useAdminProjectPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectPageQuery, AdminProjectPageQueryVariables | undefined>({ query: AdminProjectPageDocument, variables: undefined, ...options });
};
export const AdminProjectsPageDocument = gql`
    query AdminProjectsPage {
  projects(first: 100) {
    edges {
      node {
        id
        name
        description
        endDate
        startDate
        branding {
          logo
          colors {
            primary
          }
        }
      }
    }
  }
}
    `;

export function useAdminProjectsPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectsPageQuery, AdminProjectsPageQueryVariables | undefined>({ query: AdminProjectsPageDocument, variables: undefined, ...options });
};
export const CreateProjectDocument = gql`
    mutation CreateProject($input: CreateProjectInput!) {
  createProject(input: $input) {
    id
  }
}
    `;

export function useCreateProjectMutation() {
  return Urql.useMutation<CreateProjectMutation, CreateProjectMutationVariables>(CreateProjectDocument);
};
export const AdminUserPageDocument = gql`
    query AdminUserPage($id: ID!) {
  user(id: $id) {
    id
    name
    email
    membersId
    gender
    birthdate
    age
    image
    church {
      id
      name
    }
    roles {
      id
      role
      scope {
        id
        type
      }
    }
  }
}
    `;

export function useAdminUserPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminUserPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminUserPageQuery, AdminUserPageQueryVariables | undefined>({ query: AdminUserPageDocument, variables: undefined, ...options });
};
export const AdminUsersPageDocument = gql`
    query AdminUsersPage($filter: UserFilter, $first: Int, $after: String, $last: Int, $before: String) {
  users(
    filter: $filter
    first: $first
    after: $after
    last: $last
    before: $before
  ) {
    totalCount
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    edges {
      cursor
      node {
        id
        name
        email
        image
        church {
          name
        }
        roles {
          id
          role
        }
      }
    }
  }
}
    `;

export function useAdminUsersPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminUsersPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminUsersPageQuery, AdminUsersPageQueryVariables | undefined>({ query: AdminUsersPageDocument, variables: undefined, ...options });
};
export const ChallengesPageDocument = gql`
    query ChallengesPage {
  myCurrentProject {
    challenges {
      id
      name
      description
      userCompletedAt
      image
      url
      buttonText
      publishedAt
      endTime
    }
  }
}
    `;

export function useChallengesPageQuery(options?: Omit<Urql.UseQueryArgs<never, ChallengesPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ChallengesPageQuery, ChallengesPageQueryVariables | undefined>({ query: ChallengesPageDocument, variables: undefined, ...options });
};
export const StandingsPageDocument = gql`
    query StandingsPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter) {
  myCurrentProject {
    id
    leaderboard(entityType: $entityType, filter: $filter) {
      edges {
        node {
          id
          name
          score
          image
          rank
          isMe
        }
      }
      me {
        id
        name
        score
        rank
        isMe
        image
      }
    }
  }
}
    `;

export function useStandingsPageQuery(options?: Omit<Urql.UseQueryArgs<never, StandingsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<StandingsPageQuery, StandingsPageQueryVariables | undefined>({ query: StandingsPageDocument, variables: undefined, ...options });
};
export const ProfilePageDocument = gql`
    query ProfilePage {
  me {
    id
    name
    image
    projects {
      id
      achievements {
        id
        name
        description
        image
        hidden
        achievedAt
        points
      }
    }
  }
}
    `;

export function useProfilePageQuery(options?: Omit<Urql.UseQueryArgs<never, ProfilePageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ProfilePageQuery, ProfilePageQueryVariables | undefined>({ query: ProfilePageDocument, variables: undefined, ...options });
};
export const UnitPageDocument = gql`
    query UnitPage {
  myCurrentProject {
    id
    myTeam {
      id
      name
      superTeam {
        id
        name
      }
    }
  }
}
    `;

export function useUnitPageQuery(options?: Omit<Urql.UseQueryArgs<never, UnitPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<UnitPageQuery, UnitPageQueryVariables | undefined>({ query: UnitPageDocument, variables: undefined, ...options });
};