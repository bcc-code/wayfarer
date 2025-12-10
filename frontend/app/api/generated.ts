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
  JSON: { input: any; output: any; }
  Markdown: { input: any; output: any; }
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
  buttonText: Scalars['String']['output'];
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  project: Project;
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  requiresSuperTeamMembership: Scalars['Boolean']['output'];
  requiresTeamMembership: Scalars['Boolean']['output'];
  startedAt?: Maybe<Scalars['DateTime']['output']>;
  userCompletedAt?: Maybe<Scalars['DateTime']['output']>;
  userEnrolledAt?: Maybe<Scalars['DateTime']['output']>;
  visibleAt?: Maybe<Scalars['DateTime']['output']>;
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
  challengeType?: InputMaybe<ChallengeType>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  publishedAfter?: InputMaybe<Scalars['DateTime']['input']>;
  publishedBefore?: InputMaybe<Scalars['DateTime']['input']>;
};

export enum ChallengeType {
  External = 'EXTERNAL',
  Quiz = 'QUIZ',
  Simple = 'SIMPLE'
}

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

export type ChurchInProjectInput = {
  churchId: Scalars['ID']['input'];
  projectId: Scalars['ID']['input'];
};

export type ColorSet = {
  __typename?: 'ColorSet';
  accent: Scalars['String']['output'];
  accentContrast: Scalars['String']['output'];
  backgroundDefault: Scalars['String']['output'];
  backgroundIndent: Scalars['String']['output'];
  backgroundRaised: Scalars['String']['output'];
  borderDefault: Scalars['String']['output'];
  onAccent: Scalars['String']['output'];
  shadowBlank: Scalars['String']['output'];
  shadowDefault: Scalars['String']['output'];
  textDefault: Scalars['String']['output'];
  textHint: Scalars['String']['output'];
  textMuted: Scalars['String']['output'];
};

export type ColorSetInput = {
  accent: Scalars['String']['input'];
  accentContrast: Scalars['String']['input'];
  backgroundDefault: Scalars['String']['input'];
  backgroundIndent: Scalars['String']['input'];
  backgroundRaised: Scalars['String']['input'];
  borderDefault: Scalars['String']['input'];
  onAccent: Scalars['String']['input'];
  shadowBlank: Scalars['String']['input'];
  shadowDefault: Scalars['String']['input'];
  textDefault: Scalars['String']['input'];
  textHint: Scalars['String']['input'];
  textMuted: Scalars['String']['input'];
};

export type Colors = {
  __typename?: 'Colors';
  dark: ColorSet;
  light: ColorSet;
};

export type ColorsInput = {
  dark: ColorSetInput;
  light: ColorSetInput;
};

export type Consent = {
  __typename?: 'Consent';
  body: MarkdownText;
  id: Scalars['ID']['output'];
  key: Scalars['String']['output'];
  managedBy?: Maybe<Scalars['String']['output']>;
  managementType: ConsentManagementType;
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  title: Scalars['String']['output'];
  url?: Maybe<Scalars['String']['output']>;
  userHistory: Array<UserConsentHistoryEntry>;
  version: Scalars['Int']['output'];
};

export enum ConsentAction {
  Accepted = 'ACCEPTED',
  Rejected = 'REJECTED'
}

export enum ConsentManagementType {
  Local = 'LOCAL',
  Remote = 'REMOTE'
}

export type ConsentStatus = {
  __typename?: 'ConsentStatus';
  acceptedConsents: Array<UserConsent>;
  pendingConsents: Array<Consent>;
  rejectedConsents: Array<UserConsent>;
};

export type CreateChallengeInput = {
  allowSelfCompletion?: InputMaybe<Scalars['Boolean']['input']>;
  buttonText: Scalars['String']['input'];
  description?: InputMaybe<Scalars['HTML']['input']>;
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  requiresSuperTeamMembership?: InputMaybe<Scalars['Boolean']['input']>;
  requiresTeamMembership?: InputMaybe<Scalars['Boolean']['input']>;
  type: ChallengeType;
  url?: InputMaybe<Scalars['String']['input']>;
  visibleAt?: InputMaybe<Scalars['DateTime']['input']>;
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

export type CreatePredefinedAnswerInput = {
  answerOrder: Scalars['Int']['input'];
  answerText: Scalars['String']['input'];
  isCorrect: Scalars['Boolean']['input'];
};

export type CreateProjectInput = {
  branding: BrandingInput;
  description?: InputMaybe<Scalars['String']['input']>;
  endDate: Scalars['DateTime']['input'];
  name: Scalars['String']['input'];
  rules?: InputMaybe<Scalars['String']['input']>;
  startDate: Scalars['DateTime']['input'];
};

export type CreateQuizAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description: Scalars['String']['input'];
  hidden: Scalars['Boolean']['input'];
  image?: InputMaybe<Scalars['String']['input']>;
  minScorePercentage?: InputMaybe<Scalars['Int']['input']>;
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  quizId: Scalars['ID']['input'];
  requireCompletion: Scalars['Boolean']['input'];
};

export type CreateQuizInput = {
  allowRetakes: Scalars['Boolean']['input'];
  challengeId: Scalars['ID']['input'];
  completionPoints: Scalars['Int']['input'];
  description: Scalars['String']['input'];
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  randomizeQuestions: Scalars['Boolean']['input'];
  revealCorrectAnswers: Scalars['Boolean']['input'];
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
};

