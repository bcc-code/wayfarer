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
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  points: Scalars['Int']['output'];
  project: Project;
};

export type AdminMutationRoot = {
  __typename?: 'AdminMutationRoot';
  addTeamMembers: Team;
  adjustUserScore: Scalars['Boolean']['output'];
  archiveProject: Scalars['Boolean']['output'];
  assignChallengeToEvent: Challenge;
  assignTeamsToSuperTeam: SuperTeam;
  assignUserToEvent: User;
  assignUserToProject: User;
  bulkAssignUsersToTeam: Team;
  bulkCreateChallenges: Array<Challenge>;
  bulkPublishChallenges: Array<Challenge>;
  cloneProject: Project;
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
  linkAchievementToChallenge: Achievement;
  moveEvent: Event;
  publishChallenge: Challenge;
  removeTeamMembers: Team;
  removeUserFromProject: User;
  updateAchievement: Achievement;
  updateChallenge: Challenge;
  updateEvent: Event;
  updateProject: Project;
  updateStreak: Streak;
  updateSuperTeam: SuperTeam;
  updateTeam: Team;
};


export type AdminMutationRootAddTeamMembersArgs = {
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type AdminMutationRootAdjustUserScoreArgs = {
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type AdminMutationRootArchiveProjectArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootAssignChallengeToEventArgs = {
  challengeId: Scalars['ID']['input'];
  eventId: Scalars['ID']['input'];
};


export type AdminMutationRootAssignTeamsToSuperTeamArgs = {
  superTeamId: Scalars['ID']['input'];
  teamIds: Array<Scalars['ID']['input']>;
};


export type AdminMutationRootAssignUserToEventArgs = {
  eventId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type AdminMutationRootAssignUserToProjectArgs = {
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type AdminMutationRootBulkAssignUsersToTeamArgs = {
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type AdminMutationRootBulkCreateChallengesArgs = {
  inputs: Array<CreateChallengeInput>;
};


export type AdminMutationRootBulkPublishChallengesArgs = {
  ids: Array<Scalars['ID']['input']>;
  publishedAt: Scalars['DateTime']['input'];
};


export type AdminMutationRootCloneProjectArgs = {
  id: Scalars['ID']['input'];
  newName: Scalars['String']['input'];
};


export type AdminMutationRootCreateChallengeArgs = {
  input: CreateChallengeInput;
};


export type AdminMutationRootCreateEventArgs = {
  input: CreateEventInput;
  projectId: Scalars['ID']['input'];
};


export type AdminMutationRootCreateListeningAchievementArgs = {
  input: CreateListeningAchievementInput;
};


export type AdminMutationRootCreateProjectArgs = {
  input: CreateProjectInput;
};


export type AdminMutationRootCreateReadingAchievementArgs = {
  input: CreateReadingAchievementInput;
};


export type AdminMutationRootCreateSimpleAchievementArgs = {
  input: CreateSimpleAchievementInput;
};


export type AdminMutationRootCreateStreakArgs = {
  input: CreateStreakInput;
};


export type AdminMutationRootCreateStreakAchievementArgs = {
  input: CreateStreakAchievementInput;
};


export type AdminMutationRootCreateSuperTeamArgs = {
  input: CreateSuperTeamInput;
  projectId: Scalars['ID']['input'];
};


export type AdminMutationRootCreateTeamArgs = {
  input: CreateTeamInput;
  projectId: Scalars['ID']['input'];
};


export type AdminMutationRootDeleteAchievementArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootDeleteChallengeArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootDeleteEventArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootDeleteProjectArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootDeleteStreakArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootDeleteSuperTeamArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootDeleteTeamArgs = {
  id: Scalars['ID']['input'];
};


export type AdminMutationRootLinkAchievementToChallengeArgs = {
  achievementId: Scalars['ID']['input'];
  challengeId: Scalars['ID']['input'];
};


export type AdminMutationRootMoveEventArgs = {
  id: Scalars['ID']['input'];
  newProjectId: Scalars['ID']['input'];
};


export type AdminMutationRootPublishChallengeArgs = {
  id: Scalars['ID']['input'];
  publishedAt: Scalars['DateTime']['input'];
};


export type AdminMutationRootRemoveTeamMembersArgs = {
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type AdminMutationRootRemoveUserFromProjectArgs = {
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type AdminMutationRootUpdateAchievementArgs = {
  id: Scalars['ID']['input'];
  input: UpdateAchievementInput;
};


export type AdminMutationRootUpdateChallengeArgs = {
  id: Scalars['ID']['input'];
  input: UpdateChallengeInput;
};


export type AdminMutationRootUpdateEventArgs = {
  id: Scalars['ID']['input'];
  input: UpdateEventInput;
};


export type AdminMutationRootUpdateProjectArgs = {
  id: Scalars['ID']['input'];
  input: UpdateProjectInput;
};


export type AdminMutationRootUpdateStreakArgs = {
  id: Scalars['ID']['input'];
  input: UpdateStreakInput;
};


export type AdminMutationRootUpdateSuperTeamArgs = {
  id: Scalars['ID']['input'];
  input: UpdateSuperTeamInput;
};


export type AdminMutationRootUpdateTeamArgs = {
  id: Scalars['ID']['input'];
  input: UpdateTeamInput;
};

export type AdminQueryRoot = {
  __typename?: 'AdminQueryRoot';
  achievement: Achievement;
  achievements: Array<Achievement>;
  challenge: Challenge;
  challenges: Array<Challenge>;
  church: Church;
  churches: Array<Church>;
  currentEvent: Event;
  currentProject: Project;
  event: Event;
  events: Array<Event>;
  project: Project;
  projects: Array<Project>;
  streak: Streak;
  streaks: Array<Streak>;
  superteam: SuperTeam;
  superteams: Array<SuperTeam>;
  team: Team;
  teams: Array<Team>;
  user: User;
  users: Array<User>;
};


export type AdminQueryRootAchievementArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootAchievementsArgs = {
  eventId?: InputMaybe<Scalars['ID']['input']>;
  projectId: Scalars['ID']['input'];
};


export type AdminQueryRootChallengeArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootChallengesArgs = {
  eventId?: InputMaybe<Scalars['ID']['input']>;
  projectId: Scalars['ID']['input'];
};


export type AdminQueryRootChurchArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootEventArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootEventsArgs = {
  projectId?: InputMaybe<Scalars['ID']['input']>;
};


export type AdminQueryRootProjectArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootStreakArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootStreaksArgs = {
  projectId?: InputMaybe<Scalars['ID']['input']>;
};


export type AdminQueryRootSuperteamArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootSuperteamsArgs = {
  projectId?: InputMaybe<Scalars['ID']['input']>;
};


export type AdminQueryRootTeamArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootTeamsArgs = {
  projectId?: InputMaybe<Scalars['ID']['input']>;
};


export type AdminQueryRootUserArgs = {
  id: Scalars['ID']['input'];
};


export type AdminQueryRootUsersArgs = {
  filter?: InputMaybe<UserFilter>;
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
  url: Scalars['String']['output'];
};

export type ArticleInput = {
  author: Scalars['String']['input'];
  title: Scalars['String']['input'];
  url: Scalars['String']['input'];
};

export type Branding = {
  __typename?: 'Branding';
  colors: Colors;
  logo: Scalars['String']['output'];
  rounding: Scalars['Int']['output'];
};

export type BrandingInput = {
  colors: ColorsInput;
  logo: Scalars['String']['input'];
};

export type Challenge = {
  __typename?: 'Challenge';
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

export type ColorsInput = {
  primary: Scalars['String']['input'];
  secondary: Scalars['String']['input'];
  tertiary: Scalars['String']['input'];
};

export type CombinedMutation = {
  __typename?: 'CombinedMutation';
  admin: AdminMutationRoot;
  m2m: M2MMutationRoot;
  user: UserMutationRoot;
};

export type CombinedQuery = {
  __typename?: 'CombinedQuery';
  admin: AdminQueryRoot;
  m2m: M2MQueryRoot;
  user: UserQueryRoot;
};

export type CreateChallengeInput = {
  buttonText: Scalars['String']['input'];
  description: Scalars['HTML']['input'];
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  image: Scalars['String']['input'];
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  url: Scalars['String']['input'];
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
  image: Scalars['String']['input'];
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  tracks: Array<TrackInput>;
};

export type CreateProjectInput = {
  branding: BrandingInput;
  description: Scalars['String']['input'];
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
  image: Scalars['String']['input'];
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateSimpleAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  image: Scalars['String']['input'];
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateStreakAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  image: Scalars['String']['input'];
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
  leaderboard: Array<LeaderboardEntry>;
  name: Scalars['String']['output'];
  parentProject: Project;
  startDate: Scalars['DateTime']['output'];
};


export type EventLeaderboardArgs = {
  filter?: InputMaybe<LeaderboardFilter>;
  type: LeaderboardType;
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

export type LeaderboardEntry = {
  __typename?: 'LeaderboardEntry';
  description?: Maybe<Scalars['String']['output']>;
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
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

export enum LeaderboardType {
  Monthly = 'MONTHLY',
  Total = 'TOTAL',
  Weekly = 'WEEKLY'
}

export type ListeningAchievement = Achievement & {
  __typename?: 'ListeningAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
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

export type M2MMutationRoot = {
  __typename?: 'M2MMutationRoot';
  adjustSuperTeamScore: Scalars['Boolean']['output'];
  adjustTeamScore: Scalars['Boolean']['output'];
  adjustUserScore: Scalars['Boolean']['output'];
  awardAchievement: Achievement;
  awardSuperTeamAchievement: Achievement;
  awardTeamAchievement: Achievement;
  bulkAwardAchievements: Array<Achievement>;
  bulkAwardSuperTeamAchievements: Array<Achievement>;
  bulkAwardTeamAchievements: Array<Achievement>;
  bulkCompleteChallenges: Array<Challenge>;
  completeChallenge: Challenge;
  markArticleAsRead: ReadingAchievement;
  markTrackAsListened: ListeningAchievement;
  revokeAchievement: Scalars['Boolean']['output'];
  revokeSuperTeamAchievement: Scalars['Boolean']['output'];
  revokeTeamAchievement: Scalars['Boolean']['output'];
  uncompleteChallenge: Scalars['Boolean']['output'];
  unmarkArticleAsRead: ReadingAchievement;
  unmarkTrackAsListened: ListeningAchievement;
  updateStreak: StreakAchievement;
};


export type M2MMutationRootAdjustSuperTeamScoreArgs = {
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  superTeamId: Scalars['ID']['input'];
};


export type M2MMutationRootAdjustTeamScoreArgs = {
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  teamId: Scalars['ID']['input'];
};


export type M2MMutationRootAdjustUserScoreArgs = {
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootAwardAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootAwardSuperTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  superTeamId: Scalars['ID']['input'];
};


export type M2MMutationRootAwardTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  teamId: Scalars['ID']['input'];
};


export type M2MMutationRootBulkAwardAchievementsArgs = {
  achievementId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type M2MMutationRootBulkAwardSuperTeamAchievementsArgs = {
  achievementId: Scalars['ID']['input'];
  superTeamIds: Array<Scalars['ID']['input']>;
};


export type M2MMutationRootBulkAwardTeamAchievementsArgs = {
  achievementId: Scalars['ID']['input'];
  teamIds: Array<Scalars['ID']['input']>;
};


export type M2MMutationRootBulkCompleteChallengesArgs = {
  challengeId: Scalars['ID']['input'];
  completedAt?: InputMaybe<Scalars['DateTime']['input']>;
  userIds: Array<Scalars['ID']['input']>;
};


export type M2MMutationRootCompleteChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  completedAt?: InputMaybe<Scalars['DateTime']['input']>;
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootMarkArticleAsReadArgs = {
  achievementId: Scalars['ID']['input'];
  articleId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootMarkTrackAsListenedArgs = {
  achievementId: Scalars['ID']['input'];
  trackId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootRevokeAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootRevokeSuperTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  superTeamId: Scalars['ID']['input'];
};


export type M2MMutationRootRevokeTeamAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  teamId: Scalars['ID']['input'];
};


export type M2MMutationRootUncompleteChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootUnmarkArticleAsReadArgs = {
  achievementId: Scalars['ID']['input'];
  articleId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootUnmarkTrackAsListenedArgs = {
  achievementId: Scalars['ID']['input'];
  trackId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type M2MMutationRootUpdateStreakArgs = {
  achievementId: Scalars['ID']['input'];
  currentStreak: Scalars['Int']['input'];
  userId: Scalars['ID']['input'];
};

export type M2MQueryRoot = {
  __typename?: 'M2MQueryRoot';
  achievement: Achievement;
  achievements: Array<Achievement>;
  challenge: Challenge;
  challenges: Array<Challenge>;
  currentEvent: Event;
  currentProject: Project;
  event: Event;
  project: Project;
  superteam: SuperTeam;
  team: Team;
  user: User;
  users: Array<User>;
};


export type M2MQueryRootAchievementArgs = {
  id: Scalars['ID']['input'];
};


export type M2MQueryRootAchievementsArgs = {
  ids: Array<Scalars['ID']['input']>;
};


export type M2MQueryRootChallengeArgs = {
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


export type M2MQueryRootUsersArgs = {
  ids: Array<Scalars['ID']['input']>;
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
  streaks: Array<Streak>;
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

export type Streak = {
  __typename?: 'Streak';
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  listenedDays: Array<Scalars['Date']['output']>;
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
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  neededStreak: Scalars['Int']['output'];
  points: Scalars['Int']['output'];
  project: Project;
  streak: Streak;
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

export type TrackInput = {
  description: Scalars['String']['input'];
  image: Scalars['String']['input'];
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
  age: Scalars['Int']['output'];
  church: Church;
  email: Scalars['String']['output'];
  events: Array<Event>;
  gender: Gender;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  membersId: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  projects: Array<Project>;
  superTeams: Array<SuperTeam>;
  teams: Array<Team>;
};

export type UserFilter = {
  churchId?: InputMaybe<Scalars['ID']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  gender?: InputMaybe<Gender>;
  maxAge?: InputMaybe<Scalars['Int']['input']>;
  minAge?: InputMaybe<Scalars['Int']['input']>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  teamId?: InputMaybe<Scalars['ID']['input']>;
};

export type UserMutationRoot = {
  __typename?: 'UserMutationRoot';
  joinEvent: Event;
  joinProject: Project;
  joinTeam: Team;
  updateAvatar: User;
};


export type UserMutationRootJoinEventArgs = {
  eventId: Scalars['ID']['input'];
};


export type UserMutationRootJoinProjectArgs = {
  projectId: Scalars['ID']['input'];
};


export type UserMutationRootJoinTeamArgs = {
  code: Scalars['ID']['input'];
};


export type UserMutationRootUpdateAvatarArgs = {
  file: Scalars['Upload']['input'];
};

export type UserQueryRoot = {
  __typename?: 'UserQueryRoot';
  currentEvent: Event;
  currentProject: Project;
  events: Array<Event>;
  me: User;
  projects: Array<Project>;
};


export type UserQueryRootEventsArgs = {
  project: Scalars['ID']['input'];
};

export type CurrentProjectQueryVariables = Exact<{ [key: string]: never; }>;


export type CurrentProjectQuery = { __typename?: 'CombinedQuery', user: { __typename?: 'UserQueryRoot', currentProject: { __typename?: 'Project', branding: { __typename?: 'Branding', logo: string, rounding: number, colors: { __typename?: 'Colors', primary: string, secondary: string, tertiary: string } } } } };

export type AdminProjectPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectPageQuery = { __typename?: 'CombinedQuery', admin: { __typename?: 'AdminQueryRoot', project: { __typename?: 'Project', id: string, name: string, description: string, startDate: any, endDate: any, branding: { __typename?: 'Branding', logo: string, rounding: number, colors: { __typename?: 'Colors', primary: string, secondary: string, tertiary: string } }, achievements: Array<
        | { __typename?: 'ListeningAchievement', id: string, name: string, description: string, image: string }
        | { __typename?: 'ReadingAchievement', id: string, name: string, description: string, image: string }
        | { __typename?: 'SimpleAchievement', id: string, name: string, description: string, image: string }
        | { __typename?: 'StreakAchievement', id: string, name: string, description: string, image: string }
      >, challenges: Array<{ __typename?: 'Challenge', id: string, name: string, description: any, image: string, url: string, buttonText: string, publishedAt: any, endTime?: any | null }>, events: Array<{ __typename?: 'Event', id: string, name: string, description: string, startDate: any, endDate: any }>, streaks: Array<{ __typename?: 'Streak', id: string, name: string, description: string, relevantDays: Array<{ __typename?: 'DateRange', start: any, end: any }> }> } } };

export type AdminProjectsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminProjectsPageQuery = { __typename?: 'CombinedQuery', admin: { __typename?: 'AdminQueryRoot', projects: Array<{ __typename?: 'Project', id: string, name: string, description: string, endDate: any, startDate: any, branding: { __typename?: 'Branding', logo: string, colors: { __typename?: 'Colors', primary: string } } }> } };

export type ChallengesPageQueryVariables = Exact<{ [key: string]: never; }>;


export type ChallengesPageQuery = { __typename?: 'CombinedQuery', user: { __typename?: 'UserQueryRoot', currentProject: { __typename?: 'Project', challenges: Array<{ __typename?: 'Challenge', id: string, name: string, description: any, userCompletedAt?: any | null, image: string, url: string, buttonText: string, publishedAt: any, endTime?: any | null }> } } };

export type StandingsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type StandingsPageQuery = { __typename?: 'CombinedQuery', user: { __typename?: 'UserQueryRoot', currentProject: { __typename?: 'Project', id: string, leaderboard: Array<{ __typename?: 'LeaderboardEntry', name: string, description?: string | null, score: number, image?: string | null }> } } };

export type ProfilePageQueryVariables = Exact<{ [key: string]: never; }>;


export type ProfilePageQuery = { __typename?: 'CombinedQuery', user: { __typename?: 'UserQueryRoot', me: { __typename?: 'User', id: string, name: string, image?: string | null, church: { __typename?: 'Church', id: string, name: string }, projects: Array<{ __typename?: 'Project', id: string, achievements: Array<
          | { __typename?: 'ListeningAchievement', id: string, name: string, image: string, hidden: boolean, achievedAt?: any | null, points: number }
          | { __typename?: 'ReadingAchievement', id: string, name: string, image: string, hidden: boolean, achievedAt?: any | null, points: number }
          | { __typename?: 'SimpleAchievement', id: string, name: string, image: string, hidden: boolean, achievedAt?: any | null, points: number }
          | { __typename?: 'StreakAchievement', id: string, name: string, image: string, hidden: boolean, achievedAt?: any | null, points: number }
        > }> } } };

export type UnitPageQueryVariables = Exact<{ [key: string]: never; }>;


export type UnitPageQuery = { __typename?: 'CombinedQuery', user: { __typename?: 'UserQueryRoot', currentProject: { __typename?: 'Project', id: string, myTeam: { __typename?: 'Team', id: string, name: string, superTeam?: { __typename?: 'SuperTeam', id: string, name: string } | null, leaderboard: Array<{ __typename?: 'LeaderboardEntry', name: string, description?: string | null, score: number, image?: string | null }> } } } };


export const CurrentProjectDocument = gql`
    query CurrentProject {
  user {
    currentProject {
      branding {
        logo
        colors {
          primary
          secondary
          tertiary
        }
        rounding
      }
    }
  }
}
    `;

export function useCurrentProjectQuery(options?: Omit<Urql.UseQueryArgs<never, CurrentProjectQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<CurrentProjectQuery, CurrentProjectQueryVariables | undefined>({ query: CurrentProjectDocument, variables: undefined, ...options });
};
export const AdminProjectPageDocument = gql`
    query AdminProjectPage($projectId: ID!) {
  admin {
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
          secondary
          tertiary
        }
      }
      achievements {
        id
        name
        description
        image
      }
      challenges {
        id
        name
        description
        image
        url
        buttonText
        publishedAt
        endTime
      }
      events {
        id
        name
        description
        startDate
        endDate
      }
      streaks {
        id
        name
        description
        relevantDays {
          start
          end
        }
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
  admin {
    projects {
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
    `;

export function useAdminProjectsPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectsPageQuery, AdminProjectsPageQueryVariables | undefined>({ query: AdminProjectsPageDocument, variables: undefined, ...options });
};
export const ChallengesPageDocument = gql`
    query ChallengesPage {
  user {
    currentProject {
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
}
    `;

export function useChallengesPageQuery(options?: Omit<Urql.UseQueryArgs<never, ChallengesPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ChallengesPageQuery, ChallengesPageQueryVariables | undefined>({ query: ChallengesPageDocument, variables: undefined, ...options });
};
export const StandingsPageDocument = gql`
    query StandingsPage {
  user {
    currentProject {
      id
      leaderboard(type: TOTAL) {
        name
        description
        score
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
  user {
    me {
      id
      name
      image
      church {
        id
        name
      }
      projects {
        id
        achievements {
          id
          name
          image
          hidden
          achievedAt
          points
        }
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
  user {
    currentProject {
      id
      myTeam {
        id
        name
        superTeam {
          id
          name
        }
        leaderboard(type: TOTAL) {
          name
          description
          score
          image
        }
      }
    }
  }
}
    `;

export function useUnitPageQuery(options?: Omit<Urql.UseQueryArgs<never, UnitPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<UnitPageQuery, UnitPageQueryVariables | undefined>({ query: UnitPageDocument, variables: undefined, ...options });
};