export type CreateQuizQuestionInput = {
  allowMultipleSelection?: InputMaybe<Scalars['Boolean']['input']>;
  maxValue?: InputMaybe<Scalars['Float']['input']>;
  minValue?: InputMaybe<Scalars['Float']['input']>;
  predefinedAnswers?: InputMaybe<Array<CreatePredefinedAnswerInput>>;
  questionOrder: Scalars['Int']['input'];
  questionText: Scalars['String']['input'];
  questionType: QuizQuestionType;
  stepValue?: InputMaybe<Scalars['Float']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
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

export type CreateScoreAdjustmentInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  userId: Scalars['ID']['input'];
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
  streakId: Scalars['ID']['input'];
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

export type EnrollmentTargetInput = {
  allProjectMembers?: InputMaybe<Scalars['ID']['input']>;
  churchInProject?: InputMaybe<ChurchInProjectInput>;
  superTeamIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  teamIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  userIds?: InputMaybe<Array<Scalars['ID']['input']>>;
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

export type ExternalChallenge = Challenge & {
  __typename?: 'ExternalChallenge';
  buttonText: Scalars['String']['output'];
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  project: Project;
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  requiresSuperTeamMembership: Scalars['Boolean']['output'];
  requiresTeamMembership: Scalars['Boolean']['output'];
  startedAt?: Maybe<Scalars['DateTime']['output']>;
  url: Scalars['String']['output'];
  userCompletedAt?: Maybe<Scalars['DateTime']['output']>;
  userEnrolledAt?: Maybe<Scalars['DateTime']['output']>;
  visibleAt?: Maybe<Scalars['DateTime']['output']>;
};

export type FreeTextQuestion = QuizQuestion & {
  __typename?: 'FreeTextQuestion';
  id: Scalars['ID']['output'];
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export type FreeTextResponse = QuizResponse & {
  __typename?: 'FreeTextResponse';
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  question: QuizQuestion;
  submission: QuizSubmission;
  textResponse: Scalars['String']['output'];
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

export enum Gender {
  Female = 'FEMALE',
  Male = 'MALE'
}

export type JsonQuestion = QuizQuestion & {
  __typename?: 'JsonQuestion';
  id: Scalars['ID']['output'];
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export type JsonResponse = QuizResponse & {
  __typename?: 'JsonResponse';
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  jsonResponse: Scalars['JSON']['output'];
  question: QuizQuestion;
  submission: QuizSubmission;
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

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
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  rank?: Maybe<Scalars['Int']['output']>;
  score: Scalars['Int']['output'];
  tags: Array<LeaderboardEntryTag>;
};

export enum LeaderboardEntryTag {
  Admin = 'ADMIN',
  Me = 'ME',
  TeamLead = 'TEAM_LEAD'
}

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

export type MarkdownText = {
  __typename?: 'MarkdownText';
  html: Scalars['String']['output'];
  markdown: Scalars['String']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  _empty?: Maybe<Scalars['Boolean']['output']>;
  acceptConsent: UserConsent;
  addQuizQuestion: QuizQuestion;
  addTeamMembers: Team;
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
  bulkCompleteChallenges: Array<Challenge>;
  bulkEnrollUsersInChallenge: Array<Challenge>;
  bulkPublishChallenges: Array<Challenge>;
  bulkUnenrollUsersFromChallenge: Scalars['Boolean']['output'];
  completeChallenge: Challenge;
  createChallenge: Challenge;
  createConsent: Consent;
  createEvent: Event;
  createListeningAchievement: ListeningAchievement;
  createProject: Project;
  createQuiz: Quiz;
  createQuizAchievement: QuizAchievement;
  createQuizSubmission: QuizSubmission;
  createReadingAchievement: ReadingAchievement;
  createScoreAdjustment: ScoreJournal;
  createSimpleAchievement: SimpleAchievement;
  createStreak: Streak;
  createStreakAchievement: StreakAchievement;
  createSuperTeam: SuperTeam;
  createTeam: Team;
  deleteAchievement: Scalars['Boolean']['output'];
  deleteChallenge: Scalars['Boolean']['output'];
  deleteEvent: Scalars['Boolean']['output'];
  deleteProject: Scalars['Boolean']['output'];
  deleteQuiz: Scalars['Boolean']['output'];
  deleteQuizQuestion: Scalars['Boolean']['output'];
  deleteStreak: Scalars['Boolean']['output'];
  deleteSuperTeam: Scalars['Boolean']['output'];
  deleteTeam: Scalars['Boolean']['output'];
  enrollInChallenge: Challenge;
  enrollUserInChallenge: Challenge;
  finalizeQuiz: QuizSubmission;
  joinEvent: Event;
  joinProject: Project;
  joinTeam: Team;
  linkAchievementToChallenge: Achievement;
  markArticleAsRead: ReadingAchievement;
  markTrackAsListened: ListeningAchievement;
  moveEvent: Event;
  publishChallenge: Challenge;
  publishQuiz: Quiz;
  recordStreakActivity: StreakAchievement;
  regenerateJoinCode: Team;
  rejectConsent: UserConsent;
  removeTeamMembers: Team;
  removeUserFromProject: User;
  reorderQuizQuestions: Array<QuizQuestion>;
  revokeAchievement: Scalars['Boolean']['output'];
  revokeRole: Scalars['Boolean']['output'];
  revokeSuperTeamAchievement: Scalars['Boolean']['output'];
  revokeTeamAchievement: Scalars['Boolean']['output'];
  selfCompleteChallenge: SimpleChallenge;
  setChallengeRequirements: Challenge;
  setChallengeVisibility: Challenge;
  startQuiz: QuizSubmission;
  submitQuizAnswer: QuizResponse;
  uncompleteChallenge: Scalars['Boolean']['output'];
  unenrollFromChallenge: Scalars['Boolean']['output'];
  unenrollUserFromChallenge: Scalars['Boolean']['output'];
  unmarkArticleAsRead: ReadingAchievement;
  unmarkTrackAsListened: ListeningAchievement;
  updateAchievement: Achievement;
  updateAvatar: User;
  updateChallenge: Challenge;
  updateConsent: Consent;
  updateEvent: Event;
  updateListeningAchievement: ListeningAchievement;
  updateProject: Project;
  updateQuiz: Quiz;
  updateQuizQuestion: QuizQuestion;
  updateReadingAchievement: ReadingAchievement;
  updateStreak: Streak;
  updateStreakAchievement: StreakAchievement;
  updateSuperTeam: SuperTeam;
  updateTeam: Team;
};


export type MutationAcceptConsentArgs = {
  consentId: Scalars['ID']['input'];
};


export type MutationAddQuizQuestionArgs = {
  input: CreateQuizQuestionInput;
  quizId: Scalars['ID']['input'];
};


export type MutationAddTeamMembersArgs = {
  force?: InputMaybe<Scalars['Boolean']['input']>;
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
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


export type MutationBulkCompleteChallengesArgs = {
  challengeId: Scalars['ID']['input'];
  completedAt?: InputMaybe<Scalars['DateTime']['input']>;
  target: EnrollmentTargetInput;
};


export type MutationBulkEnrollUsersInChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  target: EnrollmentTargetInput;
};


export type MutationBulkPublishChallengesArgs = {
  ids: Array<Scalars['ID']['input']>;
  publishedAt: Scalars['DateTime']['input'];
};


export type MutationBulkUnenrollUsersFromChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  target: EnrollmentTargetInput;
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


export type MutationCreateConsentArgs = {
  body: Scalars['String']['input'];
  isRemote?: InputMaybe<Scalars['Boolean']['input']>;
  key: Scalars['String']['input'];
  managedBy?: InputMaybe<Scalars['String']['input']>;
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  title: Scalars['String']['input'];
  url?: InputMaybe<Scalars['String']['input']>;
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


export type MutationCreateQuizArgs = {
  input: CreateQuizInput;
};


export type MutationCreateQuizAchievementArgs = {
  input: CreateQuizAchievementInput;
};


export type MutationCreateQuizSubmissionArgs = {
  completedAt?: InputMaybe<Scalars['DateTime']['input']>;
  quizId: Scalars['ID']['input'];
  responses: Array<SubmitQuizAnswerInput>;
  userId: Scalars['ID']['input'];
};


export type MutationCreateReadingAchievementArgs = {
  input: CreateReadingAchievementInput;
};


export type MutationCreateScoreAdjustmentArgs = {
  input: CreateScoreAdjustmentInput;
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


export type MutationDeleteQuizArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteQuizQuestionArgs = {
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


export type MutationEnrollInChallengeArgs = {
  challengeId: Scalars['ID']['input'];
};


export type MutationEnrollUserInChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationFinalizeQuizArgs = {
  submissionId: Scalars['ID']['input'];
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


export type MutationPublishQuizArgs = {
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


export type MutationRejectConsentArgs = {
  consentId: Scalars['ID']['input'];
};


export type MutationRemoveTeamMembersArgs = {
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
};


export type MutationRemoveUserFromProjectArgs = {
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationReorderQuizQuestionsArgs = {
  questionIds: Array<Scalars['ID']['input']>;
  quizId: Scalars['ID']['input'];
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


export type MutationSelfCompleteChallengeArgs = {
  challengeId: Scalars['ID']['input'];
};


export type MutationSetChallengeRequirementsArgs = {
  id: Scalars['ID']['input'];
  requiresSuperTeamMembership?: InputMaybe<Scalars['Boolean']['input']>;
  requiresTeamMembership?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationSetChallengeVisibilityArgs = {
  id: Scalars['ID']['input'];
  startedAt?: InputMaybe<Scalars['DateTime']['input']>;
  visibleAt: Scalars['DateTime']['input'];
};


export type MutationStartQuizArgs = {
  quizId: Scalars['ID']['input'];
};


export type MutationSubmitQuizAnswerArgs = {
  input: SubmitQuizAnswerInput;
  submissionId: Scalars['ID']['input'];
};


export type MutationUncompleteChallengeArgs = {
  challengeId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationUnenrollFromChallengeArgs = {
  challengeId: Scalars['ID']['input'];
};


export type MutationUnenrollUserFromChallengeArgs = {
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


export type MutationUpdateConsentArgs = {
  body?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
  url?: InputMaybe<Scalars['String']['input']>;
};


export type MutationUpdateEventArgs = {
  id: Scalars['ID']['input'];
  input: UpdateEventInput;
};


export type MutationUpdateListeningAchievementArgs = {
  id: Scalars['ID']['input'];
  input: UpdateListeningAchievementInput;
};


export type MutationUpdateProjectArgs = {
  id: Scalars['ID']['input'];
  input: UpdateProjectInput;
};


export type MutationUpdateQuizArgs = {
  id: Scalars['ID']['input'];
  input: UpdateQuizInput;
};


export type MutationUpdateQuizQuestionArgs = {
  id: Scalars['ID']['input'];
  input: UpdateQuizQuestionInput;
};


export type MutationUpdateReadingAchievementArgs = {
  id: Scalars['ID']['input'];
  input: UpdateReadingAchievementInput;
};


export type MutationUpdateStreakArgs = {
  id: Scalars['ID']['input'];
  input: UpdateStreakInput;
};


export type MutationUpdateStreakAchievementArgs = {
  id: Scalars['ID']['input'];
  input: UpdateStreakAchievementInput;
};


export type MutationUpdateSuperTeamArgs = {
  id: Scalars['ID']['input'];
  input: UpdateSuperTeamInput;
};


export type MutationUpdateTeamArgs = {
  id: Scalars['ID']['input'];
  input: UpdateTeamInput;
};

export type NumberQuestion = QuizQuestion & {
  __typename?: 'NumberQuestion';
  id: Scalars['ID']['output'];
  maxValue?: Maybe<Scalars['Float']['output']>;
  minValue?: Maybe<Scalars['Float']['output']>;
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  stepValue?: Maybe<Scalars['Float']['output']>;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export type NumberResponse = QuizResponse & {
  __typename?: 'NumberResponse';
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  numberResponse: Scalars['Float']['output'];
  question: QuizQuestion;
  submission: QuizSubmission;
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor?: Maybe<Scalars['String']['output']>;
};

export type PredefinedQuestion = QuizQuestion & {
  __typename?: 'PredefinedQuestion';
  allowMultipleSelection: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  predefinedAnswers: Array<QuizPredefinedAnswer>;
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export type PredefinedResponse = QuizResponse & {
  __typename?: 'PredefinedResponse';
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  isCorrect?: Maybe<Scalars['Boolean']['output']>;
  question: QuizQuestion;
  selectedAnswerIds: Array<Scalars['ID']['output']>;
  selectedAnswers: Array<QuizPredefinedAnswer>;
  submission: QuizSubmission;
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
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
  journal: ScoreJournalConnection;
  leaderboard: LeaderboardConnection;
  myTeam?: Maybe<Team>;
  name: Scalars['String']['output'];
  rules?: Maybe<MarkdownText>;
  startDate: Scalars['DateTime']['output'];
  streaks: Array<Streak>;
  teams: Array<Team>;
};


export type ProjectJournalArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<ScoreJournalFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
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
  consent: Consent;
  consents: Array<Consent>;
  currentEvent: Event;
  currentProject: Project;
  event: Event;
  events: EventConnection;
  me: User;
  myCurrentEvent: Event;
  myCurrentProject: Project;
  myEvents: Array<Event>;
  myProjects: Array<Project>;
  pendingConsents: Array<Consent>;
  project: Project;
  projects: ProjectConnection;
  quiz: Quiz;
  quizSubmission: QuizSubmission;
  quizSubmissions: QuizSubmissionConnection;
  quizzes: QuizConnection;
  scoreJournal: ScoreJournalConnection;
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


export type QueryConsentArgs = {
  id: Scalars['ID']['input'];
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


export type QueryQuizArgs = {
  id: Scalars['ID']['input'];
};


export type QueryQuizSubmissionArgs = {
  id: Scalars['ID']['input'];
};


export type QueryQuizSubmissionsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  quizId: Scalars['ID']['input'];
  userId?: InputMaybe<Scalars['ID']['input']>;
};


export type QueryQuizzesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<QuizFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryScoreJournalArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<ScoreJournalFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
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

export type Quiz = {
  __typename?: 'Quiz';
  allowRetakes: Scalars['Boolean']['output'];
  challenge: Challenge;
  completionPoints: Scalars['Int']['output'];
  description: Scalars['String']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  project: Project;
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  questions: Array<QuizQuestion>;
  randomizeQuestions: Scalars['Boolean']['output'];
  revealCorrectAnswers: Scalars['Boolean']['output'];
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
  userActiveSubmission?: Maybe<QuizSubmission>;
  userCanStart: Scalars['Boolean']['output'];
  userSubmissions: Array<QuizSubmission>;
};

export type QuizAchievement = Achievement & {
  __typename?: 'QuizAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  description: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  minScorePercentage?: Maybe<Scalars['Int']['output']>;
  name: Scalars['String']['output'];
  points: Scalars['Int']['output'];
  project: Project;
  quiz: Quiz;
  requireCompletion: Scalars['Boolean']['output'];
};

export type QuizChallenge = Challenge & {
  __typename?: 'QuizChallenge';
  buttonText: Scalars['String']['output'];
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  project: Project;
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  quiz: Quiz;
  requiresSuperTeamMembership: Scalars['Boolean']['output'];
  requiresTeamMembership: Scalars['Boolean']['output'];
  startedAt?: Maybe<Scalars['DateTime']['output']>;
  userCompletedAt?: Maybe<Scalars['DateTime']['output']>;
  userEnrolledAt?: Maybe<Scalars['DateTime']['output']>;
  visibleAt?: Maybe<Scalars['DateTime']['output']>;
};

export type QuizConnection = {
  __typename?: 'QuizConnection';
  edges: Array<QuizEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type QuizEdge = {
  __typename?: 'QuizEdge';
  cursor: Scalars['String']['output'];
  node: Quiz;
};

export type QuizFilter = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  publishedAfter?: InputMaybe<Scalars['DateTime']['input']>;
  publishedBefore?: InputMaybe<Scalars['DateTime']['input']>;
};

export type QuizPredefinedAnswer = {
  __typename?: 'QuizPredefinedAnswer';
  answerOrder: Scalars['Int']['output'];
  answerText: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  isCorrect?: Maybe<Scalars['Boolean']['output']>;
  question: QuizQuestion;
};

export type QuizQuestion = {
  id: Scalars['ID']['output'];
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export enum QuizQuestionType {
  FreeText = 'FREE_TEXT',
  Json = 'JSON',
  Number = 'NUMBER',
  Predefined = 'PREDEFINED'
}

export type QuizResponse = {
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  question: QuizQuestion;
  submission: QuizSubmission;
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

export type QuizSubmission = {
  __typename?: 'QuizSubmission';
  completedAt?: Maybe<Scalars['DateTime']['output']>;
  expiresAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  isExpired: Scalars['Boolean']['output'];
  maxScore?: Maybe<Scalars['Int']['output']>;
  orderedQuestions: Array<QuizQuestion>;
  pointsAwarded?: Maybe<Scalars['Int']['output']>;
  questionOrder: Array<Scalars['ID']['output']>;
  quiz: Quiz;
  responses: Array<QuizResponse>;
  score?: Maybe<Scalars['Int']['output']>;
  scorePercentage?: Maybe<Scalars['Float']['output']>;
  startedAt: Scalars['DateTime']['output'];
  user: User;
};

export type QuizSubmissionConnection = {
  __typename?: 'QuizSubmissionConnection';
  edges: Array<QuizSubmissionEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type QuizSubmissionEdge = {
  __typename?: 'QuizSubmissionEdge';
  cursor: Scalars['String']['output'];
  node: QuizSubmission;
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

export type ScoreJournal = {
  __typename?: 'ScoreJournal';
  awardedBy?: Maybe<User>;
  challenge?: Maybe<Challenge>;
  createdAt: Scalars['DateTime']['output'];
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  points: Scalars['Int']['output'];
  project: Project;
  reason?: Maybe<Scalars['String']['output']>;
  source?: Maybe<ScoreSource>;
  sourceType: ScoreSourceType;
  user: User;
};

export type ScoreJournalConnection = {
  __typename?: 'ScoreJournalConnection';
  edges: Array<ScoreJournalEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ScoreJournalEdge = {
  __typename?: 'ScoreJournalEdge';
  cursor: Scalars['String']['output'];
  node: ScoreJournal;
};

export type ScoreJournalFilter = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  sourceType?: InputMaybe<ScoreSourceType>;
};

export type ScoreSource = Event | ExternalChallenge | ListeningAchievement | QuizChallenge | ReadingAchievement | SimpleAchievement | SimpleChallenge | StreakAchievement;

export enum ScoreSourceType {
  Achievement = 'ACHIEVEMENT',
  Challenge = 'CHALLENGE',
  Event = 'EVENT',
  Manual = 'MANUAL'
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

export type SimpleChallenge = Challenge & {
  __typename?: 'SimpleChallenge';
  allowSelfCompletion: Scalars['Boolean']['output'];
  buttonText: Scalars['String']['output'];
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  image?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  project: Project;
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  requiresSuperTeamMembership: Scalars['Boolean']['output'];
  requiresTeamMembership: Scalars['Boolean']['output'];
  startedAt?: Maybe<Scalars['DateTime']['output']>;
  userCompletedAt?: Maybe<Scalars['DateTime']['output']>;
  userEnrolledAt?: Maybe<Scalars['DateTime']['output']>;
  visibleAt?: Maybe<Scalars['DateTime']['output']>;
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

export type SubmitQuizAnswerInput = {
  jsonResponse?: InputMaybe<Scalars['JSON']['input']>;
  numberResponse?: InputMaybe<Scalars['Float']['input']>;
  questionId: Scalars['ID']['input'];
  selectedAnswerIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  textResponse?: InputMaybe<Scalars['String']['input']>;
  timeSpentSeconds?: InputMaybe<Scalars['Int']['input']>;
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
  memberLeaderboard: Array<LeaderboardEntry>;
  members: Array<TeamMember>;
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

export type TeamMember = {
  __typename?: 'TeamMember';
  church: Church;
  id: Scalars['ID']['output'];
  isTeamLead: Scalars['Boolean']['output'];
  joinedAt: Scalars['String']['output'];
  name: Scalars['String']['output'];
  user: User;
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
  allowSelfCompletion?: InputMaybe<Scalars['Boolean']['input']>;
  buttonText?: InputMaybe<Scalars['String']['input']>;
  description?: InputMaybe<Scalars['HTML']['input']>;
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  requiresSuperTeamMembership?: InputMaybe<Scalars['Boolean']['input']>;
  requiresTeamMembership?: InputMaybe<Scalars['Boolean']['input']>;
  startedAt?: InputMaybe<Scalars['DateTime']['input']>;
  url?: InputMaybe<Scalars['String']['input']>;
  visibleAt?: InputMaybe<Scalars['DateTime']['input']>;
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

export type UpdateListeningAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden?: InputMaybe<Scalars['Boolean']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  points?: InputMaybe<Scalars['Int']['input']>;
  tracks?: InputMaybe<Array<TrackInput>>;
};

export type UpdateProjectInput = {
  branding?: InputMaybe<BrandingInput>;
  description?: InputMaybe<Scalars['String']['input']>;
  endDate?: InputMaybe<Scalars['DateTime']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  rules?: InputMaybe<Scalars['String']['input']>;
  startDate?: InputMaybe<Scalars['DateTime']['input']>;
};

export type UpdateQuizInput = {
  allowRetakes?: InputMaybe<Scalars['Boolean']['input']>;
  completionPoints?: InputMaybe<Scalars['Int']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  randomizeQuestions?: InputMaybe<Scalars['Boolean']['input']>;
  revealCorrectAnswers?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
};

export type UpdateQuizQuestionInput = {
  allowMultipleSelection?: InputMaybe<Scalars['Boolean']['input']>;
  maxValue?: InputMaybe<Scalars['Float']['input']>;
  minValue?: InputMaybe<Scalars['Float']['input']>;
  predefinedAnswers?: InputMaybe<Array<CreatePredefinedAnswerInput>>;
  questionOrder?: InputMaybe<Scalars['Int']['input']>;
  questionText?: InputMaybe<Scalars['String']['input']>;
  stepValue?: InputMaybe<Scalars['Float']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
};

export type UpdateReadingAchievementInput = {
  articles?: InputMaybe<Array<ArticleInput>>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden?: InputMaybe<Scalars['Boolean']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  points?: InputMaybe<Scalars['Int']['input']>;
};

export type UpdateStreakAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden?: InputMaybe<Scalars['Boolean']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  neededStreak?: InputMaybe<Scalars['Int']['input']>;
  points?: InputMaybe<Scalars['Int']['input']>;
  streakId?: InputMaybe<Scalars['ID']['input']>;
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
  consentStatus: ConsentStatus;
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

export type UserConsent = {
  __typename?: 'UserConsent';
  action: ConsentAction;
  actionDate: Scalars['DateTime']['output'];
  consent: Consent;
  id: Scalars['ID']['output'];
};

export type UserConsentHistoryEntry = {
  __typename?: 'UserConsentHistoryEntry';
  action: ConsentAction;
  consent: Consent;
  externalConsentId?: Maybe<Scalars['String']['output']>;
  externalTimestamp?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  occurredAt: Scalars['DateTime']['output'];
  source?: Maybe<Scalars['String']['output']>;
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

export type ProjectRulesQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectRulesQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', rules?: { __typename?: 'MarkdownText', markdown: string, html: string } | null } };

export type PointHistoryQueryVariables = Exact<{
  last?: InputMaybe<Scalars['Int']['input']>;
}>;


export type PointHistoryQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', journal: { __typename?: 'ScoreJournalConnection', edges: Array<{ __typename?: 'ScoreJournalEdge', node: { __typename?: 'ScoreJournal', id: string, sourceType: ScoreSourceType, reason?: string | null, points: number, createdAt: any, source?:
            | { __typename: 'Event', id: string, name: string }
            | { __typename: 'ExternalChallenge', id: string, name: string }
            | { __typename: 'ListeningAchievement', id: string, name: string }
            | { __typename: 'QuizChallenge', id: string, name: string }
            | { __typename: 'ReadingAchievement', id: string, name: string }
            | { __typename: 'SimpleAchievement', id: string, name: string }
            | { __typename: 'SimpleChallenge', id: string, name: string }
            | { __typename: 'StreakAchievement', id: string, name: string }
           | null } }> } } };

export type GetMeQueryVariables = Exact<{ [key: string]: never; }>;


export type GetMeQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string, email: string, image?: string | null, membersId: string, gender: Gender, birthdate: string, age?: number | null, church: { __typename?: 'Church', id: string, name: string, country: string, category: ChurchCategory }, roles: Array<{ __typename?: 'UserRole', id: string, role: RoleType, scope?: { __typename?: 'RoleScope', id: string, type: ScopeType, church?: { __typename?: 'Church', id: string } | null, team?: { __typename?: 'Team', id: string } | null, project?: { __typename?: 'Project', id: string } | null } | null }> } };

export type DeleteAchievementMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteAchievementMutation = { __typename?: 'Mutation', deleteAchievement: boolean };

export type UpdateAchievementMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateAchievementInput;
}>;


export type UpdateAchievementMutation = { __typename?: 'Mutation', updateAchievement:
    | { __typename?: 'ListeningAchievement', id: string }
    | { __typename?: 'QuizAchievement', id: string }
    | { __typename?: 'ReadingAchievement', id: string }
    | { __typename?: 'SimpleAchievement', id: string }
    | { __typename?: 'StreakAchievement', id: string }
   };

export type CreateSimpleAchievementMutationVariables = Exact<{
  input: CreateSimpleAchievementInput;
}>;


export type CreateSimpleAchievementMutation = { __typename?: 'Mutation', createSimpleAchievement: { __typename?: 'SimpleAchievement', id: string } };

export type CreateStreakAchievementMutationVariables = Exact<{
  input: CreateStreakAchievementInput;
}>;


export type CreateStreakAchievementMutation = { __typename?: 'Mutation', createStreakAchievement: { __typename?: 'StreakAchievement', id: string } };

export type CreateReadingAchievementMutationVariables = Exact<{
  input: CreateReadingAchievementInput;
}>;


export type CreateReadingAchievementMutation = { __typename?: 'Mutation', createReadingAchievement: { __typename?: 'ReadingAchievement', id: string } };

export type CreateListeningAchievementMutationVariables = Exact<{
  input: CreateListeningAchievementInput;
}>;


export type CreateListeningAchievementMutation = { __typename?: 'Mutation', createListeningAchievement: { __typename?: 'ListeningAchievement', id: string } };

export type DeleteChallengeMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteChallengeMutation = { __typename?: 'Mutation', deleteChallenge: boolean };

export type UpdateChallengeMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateChallengeInput;
}>;


export type UpdateChallengeMutation = { __typename?: 'Mutation', updateChallenge:
    | { __typename?: 'ExternalChallenge', id: string }
    | { __typename?: 'QuizChallenge', id: string }
    | { __typename?: 'SimpleChallenge', id: string }
   };

export type CreateChallengeMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  eventId: Scalars['ID']['input'];
  input: CreateChallengeInput;
}>;


export type CreateChallengeMutation = { __typename?: 'Mutation', createChallenge:
    | { __typename?: 'ExternalChallenge', id: string }
    | { __typename?: 'QuizChallenge', id: string }
    | { __typename?: 'SimpleChallenge', id: string }
   };

export type AcceptConsentMutationVariables = Exact<{
  consentId: Scalars['ID']['input'];
}>;


export type AcceptConsentMutation = { __typename?: 'Mutation', acceptConsent: { __typename?: 'UserConsent', id: string, action: ConsentAction, actionDate: any } };

export type RejectConsentMutationVariables = Exact<{
  consentId: Scalars['ID']['input'];
}>;


export type RejectConsentMutation = { __typename?: 'Mutation', rejectConsent: { __typename?: 'UserConsent', id: string, action: ConsentAction, actionDate: any } };

export type DeleteEventMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteEventMutation = { __typename?: 'Mutation', deleteEvent: boolean };

export type UpdateEventMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateEventInput;
}>;


export type UpdateEventMutation = { __typename?: 'Mutation', updateEvent: { __typename?: 'Event', id: string } };

export type CreateEventMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  input: CreateEventInput;
}>;


export type CreateEventMutation = { __typename?: 'Mutation', createEvent: { __typename?: 'Event', id: string } };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'Project', id: string } };

export type UpdateProjectMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateProjectInput;
}>;


export type UpdateProjectMutation = { __typename?: 'Mutation', updateProject: { __typename?: 'Project', id: string } };

export type DeleteProjectMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteProjectMutation = { __typename?: 'Mutation', deleteProject: boolean };

export type AssignRoleMutationVariables = Exact<{
  input: AssignRoleInput;
}>;


export type AssignRoleMutation = { __typename?: 'Mutation', assignRole: { __typename?: 'UserRole', id: string, role: RoleType, scope?: { __typename?: 'RoleScope', id: string, type: ScopeType } | null } };

export type RevokeRoleMutationVariables = Exact<{
  input: RevokeRoleInput;
}>;


export type RevokeRoleMutation = { __typename?: 'Mutation', revokeRole: boolean };

export type DeleteStreakMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteStreakMutation = { __typename?: 'Mutation', deleteStreak: boolean };

export type UpdateStreakMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateStreakInput;
}>;


export type UpdateStreakMutation = { __typename?: 'Mutation', updateStreak: { __typename?: 'Streak', id: string } };

export type CreateStreakMutationVariables = Exact<{
  input: CreateStreakInput;
}>;


export type CreateStreakMutation = { __typename?: 'Mutation', createStreak: { __typename?: 'Streak', id: string } };

export type CreateTeamMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  input: CreateTeamInput;
}>;


export type CreateTeamMutation = { __typename?: 'Mutation', createTeam: { __typename?: 'Team', id: string } };

export type UpdateTeamMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateTeamInput;
}>;


export type UpdateTeamMutation = { __typename?: 'Mutation', updateTeam: { __typename?: 'Team', id: string } };

export type DeleteTeamMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteTeamMutation = { __typename?: 'Mutation', deleteTeam: boolean };

export type AddTeamMembersMutationVariables = Exact<{
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']> | Scalars['ID']['input'];
  force?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type AddTeamMembersMutation = { __typename?: 'Mutation', addTeamMembers: { __typename?: 'Team', id: string } };

export type RemoveTeamMembersMutationVariables = Exact<{
  teamId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']> | Scalars['ID']['input'];
}>;


export type RemoveTeamMembersMutation = { __typename?: 'Mutation', removeTeamMembers: { __typename?: 'Team', id: string } };

export type RegenerateJoinCodeMutationVariables = Exact<{
  teamId: Scalars['ID']['input'];
}>;


export type RegenerateJoinCodeMutation = { __typename?: 'Mutation', regenerateJoinCode: { __typename?: 'Team', id: string, joinCode: string } };

export type AssignTeamLeadMutationVariables = Exact<{
  teamId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
}>;


export type AssignTeamLeadMutation = { __typename?: 'Mutation', assignTeamLead: { __typename?: 'Team', id: string } };

export type ChallengePageQueryVariables = Exact<{
  challengeId: Scalars['ID']['input'];
}>;


export type ChallengePageQuery = { __typename?: 'Query', challenge:
    | { __typename: 'ExternalChallenge', url: string, id: string, name: string, description: any, userEnrolledAt?: any | null, userCompletedAt?: any | null }
    | { __typename: 'QuizChallenge', id: string, name: string, description: any, userEnrolledAt?: any | null, userCompletedAt?: any | null, quiz: { __typename?: 'Quiz', id: string, name: string, description: string, timeoutSeconds?: number | null, randomizeQuestions: boolean, revealCorrectAnswers: boolean, allowRetakes: boolean, completionPoints: number, publishedAt?: any | null, endTime?: any | null, userCanStart: boolean, userSubmissions: Array<{ __typename?: 'QuizSubmission', id: string, startedAt: any, completedAt?: any | null, expiresAt?: any | null, isExpired: boolean, score?: number | null, maxScore?: number | null, scorePercentage?: number | null, orderedQuestions: Array<
            | { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
            | { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
            | { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
            | { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> }
          >, responses: Array<
            | { __typename: 'FreeTextResponse', textResponse: string, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null }
            | { __typename: 'JsonResponse', jsonResponse: any, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null }
            | { __typename: 'NumberResponse', numberResponse: number, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null }
            | { __typename: 'PredefinedResponse', isCorrect?: boolean | null, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, selectedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> }
          > }> } }
    | { __typename: 'SimpleChallenge', allowSelfCompletion: boolean, id: string, name: string, description: any, userEnrolledAt?: any | null, userCompletedAt?: any | null }
   };

export type ChallengesPageQueryVariables = Exact<{ [key: string]: never; }>;


export type ChallengesPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', challenges: Array<
      | { __typename?: 'ExternalChallenge', id: string, name: string, description: any, image?: string | null, buttonText: string, publishedAt?: any | null, endTime?: any | null, visibleAt?: any | null }
      | { __typename?: 'QuizChallenge', id: string, name: string, description: any, image?: string | null, buttonText: string, publishedAt?: any | null, endTime?: any | null, visibleAt?: any | null }
      | { __typename?: 'SimpleChallenge', id: string, name: string, description: any, image?: string | null, buttonText: string, publishedAt?: any | null, endTime?: any | null, visibleAt?: any | null }
    > } };

export type ProfilePageQueryVariables = Exact<{ [key: string]: never; }>;


export type ProfilePageQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string, consentStatus: { __typename?: 'ConsentStatus', pendingConsents: Array<{ __typename: 'Consent', id: string, key: string, version: number, title: string, managementType: ConsentManagementType, managedBy?: string | null, body: { __typename?: 'MarkdownText', html: string } }> } }, myCurrentProject: { __typename?: 'Project', id: string, name: string, achievements: Array<
      | { __typename?: 'ListeningAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
      | { __typename?: 'QuizAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
      | { __typename?: 'ReadingAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
      | { __typename?: 'SimpleAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
      | { __typename?: 'StreakAchievement', id: string, name: string, description: string, image?: string | null, hidden: boolean, achievedAt?: any | null, points: number }
    >, leaderboard: { __typename?: 'LeaderboardConnection', me?: { __typename?: 'LeaderboardEntry', score: number, rank?: number | null } | null } } };

export type ConsentsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type ConsentsPageQuery = { __typename?: 'Query', me: { __typename?: 'User', consentStatus: { __typename?: 'ConsentStatus', pendingConsents: Array<{ __typename: 'Consent', id: string, key: string, version: number, title: string, publishedAt?: any | null, managedBy?: string | null, managementType: ConsentManagementType, url?: string | null, body: { __typename?: 'MarkdownText', html: string } }>, acceptedConsents: Array<{ __typename: 'UserConsent', id: string, action: ConsentAction, actionDate: any, consent: { __typename?: 'Consent', title: string, managedBy?: string | null, managementType: ConsentManagementType, url?: string | null, body: { __typename?: 'MarkdownText', html: string } } }>, rejectedConsents: Array<{ __typename: 'UserConsent', id: string, action: ConsentAction, actionDate: any, consent: { __typename?: 'Consent', title: string, managedBy?: string | null, managementType: ConsentManagementType, url?: string | null, body: { __typename?: 'MarkdownText', html: string } } }> } } };

export type StandingsGlobalPageQueryVariables = Exact<{
  entityType: LeaderboardEntityType;
  filter?: InputMaybe<LeaderboardFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
}>;


export type StandingsGlobalPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', id: string, leaderboard: { __typename?: 'LeaderboardConnection', edges: Array<{ __typename?: 'LeaderboardEdge', node: { __typename?: 'LeaderboardEntry', id: string, name: string, description: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } }>, me?: { __typename?: 'LeaderboardEntry', id: string, name: string, description: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } | null } } };

export type StandingsLocalPageQueryVariables = Exact<{
  entityType: LeaderboardEntityType;
  filter?: InputMaybe<LeaderboardFilter>;
}>;


export type StandingsLocalPageQuery = { __typename?: 'Query', me: { __typename?: 'User', church: { __typename?: 'Church', id: string, name: string } }, myCurrentProject: { __typename?: 'Project', id: string, leaderboard: { __typename?: 'LeaderboardConnection', edges: Array<{ __typename?: 'LeaderboardEdge', node: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } }>, me?: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } | null } } };

export type StandingsUnitPageQueryVariables = Exact<{ [key: string]: never; }>;


export type StandingsUnitPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', id: string, myTeam?: { __typename?: 'Team', id: string, name: string, memberLeaderboard: Array<{ __typename?: 'LeaderboardEntry', id: string, name: string, tags: Array<LeaderboardEntryTag>, rank?: number | null, score: number }> } | null } };

export type AdminSidebarQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminSidebarQuery = { __typename?: 'Query', projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, endDate: any, startDate: any } }> } };

export type CurrentProjectQueryVariables = Exact<{ [key: string]: never; }>;


export type CurrentProjectQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } };

export type AdminHomePageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminHomePageQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string }, projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, description: string, endDate: any, startDate: any, branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string }, dark: { __typename?: 'ColorSet', accent: string } } } } }> } };

export type AdminProjectAchievementPageQueryVariables = Exact<{
  achievementId: Scalars['ID']['input'];
}>;


export type AdminProjectAchievementPageQuery = { __typename?: 'Query', achievement:
    | { __typename?: 'ListeningAchievement', id: string, name: string, description: string, image?: string | null, achievedAt?: any | null, points: number, hidden: boolean, project: { __typename?: 'Project', id: string, name: string } }
    | { __typename?: 'QuizAchievement', id: string, name: string, description: string, image?: string | null, achievedAt?: any | null, points: number, hidden: boolean, project: { __typename?: 'Project', id: string, name: string } }
    | { __typename?: 'ReadingAchievement', id: string, name: string, description: string, image?: string | null, achievedAt?: any | null, points: number, hidden: boolean, project: { __typename?: 'Project', id: string, name: string } }
    | { __typename?: 'SimpleAchievement', id: string, name: string, description: string, image?: string | null, achievedAt?: any | null, points: number, hidden: boolean, project: { __typename?: 'Project', id: string, name: string } }
    | { __typename?: 'StreakAchievement', id: string, name: string, description: string, image?: string | null, achievedAt?: any | null, points: number, hidden: boolean, project: { __typename?: 'Project', id: string, name: string } }
   };

export type AdminProjectAchievementsNewPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectAchievementsNewPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string } };

export type AdminProjectChallengePageQueryVariables = Exact<{
  challengeId: Scalars['ID']['input'];
}>;


export type AdminProjectChallengePageQuery = { __typename?: 'Query', challenge:
    | { __typename?: 'ExternalChallenge', id: string, name: string, description: any, image?: string | null, buttonText: string, publishedAt?: any | null, endTime?: any | null, project: { __typename?: 'Project', id: string, name: string } }
    | { __typename?: 'QuizChallenge', id: string, name: string, description: any, image?: string | null, buttonText: string, publishedAt?: any | null, endTime?: any | null, project: { __typename?: 'Project', id: string, name: string } }
    | { __typename?: 'SimpleChallenge', id: string, name: string, description: any, image?: string | null, buttonText: string, publishedAt?: any | null, endTime?: any | null, project: { __typename?: 'Project', id: string, name: string } }
   };

export type AdminProjectChallengeNewPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectChallengeNewPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string }, events: { __typename?: 'EventConnection', edges: Array<{ __typename?: 'EventEdge', node: { __typename?: 'Event', id: string, name: string } }> } };

export type AdminProjectEditPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectEditPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, description: string, startDate: any, endDate: any, archivedAt?: boolean | null, branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } }, rules?: { __typename?: 'MarkdownText', markdown: string, html: string } | null } };

export type AdminProjectEventPageQueryVariables = Exact<{
  eventId: Scalars['ID']['input'];
}>;


export type AdminProjectEventPageQuery = { __typename?: 'Query', event: { __typename?: 'Event', id: string, name: string, description: string, startDate: any, endDate: any, parentProject: { __typename?: 'Project', id: string, name: string } } };

export type AdminProjectPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, description: string, startDate: any, endDate: any, branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string }, dark: { __typename?: 'ColorSet', accent: string } } } }, achievements: { __typename?: 'AchievementConnection', edges: Array<{ __typename?: 'AchievementEdge', node:
        | { __typename?: 'ListeningAchievement', id: string, name: string, description: string, image?: string | null, points: number, hidden: boolean }
        | { __typename?: 'QuizAchievement', id: string, name: string, description: string, image?: string | null, points: number, hidden: boolean }
        | { __typename?: 'ReadingAchievement', id: string, name: string, description: string, image?: string | null, points: number, hidden: boolean }
        | { __typename?: 'SimpleAchievement', id: string, name: string, description: string, image?: string | null, points: number, hidden: boolean }
        | { __typename?: 'StreakAchievement', id: string, name: string, description: string, image?: string | null, points: number, hidden: boolean }
       }> }, events: { __typename?: 'EventConnection', edges: Array<{ __typename?: 'EventEdge', node: { __typename?: 'Event', id: string, name: string, description: string } }> }, challenges: { __typename?: 'ChallengeConnection', edges: Array<{ __typename?: 'ChallengeEdge', node:
        | { __typename?: 'ExternalChallenge', id: string, name: string, description: any, image?: string | null }
        | { __typename?: 'QuizChallenge', id: string, name: string, description: any, image?: string | null }
        | { __typename?: 'SimpleChallenge', id: string, name: string, description: any, image?: string | null }
       }> }, streaks: { __typename?: 'StreakConnection', edges: Array<{ __typename?: 'StreakEdge', node: { __typename?: 'Streak', id: string, name: string, description: string } }> } };

export type AdminProjectStreakPageQueryVariables = Exact<{
  streakId: Scalars['ID']['input'];
}>;


export type AdminProjectStreakPageQuery = { __typename?: 'Query', streak: { __typename?: 'Streak', id: string, name: string, description: string, status: number, relevantDays: Array<{ __typename?: 'DateRange', start: any, end: any }>, project: { __typename?: 'Project', id: string, name: string } } };

export type AdminProjectsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminProjectsPageQuery = { __typename?: 'Query', projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, description: string, endDate: any, startDate: any, branding: { __typename?: 'Branding', logo?: string | null, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string }, dark: { __typename?: 'ColorSet', accent: string } } } } }> } };

export type AdminTeamPageQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type AdminTeamPageQuery = { __typename?: 'Query', team: { __typename?: 'Team', id: string, name: string, description: string, joinCode: string, members: Array<{ __typename?: 'TeamMember', id: string, name: string, isTeamLead: boolean, joinedAt: string, user: { __typename?: 'User', id: string, email: string, image?: string | null }, church: { __typename?: 'Church', id: string, name: string } }>, parentProject: { __typename?: 'Project', id: string, name: string }, superTeam?: { __typename?: 'SuperTeam', id: string, name: string } | null } };

export type AdminTeamsPageQueryVariables = Exact<{
  filter?: InputMaybe<TeamFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
}>;


export type AdminTeamsPageQuery = { __typename?: 'Query', teams: { __typename?: 'TeamConnection', totalCount: number, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null }, edges: Array<{ __typename?: 'TeamEdge', cursor: string, node: { __typename?: 'Team', id: string, name: string, description: string, members: Array<{ __typename?: 'TeamMember', id: string }>, parentProject: { __typename?: 'Project', id: string, name: string }, superTeam?: { __typename?: 'SuperTeam', id: string, name: string } | null } }> } };

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


export const ProjectRulesDocument = gql`
    query ProjectRules {
  myCurrentProject {
    rules {
      markdown
      html
    }
  }
}
    `;

export function useProjectRulesQuery(options?: Omit<Urql.UseQueryArgs<never, ProjectRulesQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ProjectRulesQuery, ProjectRulesQueryVariables | undefined>({ query: ProjectRulesDocument, variables: undefined, ...options });
};
export const PointHistoryDocument = gql`
    query PointHistory($last: Int) {
  myCurrentProject {
    journal(last: $last) {
      edges {
        node {
          id
          sourceType
          reason
          source {
            __typename
            ... on Achievement {
              id
              name
            }
            ... on Challenge {
              id
              name
            }
            ... on Event {
              id
              name
            }
            ... on SimpleAchievement {
              id
              name
            }
            ... on ReadingAchievement {
              id
              name
            }
            ... on ListeningAchievement {
              id
              name
            }
            ... on StreakAchievement {
              id
              name
            }
          }
          points
          createdAt
        }
      }
    }
  }
}
    `;

export function usePointHistoryQuery(options?: Omit<Urql.UseQueryArgs<never, PointHistoryQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<PointHistoryQuery, PointHistoryQueryVariables | undefined>({ query: PointHistoryDocument, variables: undefined, ...options });
};
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
    age
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
export const DeleteAchievementDocument = gql`
    mutation DeleteAchievement($id: ID!) {
  deleteAchievement(id: $id)
}
    `;

export function useDeleteAchievementMutation() {
  return Urql.useMutation<DeleteAchievementMutation, DeleteAchievementMutationVariables>(DeleteAchievementDocument);
};
export const UpdateAchievementDocument = gql`
    mutation UpdateAchievement($id: ID!, $input: UpdateAchievementInput!) {
  updateAchievement(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateAchievementMutation() {
  return Urql.useMutation<UpdateAchievementMutation, UpdateAchievementMutationVariables>(UpdateAchievementDocument);
};
export const CreateSimpleAchievementDocument = gql`
    mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
  createSimpleAchievement(input: $input) {
    id
  }
}
    `;

export function useCreateSimpleAchievementMutation() {
  return Urql.useMutation<CreateSimpleAchievementMutation, CreateSimpleAchievementMutationVariables>(CreateSimpleAchievementDocument);
};
export const CreateStreakAchievementDocument = gql`
    mutation CreateStreakAchievement($input: CreateStreakAchievementInput!) {
  createStreakAchievement(input: $input) {
    id
  }
}
    `;

export function useCreateStreakAchievementMutation() {
  return Urql.useMutation<CreateStreakAchievementMutation, CreateStreakAchievementMutationVariables>(CreateStreakAchievementDocument);
};
export const CreateReadingAchievementDocument = gql`
    mutation CreateReadingAchievement($input: CreateReadingAchievementInput!) {
  createReadingAchievement(input: $input) {
    id
  }
}
    `;

export function useCreateReadingAchievementMutation() {
  return Urql.useMutation<CreateReadingAchievementMutation, CreateReadingAchievementMutationVariables>(CreateReadingAchievementDocument);
};
export const CreateListeningAchievementDocument = gql`
    mutation CreateListeningAchievement($input: CreateListeningAchievementInput!) {
  createListeningAchievement(input: $input) {
    id
  }
}
    `;

export function useCreateListeningAchievementMutation() {
  return Urql.useMutation<CreateListeningAchievementMutation, CreateListeningAchievementMutationVariables>(CreateListeningAchievementDocument);
};
export const DeleteChallengeDocument = gql`
    mutation DeleteChallenge($id: ID!) {
  deleteChallenge(id: $id)
}
    `;

export function useDeleteChallengeMutation() {
  return Urql.useMutation<DeleteChallengeMutation, DeleteChallengeMutationVariables>(DeleteChallengeDocument);
};
export const UpdateChallengeDocument = gql`
    mutation UpdateChallenge($id: ID!, $input: UpdateChallengeInput!) {
  updateChallenge(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateChallengeMutation() {
  return Urql.useMutation<UpdateChallengeMutation, UpdateChallengeMutationVariables>(UpdateChallengeDocument);
};
export const CreateChallengeDocument = gql`
    mutation CreateChallenge($projectId: ID!, $eventId: ID!, $input: CreateChallengeInput!) {
  createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
    id
  }
}
    `;

export function useCreateChallengeMutation() {
  return Urql.useMutation<CreateChallengeMutation, CreateChallengeMutationVariables>(CreateChallengeDocument);
};
export const AcceptConsentDocument = gql`
    mutation AcceptConsent($consentId: ID!) {
  acceptConsent(consentId: $consentId) {
    id
    action
    actionDate
  }
}
    `;

export function useAcceptConsentMutation() {
  return Urql.useMutation<AcceptConsentMutation, AcceptConsentMutationVariables>(AcceptConsentDocument);
};
export const RejectConsentDocument = gql`
    mutation RejectConsent($consentId: ID!) {
  rejectConsent(consentId: $consentId) {
    id
    action
    actionDate
  }
}
    `;

export function useRejectConsentMutation() {
  return Urql.useMutation<RejectConsentMutation, RejectConsentMutationVariables>(RejectConsentDocument);
};
export const DeleteEventDocument = gql`
    mutation DeleteEvent($id: ID!) {
  deleteEvent(id: $id)
}
    `;

export function useDeleteEventMutation() {
  return Urql.useMutation<DeleteEventMutation, DeleteEventMutationVariables>(DeleteEventDocument);
};
export const UpdateEventDocument = gql`
    mutation UpdateEvent($id: ID!, $input: UpdateEventInput!) {
  updateEvent(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateEventMutation() {
  return Urql.useMutation<UpdateEventMutation, UpdateEventMutationVariables>(UpdateEventDocument);
};
export const CreateEventDocument = gql`
    mutation CreateEvent($projectId: ID!, $input: CreateEventInput!) {
  createEvent(projectId: $projectId, input: $input) {
    id
  }
}
    `;

export function useCreateEventMutation() {
  return Urql.useMutation<CreateEventMutation, CreateEventMutationVariables>(CreateEventDocument);
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
export const DeleteProjectDocument = gql`
    mutation DeleteProject($id: ID!) {
  deleteProject(id: $id)
}
    `;

export function useDeleteProjectMutation() {
  return Urql.useMutation<DeleteProjectMutation, DeleteProjectMutationVariables>(DeleteProjectDocument);
};
export const AssignRoleDocument = gql`
    mutation AssignRole($input: AssignRoleInput!) {
  assignRole(input: $input) {
    id
    role
    scope {
      id
      type
    }
  }
}
    `;

export function useAssignRoleMutation() {
  return Urql.useMutation<AssignRoleMutation, AssignRoleMutationVariables>(AssignRoleDocument);
};
export const RevokeRoleDocument = gql`
    mutation RevokeRole($input: RevokeRoleInput!) {
  revokeRole(input: $input)
}
    `;

export function useRevokeRoleMutation() {
  return Urql.useMutation<RevokeRoleMutation, RevokeRoleMutationVariables>(RevokeRoleDocument);
};
export const DeleteStreakDocument = gql`
    mutation DeleteStreak($id: ID!) {
  deleteStreak(id: $id)
}
    `;

export function useDeleteStreakMutation() {
  return Urql.useMutation<DeleteStreakMutation, DeleteStreakMutationVariables>(DeleteStreakDocument);
};
export const UpdateStreakDocument = gql`
    mutation UpdateStreak($id: ID!, $input: UpdateStreakInput!) {
  updateStreak(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateStreakMutation() {
  return Urql.useMutation<UpdateStreakMutation, UpdateStreakMutationVariables>(UpdateStreakDocument);
};
export const CreateStreakDocument = gql`
    mutation CreateStreak($input: CreateStreakInput!) {
  createStreak(input: $input) {
    id
  }
}
    `;

export function useCreateStreakMutation() {
  return Urql.useMutation<CreateStreakMutation, CreateStreakMutationVariables>(CreateStreakDocument);
};
export const CreateTeamDocument = gql`
    mutation CreateTeam($projectId: ID!, $input: CreateTeamInput!) {
  createTeam(projectId: $projectId, input: $input) {
    id
  }
}
    `;

export function useCreateTeamMutation() {
  return Urql.useMutation<CreateTeamMutation, CreateTeamMutationVariables>(CreateTeamDocument);
};
export const UpdateTeamDocument = gql`
    mutation UpdateTeam($id: ID!, $input: UpdateTeamInput!) {
  updateTeam(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateTeamMutation() {
  return Urql.useMutation<UpdateTeamMutation, UpdateTeamMutationVariables>(UpdateTeamDocument);
};
export const DeleteTeamDocument = gql`
    mutation DeleteTeam($id: ID!) {
  deleteTeam(id: $id)
}
    `;

export function useDeleteTeamMutation() {
  return Urql.useMutation<DeleteTeamMutation, DeleteTeamMutationVariables>(DeleteTeamDocument);
};
export const AddTeamMembersDocument = gql`
    mutation AddTeamMembers($teamId: ID!, $userIds: [ID!]!, $force: Boolean) {
  addTeamMembers(teamId: $teamId, userIds: $userIds, force: $force) {
    id
  }
}
    `;

export function useAddTeamMembersMutation() {
  return Urql.useMutation<AddTeamMembersMutation, AddTeamMembersMutationVariables>(AddTeamMembersDocument);
};
export const RemoveTeamMembersDocument = gql`
    mutation RemoveTeamMembers($teamId: ID!, $userIds: [ID!]!) {
  removeTeamMembers(teamId: $teamId, userIds: $userIds) {
    id
  }
}
    `;

export function useRemoveTeamMembersMutation() {
  return Urql.useMutation<RemoveTeamMembersMutation, RemoveTeamMembersMutationVariables>(RemoveTeamMembersDocument);
};
export const RegenerateJoinCodeDocument = gql`
    mutation RegenerateJoinCode($teamId: ID!) {
  regenerateJoinCode(teamId: $teamId) {
    id
    joinCode
  }
}
    `;

export function useRegenerateJoinCodeMutation() {
  return Urql.useMutation<RegenerateJoinCodeMutation, RegenerateJoinCodeMutationVariables>(RegenerateJoinCodeDocument);
};
export const AssignTeamLeadDocument = gql`
    mutation AssignTeamLead($teamId: ID!, $userId: ID!) {
  assignTeamLead(teamId: $teamId, userId: $userId) {
    id
  }
}
    `;

export function useAssignTeamLeadMutation() {
  return Urql.useMutation<AssignTeamLeadMutation, AssignTeamLeadMutationVariables>(AssignTeamLeadDocument);
};
export const ChallengePageDocument = gql`
    query ChallengePage($challengeId: ID!) {
  challenge(id: $challengeId) {
    __typename
    id
    name
    description
    userEnrolledAt
    userCompletedAt
    ... on SimpleChallenge {
      allowSelfCompletion
    }
    ... on ExternalChallenge {
      url
    }
    ... on QuizChallenge {
      quiz {
        id
        name
        description
        timeoutSeconds
        randomizeQuestions
        revealCorrectAnswers
        allowRetakes
        completionPoints
        publishedAt
        endTime
        userCanStart
        userSubmissions {
          id
          startedAt
          completedAt
          expiresAt
          isExpired
          score
          maxScore
          scorePercentage
          scorePercentage
          orderedQuestions {
            __typename
            id
            questionText
            questionOrder
            timeoutSeconds
            ... on NumberQuestion {
              minValue
              maxValue
              stepValue
            }
            ... on PredefinedQuestion {
              allowMultipleSelection
              predefinedAnswers {
                id
                answerText
                answerOrder
                isCorrect
              }
            }
          }
          responses {
            __typename
            id
            answeredAt
            timeSpentSeconds
            ... on FreeTextResponse {
              textResponse
            }
            ... on JsonResponse {
              jsonResponse
            }
            ... on NumberResponse {
              numberResponse
            }
            ... on PredefinedResponse {
              isCorrect
              selectedAnswers {
                id
                answerText
                answerOrder
                isCorrect
              }
            }
          }
        }
      }
    }
  }
}
    `;

export function useChallengePageQuery(options?: Omit<Urql.UseQueryArgs<never, ChallengePageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ChallengePageQuery, ChallengePageQueryVariables | undefined>({ query: ChallengePageDocument, variables: undefined, ...options });
};
export const ChallengesPageDocument = gql`
    query ChallengesPage {
  myCurrentProject {
    challenges {
      id
      name
      description
      image
      buttonText
      publishedAt
      endTime
      visibleAt
    }
  }
}
    `;

export function useChallengesPageQuery(options?: Omit<Urql.UseQueryArgs<never, ChallengesPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ChallengesPageQuery, ChallengesPageQueryVariables | undefined>({ query: ChallengesPageDocument, variables: undefined, ...options });
};
export const ProfilePageDocument = gql`
    query ProfilePage {
  me {
    id
    name
    consentStatus {
      pendingConsents {
        __typename
        id
        key
        version
        title
        body {
          html
        }
        managementType
        managedBy
      }
    }
  }
  myCurrentProject {
    id
    name
    achievements {
      id
      name
      description
      image
      hidden
      achievedAt
      points
    }
    leaderboard(entityType: PERSONS) {
      me {
        score
        rank
      }
    }
  }
}
    `;

export function useProfilePageQuery(options?: Omit<Urql.UseQueryArgs<never, ProfilePageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ProfilePageQuery, ProfilePageQueryVariables | undefined>({ query: ProfilePageDocument, variables: undefined, ...options });
};
export const ConsentsPageDocument = gql`
    query ConsentsPage {
  me {
    consentStatus {
      pendingConsents {
        __typename
        id
        key
        version
        title
        body {
          html
        }
        publishedAt
        managedBy
        managementType
        url
      }
      acceptedConsents {
        __typename
        id
        consent {
          title
          body {
            html
          }
          managedBy
          managementType
          url
        }
        action
        actionDate
      }
      rejectedConsents {
        __typename
        id
        consent {
          title
          body {
            html
          }
          managedBy
          managementType
          url
        }
        action
        actionDate
      }
    }
  }
}
    `;

export function useConsentsPageQuery(options?: Omit<Urql.UseQueryArgs<never, ConsentsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ConsentsPageQuery, ConsentsPageQueryVariables | undefined>({ query: ConsentsPageDocument, variables: undefined, ...options });
};
export const StandingsGlobalPageDocument = gql`
    query StandingsGlobalPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter, $first: Int) {
  myCurrentProject {
    id
    leaderboard(entityType: $entityType, filter: $filter, first: $first) {
      edges {
        node {
          id
          name
          description
          score
          rank
          tags
        }
      }
      me {
        id
        name
        description
        score
        rank
        tags
      }
    }
  }
}
    `;

export function useStandingsGlobalPageQuery(options?: Omit<Urql.UseQueryArgs<never, StandingsGlobalPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<StandingsGlobalPageQuery, StandingsGlobalPageQueryVariables | undefined>({ query: StandingsGlobalPageDocument, variables: undefined, ...options });
};
export const StandingsLocalPageDocument = gql`
    query StandingsLocalPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter) {
  me {
    church {
      id
      name
    }
  }
  myCurrentProject {
    id
    leaderboard(entityType: $entityType, filter: $filter) {
      edges {
        node {
          id
          name
          score
          rank
          tags
        }
      }
      me {
        id
        name
        score
        rank
        tags
      }
    }
  }
}
    `;

export function useStandingsLocalPageQuery(options?: Omit<Urql.UseQueryArgs<never, StandingsLocalPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<StandingsLocalPageQuery, StandingsLocalPageQueryVariables | undefined>({ query: StandingsLocalPageDocument, variables: undefined, ...options });
};
export const StandingsUnitPageDocument = gql`
    query StandingsUnitPage {
  myCurrentProject {
    id
    myTeam {
      id
      name
      memberLeaderboard {
        id
        name
        tags
        rank
        score
      }
    }
  }
}
    `;

export function useStandingsUnitPageQuery(options?: Omit<Urql.UseQueryArgs<never, StandingsUnitPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<StandingsUnitPageQuery, StandingsUnitPageQueryVariables | undefined>({ query: StandingsUnitPageDocument, variables: undefined, ...options });
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
        dark {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
        light {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
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
            light {
              accent
            }
            dark {
              accent
            }
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
export const AdminProjectAchievementPageDocument = gql`
    query AdminProjectAchievementPage($achievementId: ID!) {
  achievement(id: $achievementId) {
    id
    name
    description
    image
    achievedAt
    points
    hidden
    project {
      id
      name
    }
  }
}
    `;

export function useAdminProjectAchievementPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectAchievementPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectAchievementPageQuery, AdminProjectAchievementPageQueryVariables | undefined>({ query: AdminProjectAchievementPageDocument, variables: undefined, ...options });
};
export const AdminProjectAchievementsNewPageDocument = gql`
    query AdminProjectAchievementsNewPage($projectId: ID!) {
  project(id: $projectId) {
    id
    name
  }
}
    `;

export function useAdminProjectAchievementsNewPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectAchievementsNewPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectAchievementsNewPageQuery, AdminProjectAchievementsNewPageQueryVariables | undefined>({ query: AdminProjectAchievementsNewPageDocument, variables: undefined, ...options });
};
export const AdminProjectChallengePageDocument = gql`
    query AdminProjectChallengePage($challengeId: ID!) {
  challenge(id: $challengeId) {
    id
    name
    description
    image
    buttonText
    publishedAt
    endTime
    project {
      id
      name
    }
  }
}
    `;

export function useAdminProjectChallengePageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectChallengePageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectChallengePageQuery, AdminProjectChallengePageQueryVariables | undefined>({ query: AdminProjectChallengePageDocument, variables: undefined, ...options });
};
export const AdminProjectChallengeNewPageDocument = gql`
    query AdminProjectChallengeNewPage($projectId: ID!) {
  project(id: $projectId) {
    id
    name
  }
  events(first: 100, filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
      }
    }
  }
}
    `;

export function useAdminProjectChallengeNewPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectChallengeNewPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectChallengeNewPageQuery, AdminProjectChallengeNewPageQueryVariables | undefined>({ query: AdminProjectChallengeNewPageDocument, variables: undefined, ...options });
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
        light {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
        dark {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
      }
    }
    rules {
      markdown
      html
    }
  }
}
    `;

export function useAdminProjectEditPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectEditPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectEditPageQuery, AdminProjectEditPageQueryVariables | undefined>({ query: AdminProjectEditPageDocument, variables: undefined, ...options });
};
export const AdminProjectEventPageDocument = gql`
    query AdminProjectEventPage($eventId: ID!) {
  event(id: $eventId) {
    id
    name
    description
    startDate
    endDate
    parentProject {
      id
      name
    }
  }
}
    `;

export function useAdminProjectEventPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectEventPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectEventPageQuery, AdminProjectEventPageQueryVariables | undefined>({ query: AdminProjectEventPageDocument, variables: undefined, ...options });
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
        light {
          accent
        }
        dark {
          accent
        }
      }
    }
  }
  achievements(first: 50, filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
        description
        image
        points
        hidden
      }
    }
  }
  events(first: 50, filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
        description
      }
    }
  }
  challenges(first: 50, filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
        description
        image
      }
    }
  }
  streaks(first: 50, filter: {projectId: $projectId}) {
    edges {
      node {
        id
        name
        description
      }
    }
  }
}
    `;

export function useAdminProjectPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectPageQuery, AdminProjectPageQueryVariables | undefined>({ query: AdminProjectPageDocument, variables: undefined, ...options });
};
export const AdminProjectStreakPageDocument = gql`
    query AdminProjectStreakPage($streakId: ID!) {
  streak(id: $streakId) {
    id
    name
    description
    status
    relevantDays {
      start
      end
    }
    project {
      id
      name
    }
  }
}
    `;

export function useAdminProjectStreakPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectStreakPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectStreakPageQuery, AdminProjectStreakPageQueryVariables | undefined>({ query: AdminProjectStreakPageDocument, variables: undefined, ...options });
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
            light {
              accent
            }
            dark {
              accent
            }
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
export const AdminTeamPageDocument = gql`
    query AdminTeamPage($id: ID!) {
  team(id: $id) {
    id
    name
    description
    joinCode
    members {
      id
      name
      isTeamLead
      joinedAt
      user {
        id
        email
        image
      }
      church {
        id
        name
      }
    }
    parentProject {
      id
      name
    }
    superTeam {
      id
      name
    }
  }
}
    `;

export function useAdminTeamPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminTeamPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminTeamPageQuery, AdminTeamPageQueryVariables | undefined>({ query: AdminTeamPageDocument, variables: undefined, ...options });
};
export const AdminTeamsPageDocument = gql`
    query AdminTeamsPage($filter: TeamFilter, $first: Int, $after: String, $last: Int, $before: String) {
  teams(
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
        description
        members {
          id
        }
        parentProject {
          id
          name
        }
        superTeam {
          id
          name
        }
      }
    }
  }
}
    `;

export function useAdminTeamsPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminTeamsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminTeamsPageQuery, AdminTeamsPageQueryVariables | undefined>({ query: AdminTeamsPageDocument, variables: undefined, ...options });
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