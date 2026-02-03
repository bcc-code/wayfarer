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
  awardableFrom?: Maybe<Scalars['DateTime']['output']>;
  celebratedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  descriptionCompleted: Scalars['String']['output'];
  descriptionPending: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  /** @deprecated Use imageCompletedObject instead */
  imageCompleted: Scalars['String']['output'];
  imageCompletedObject: Image;
  /** @deprecated Use imagePendingObject instead */
  imagePending: Scalars['String']['output'];
  imagePendingObject: Image;
  name: Scalars['String']['output'];
  notificationText: Scalars['String']['output'];
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

export type AdminDashboardStats = {
  __typename?: 'AdminDashboardStats';
  activeProjectsCount: Scalars['Int']['output'];
  newUsersLast7Days: Scalars['Int']['output'];
  totalChallenges: Scalars['Int']['output'];
  totalPointsAwarded: Scalars['Int']['output'];
  totalProjects: Scalars['Int']['output'];
  totalUsers: Scalars['Int']['output'];
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

export type AssignRoleInput = {
  role: RoleType;
  scopeId?: InputMaybe<Scalars['ID']['input']>;
  scopeType?: InputMaybe<ScopeType>;
  userId: Scalars['ID']['input'];
};

export type Branding = {
  __typename?: 'Branding';
  /** @deprecated Use bannerImage instead */
  banner?: Maybe<Scalars['String']['output']>;
  bannerImage?: Maybe<Image>;
  colors: Colors;
  /** @deprecated Use logoImage instead */
  logo?: Maybe<Scalars['String']['output']>;
  logoImage?: Maybe<Image>;
  rounding: Scalars['Int']['output'];
};

export type BrandingInput = {
  banner?: InputMaybe<Scalars['String']['input']>;
  colors: ColorsInput;
  logo?: InputMaybe<Scalars['String']['input']>;
  rounding: Scalars['Int']['input'];
};

export type Challenge = {
  buttonText?: Maybe<Scalars['String']['output']>;
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
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
  Plugin = 'PLUGIN',
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
  shortText: Scalars['String']['output'];
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

export type ContentAchievement = Achievement & {
  __typename?: 'ContentAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  awardableFrom?: Maybe<Scalars['DateTime']['output']>;
  celebratedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  completedItemCount: Scalars['Int']['output'];
  descriptionCompleted: Scalars['String']['output'];
  descriptionPending: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  /** @deprecated Use imageCompletedObject instead */
  imageCompleted: Scalars['String']['output'];
  imageCompletedObject: Image;
  /** @deprecated Use imagePendingObject instead */
  imagePending: Scalars['String']['output'];
  imagePendingObject: Image;
  items: Array<ContentItem>;
  name: Scalars['String']['output'];
  nextItem?: Maybe<ContentItem>;
  notificationText: Scalars['String']['output'];
  points: Scalars['Int']['output'];
  project: Project;
  totalItems: Scalars['Int']['output'];
  userCompletedItems: Array<ContentItem>;
};

export type ContentItem = {
  __typename?: 'ContentItem';
  externalContent: ExternalContent;
  id: Scalars['ID']['output'];
  sortOrder: Scalars['Int']['output'];
};

export type ContentItemInput = {
  externalContentId: Scalars['ID']['input'];
};

export type CreateChallengeInput = {
  allowSelfCompletion?: InputMaybe<Scalars['Boolean']['input']>;
  buttonText?: InputMaybe<Scalars['String']['input']>;
  description?: InputMaybe<Scalars['HTML']['input']>;
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  pluginChallengeId?: InputMaybe<Scalars['String']['input']>;
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

export type CreateContentAchievementFromExternalContentInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted: Scalars['String']['input'];
  descriptionPending: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  externalContentIds: Array<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  imageCompleted: Scalars['String']['input'];
  imagePending: Scalars['String']['input'];
  name: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateContentAchievementInput = {
  awardableFrom?: InputMaybe<Scalars['DateTime']['input']>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted: Scalars['String']['input'];
  descriptionPending: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  imageCompleted: Scalars['String']['input'];
  imagePending: Scalars['String']['input'];
  items: Array<ContentItemInput>;
  name: Scalars['String']['input'];
  notificationText: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateEventInput = {
  description: Scalars['String']['input'];
  endDate: Scalars['DateTime']['input'];
  name: Scalars['String']['input'];
  startDate: Scalars['DateTime']['input'];
};

export type CreateOrderingItemInput = {
  correctOrder: Scalars['Int']['input'];
  itemText: Scalars['String']['input'];
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
  infoMessage?: InputMaybe<Scalars['String']['input']>;
  infoMessageEnd?: InputMaybe<Scalars['DateTime']['input']>;
  infoMessageStart?: InputMaybe<Scalars['DateTime']['input']>;
  name: Scalars['String']['input'];
  rules?: InputMaybe<Scalars['String']['input']>;
  startDate: Scalars['DateTime']['input'];
};

export type CreateQuizAchievementInput = {
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted: Scalars['String']['input'];
  descriptionPending: Scalars['String']['input'];
  hidden: Scalars['Boolean']['input'];
  imageCompleted: Scalars['String']['input'];
  imagePending: Scalars['String']['input'];
  minScorePercentage?: InputMaybe<Scalars['Int']['input']>;
  name: Scalars['String']['input'];
  notificationText: Scalars['String']['input'];
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
  randomizeQuestions: Scalars['Boolean']['input'];
  revealCorrectAnswers: Scalars['Boolean']['input'];
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
};

export type CreateQuizQuestionInput = {
  allowMultipleSelection?: InputMaybe<Scalars['Boolean']['input']>;
  maxValue?: InputMaybe<Scalars['Float']['input']>;
  minValue?: InputMaybe<Scalars['Float']['input']>;
  orderingItems?: InputMaybe<Array<CreateOrderingItemInput>>;
  points?: InputMaybe<Scalars['Int']['input']>;
  predefinedAnswers?: InputMaybe<Array<CreatePredefinedAnswerInput>>;
  questionOrder: Scalars['Int']['input'];
  questionText: Scalars['String']['input'];
  questionType: QuizQuestionType;
  stepValue?: InputMaybe<Scalars['Float']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
};

export type CreateQuizSessionInput = {
  finishAt?: InputMaybe<Scalars['DateTime']['input']>;
  lockAt?: InputMaybe<Scalars['DateTime']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  openAt?: InputMaybe<Scalars['DateTime']['input']>;
  quizId: Scalars['ID']['input'];
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
  awardableFrom?: InputMaybe<Scalars['DateTime']['input']>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted: Scalars['String']['input'];
  descriptionPending: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  imageCompleted: Scalars['String']['input'];
  imagePending: Scalars['String']['input'];
  name: Scalars['String']['input'];
  notificationText: Scalars['String']['input'];
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateStreakAchievementInput = {
  awardableFrom?: InputMaybe<Scalars['DateTime']['input']>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted: Scalars['String']['input'];
  descriptionPending: Scalars['String']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden: Scalars['Boolean']['input'];
  imageCompleted: Scalars['String']['input'];
  imagePending: Scalars['String']['input'];
  name: Scalars['String']['input'];
  neededStreak: Scalars['Int']['input'];
  notificationText: Scalars['String']['input'];
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

export type CreateTeamScoreAdjustmentInput = {
  distributionMode: TeamScoreDistributionMode;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  points: Scalars['Int']['input'];
  projectId: Scalars['ID']['input'];
  reason?: InputMaybe<Scalars['String']['input']>;
  teamId: Scalars['ID']['input'];
};

export type CreateUserInput = {
  age: Scalars['Int']['input'];
  churchId: Scalars['ID']['input'];
  email: Scalars['String']['input'];
  gender: Gender;
  membersId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type CreateWebhookInput = {
  eventType: WebhookEventType;
  includeEventData?: InputMaybe<Scalars['Boolean']['input']>;
  includeUserData?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
  secret?: InputMaybe<Scalars['String']['input']>;
  url: Scalars['String']['input'];
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

export type DeviceMetadata = {
  appVersion?: InputMaybe<Scalars['String']['input']>;
  contextUrl?: InputMaybe<Scalars['String']['input']>;
  locale?: InputMaybe<Scalars['String']['input']>;
  platform: Scalars['String']['input'];
  screenHeight: Scalars['Int']['input'];
  screenWidth: Scalars['Int']['input'];
  timezone?: InputMaybe<Scalars['String']['input']>;
  userAgent: Scalars['String']['input'];
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
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
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

export type ExternalContent = {
  __typename?: 'ExternalContent';
  contentId?: Maybe<Scalars['String']['output']>;
  contentType: ExternalContentType;
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  planId: Scalars['String']['output'];
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  source: Scalars['String']['output'];
  syncedAt: Scalars['DateTime']['output'];
  taskId: Scalars['String']['output'];
  title?: Maybe<Scalars['String']['output']>;
  translations: Array<ExternalContentTranslation>;
  updatedAt: Scalars['DateTime']['output'];
  url?: Maybe<Scalars['String']['output']>;
};

export type ExternalContentConnection = {
  __typename?: 'ExternalContentConnection';
  edges: Array<ExternalContentEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ExternalContentEdge = {
  __typename?: 'ExternalContentEdge';
  cursor: Scalars['String']['output'];
  node: ExternalContent;
};

export type ExternalContentFilter = {
  contentId?: InputMaybe<Scalars['String']['input']>;
  contentType?: InputMaybe<ExternalContentType>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  planId?: InputMaybe<Scalars['String']['input']>;
  publishedAfter?: InputMaybe<Scalars['DateTime']['input']>;
  publishedBefore?: InputMaybe<Scalars['DateTime']['input']>;
  source?: InputMaybe<Scalars['String']['input']>;
  taskId?: InputMaybe<Scalars['String']['input']>;
};

export enum ExternalContentSortBy {
  CreatedAtAsc = 'CREATED_AT_ASC',
  CreatedAtDesc = 'CREATED_AT_DESC',
  PublishedAtAsc = 'PUBLISHED_AT_ASC',
  PublishedAtDesc = 'PUBLISHED_AT_DESC'
}

export type ExternalContentTranslation = {
  __typename?: 'ExternalContentTranslation';
  languageCode: Scalars['String']['output'];
  title?: Maybe<Scalars['String']['output']>;
};

export enum ExternalContentType {
  Article = 'ARTICLE',
  BibleVerse = 'BIBLE_VERSE',
  BookChapter = 'BOOK_CHAPTER',
  ExternalLink = 'EXTERNAL_LINK',
  Media = 'MEDIA',
  Quiz = 'QUIZ',
  Song = 'SONG',
  Text = 'TEXT'
}

export type FeedbackConnection = {
  __typename?: 'FeedbackConnection';
  edges: Array<FeedbackEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type FeedbackEdge = {
  __typename?: 'FeedbackEdge';
  cursor: Scalars['String']['output'];
  node: UserFeedback;
};

export type FeedbackFilter = {
  tags?: InputMaybe<Array<Scalars['String']['input']>>;
  userId?: InputMaybe<Scalars['ID']['input']>;
};

export type FileUpload = {
  __typename?: 'FileUpload';
  blurhash?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['DateTime']['output'];
  fileSize: Scalars['Int']['output'];
  filename: Scalars['String']['output'];
  height?: Maybe<Scalars['Int']['output']>;
  id: Scalars['ID']['output'];
  mimeType: Scalars['String']['output'];
  publicUrl: Scalars['String']['output'];
  storedFilename: Scalars['String']['output'];
  uploadedBy: Scalars['ID']['output'];
  width?: Maybe<Scalars['Int']['output']>;
};

export type FirebaseTokenResponse = {
  __typename?: 'FirebaseTokenResponse';
  expiresIn: Scalars['Int']['output'];
  token: Scalars['String']['output'];
};

export enum ForwardDestination {
  BccMediaSupport = 'BCC_MEDIA_SUPPORT',
  SsfTicket = 'SSF_TICKET'
}

export type FreeTextQuestion = QuizQuestion & {
  __typename?: 'FreeTextQuestion';
  id: Scalars['ID']['output'];
  points?: Maybe<Scalars['Int']['output']>;
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export type FreeTextResponse = QuizResponse & {
  __typename?: 'FreeTextResponse';
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  pointsEarned?: Maybe<Scalars['Int']['output']>;
  question: QuizQuestion;
  submission: QuizSubmission;
  textResponse: Scalars['String']['output'];
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

export enum Gender {
  Female = 'FEMALE',
  Male = 'MALE'
}

export type GrantQuizSessionAccessInput = {
  allProjectUsers?: InputMaybe<Scalars['Boolean']['input']>;
  churchIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  sessionId: Scalars['ID']['input'];
  superTeamIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  teamIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  userIds?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type Image = {
  __typename?: 'Image';
  blurhash?: Maybe<Scalars['String']['output']>;
  height?: Maybe<Scalars['Int']['output']>;
  url: Scalars['String']['output'];
  width?: Maybe<Scalars['Int']['output']>;
};

export type JsonQuestion = QuizQuestion & {
  __typename?: 'JsonQuestion';
  id: Scalars['ID']['output'];
  points?: Maybe<Scalars['Int']['output']>;
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
  pointsEarned?: Maybe<Scalars['Int']['output']>;
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
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
  lastScoreAt?: Maybe<Scalars['DateTime']['output']>;
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
  adminSetUserConsent: UserConsentHistoryEntry;
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
  clearAllCache: Scalars['Boolean']['output'];
  completeChallenge: Challenge;
  createChallenge: Challenge;
  createConsent: Consent;
  createContentAchievement: ContentAchievement;
  createContentAchievementFromExternalContent: ContentAchievement;
  createEvent: Event;
  createProject: Project;
  createQuiz: Quiz;
  createQuizAchievement: QuizAchievement;
  createQuizSession: QuizSession;
  createQuizSubmission: QuizSubmission;
  createScoreAdjustment: ScoreJournal;
  createSimpleAchievement: SimpleAchievement;
  createStreak: Streak;
  createStreakAchievement: StreakAchievement;
  createSuperTeam: SuperTeam;
  createTeam: Team;
  createTeamScoreAdjustment: Array<ScoreJournal>;
  createWebhook: Webhook;
  deleteAchievement: Scalars['Boolean']['output'];
  deleteChallenge: Scalars['Boolean']['output'];
  deleteEvent: Scalars['Boolean']['output'];
  deleteFeedback: Scalars['Boolean']['output'];
  deleteProject: Scalars['Boolean']['output'];
  deleteQuiz: Scalars['Boolean']['output'];
  deleteQuizQuestion: Scalars['Boolean']['output'];
  deleteQuizSession: Scalars['Boolean']['output'];
  deleteScoreJournalEntry: Scalars['Boolean']['output'];
  deleteStreak: Scalars['Boolean']['output'];
  deleteSuperTeam: Scalars['Boolean']['output'];
  deleteTeam: Scalars['Boolean']['output'];
  deleteWebhook: Scalars['Boolean']['output'];
  enrollInChallenge: Challenge;
  enrollUserInChallenge: Challenge;
  finalizeQuiz: QuizSubmission;
  finishQuizSession: QuizSession;
  forwardFeedbackToDesk: Scalars['Boolean']['output'];
  grantQuizSessionAccess: Scalars['Int']['output'];
  joinEvent: Event;
  joinProject: Project;
  joinTeam: Team;
  linkAchievementToChallenge: Achievement;
  lockQuizSession: QuizSession;
  markAchievementCelebrated: Scalars['Boolean']['output'];
  markContentItemCompleted: Array<ContentAchievement>;
  markFeedbackHandled: UserFeedback;
  moveEvent: Event;
  openQuizSession: QuizSession;
  publishChallenge: Challenge;
  recordStreakActivity: StreakAchievement;
  regenerateJoinCode: Team;
  registerPushSubscription: PushSubscription;
  rejectConsent: UserConsent;
  removeTeamMembers: Team;
  removeUserFromProject: User;
  reopenQuizSession: QuizSession;
  reorderAchievements: Array<Achievement>;
  reorderQuizQuestions: Array<QuizQuestion>;
  resetQuizSessionSubmission: Scalars['Boolean']['output'];
  revokeAchievement: Scalars['Boolean']['output'];
  revokeAllQuizSessionAccess: Scalars['Boolean']['output'];
  revokeQuizSessionAccess: Scalars['Boolean']['output'];
  revokeRole: Scalars['Boolean']['output'];
  revokeSuperTeamAchievement: Scalars['Boolean']['output'];
  revokeTeamAchievement: Scalars['Boolean']['output'];
  selfCompleteChallenge: SimpleChallenge;
  sendPushNotification: SendPushNotificationResult;
  setChallengeRequirements: Challenge;
  setChallengeVisibility: Challenge;
  setNotificationPreference: PushNotificationPreference;
  startQuizSession: QuizSubmission;
  submitFeedback: UserFeedback;
  submitQuizAnswer: QuizResponse;
  testWebhook: WebhookLog;
  uncompleteChallenge: Scalars['Boolean']['output'];
  unenrollFromChallenge: Scalars['Boolean']['output'];
  unenrollUserFromChallenge: Scalars['Boolean']['output'];
  unmarkContentItemCompleted: Array<ContentAchievement>;
  unregisterPushSubscription: Scalars['Boolean']['output'];
  updateAchievement: Achievement;
  updateAvatar: User;
  updateChallenge: Challenge;
  updateChurch: Church;
  updateConsent: Consent;
  updateContentAchievement: ContentAchievement;
  updateEvent: Event;
  updateFeedbackTags: UserFeedback;
  updateProject: Project;
  updateQuiz: Quiz;
  updateQuizAnswer: QuizResponse;
  updateQuizQuestion: QuizQuestion;
  updateQuizSession: QuizSession;
  updateStreak: Streak;
  updateStreakAchievement: StreakAchievement;
  updateSuperTeam: SuperTeam;
  updateTeam: Team;
  updateWebhook: Webhook;
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


export type MutationAdminSetUserConsentArgs = {
  action: ConsentAction;
  consentId: Scalars['ID']['input'];
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
  eventId?: InputMaybe<Scalars['ID']['input']>;
  input: CreateChallengeInput;
  projectId: Scalars['ID']['input'];
};


export type MutationCreateConsentArgs = {
  body: Scalars['String']['input'];
  isRemote?: InputMaybe<Scalars['Boolean']['input']>;
  key: Scalars['String']['input'];
  managedBy?: InputMaybe<Scalars['String']['input']>;
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  shortText?: InputMaybe<Scalars['String']['input']>;
  title: Scalars['String']['input'];
  url?: InputMaybe<Scalars['String']['input']>;
};


export type MutationCreateContentAchievementArgs = {
  input: CreateContentAchievementInput;
};


export type MutationCreateContentAchievementFromExternalContentArgs = {
  input: CreateContentAchievementFromExternalContentInput;
};


export type MutationCreateEventArgs = {
  input: CreateEventInput;
  projectId: Scalars['ID']['input'];
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


export type MutationCreateQuizSessionArgs = {
  input: CreateQuizSessionInput;
};


export type MutationCreateQuizSubmissionArgs = {
  completedAt?: InputMaybe<Scalars['DateTime']['input']>;
  quizId: Scalars['ID']['input'];
  responses: Array<SubmitQuizAnswerInput>;
  userId: Scalars['ID']['input'];
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


export type MutationCreateTeamScoreAdjustmentArgs = {
  input: CreateTeamScoreAdjustmentInput;
};


export type MutationCreateWebhookArgs = {
  input: CreateWebhookInput;
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


export type MutationDeleteFeedbackArgs = {
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


export type MutationDeleteQuizSessionArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteScoreJournalEntryArgs = {
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


export type MutationDeleteWebhookArgs = {
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


export type MutationFinishQuizSessionArgs = {
  id: Scalars['ID']['input'];
};


export type MutationForwardFeedbackToDeskArgs = {
  destination: ForwardDestination;
  feedbackId: Scalars['ID']['input'];
};


export type MutationGrantQuizSessionAccessArgs = {
  input: GrantQuizSessionAccessInput;
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


export type MutationLockQuizSessionArgs = {
  id: Scalars['ID']['input'];
};


export type MutationMarkAchievementCelebratedArgs = {
  achievementId: Scalars['ID']['input'];
};


export type MutationMarkContentItemCompletedArgs = {
  externalContentId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationMarkFeedbackHandledArgs = {
  feedbackId: Scalars['ID']['input'];
};


export type MutationMoveEventArgs = {
  id: Scalars['ID']['input'];
  newProjectId: Scalars['ID']['input'];
};


export type MutationOpenQuizSessionArgs = {
  id: Scalars['ID']['input'];
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


export type MutationRegisterPushSubscriptionArgs = {
  input: RegisterPushSubscriptionInput;
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


export type MutationReopenQuizSessionArgs = {
  id: Scalars['ID']['input'];
};


export type MutationReorderAchievementsArgs = {
  achievementIds: Array<Scalars['ID']['input']>;
  projectId: Scalars['ID']['input'];
};


export type MutationReorderQuizQuestionsArgs = {
  questionIds: Array<Scalars['ID']['input']>;
  quizId: Scalars['ID']['input'];
};


export type MutationResetQuizSessionSubmissionArgs = {
  sessionId: Scalars['ID']['input'];
};


export type MutationRevokeAchievementArgs = {
  achievementId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationRevokeAllQuizSessionAccessArgs = {
  sessionId: Scalars['ID']['input'];
};


export type MutationRevokeQuizSessionAccessArgs = {
  sessionId: Scalars['ID']['input'];
  userIds: Array<Scalars['ID']['input']>;
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


export type MutationSendPushNotificationArgs = {
  input: SendPushNotificationInput;
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


export type MutationSetNotificationPreferenceArgs = {
  input: SetNotificationPreferenceInput;
};


export type MutationStartQuizSessionArgs = {
  sessionId: Scalars['ID']['input'];
};


export type MutationSubmitFeedbackArgs = {
  input: SubmitFeedbackInput;
};


export type MutationSubmitQuizAnswerArgs = {
  input: SubmitQuizAnswerInput;
  submissionId: Scalars['ID']['input'];
};


export type MutationTestWebhookArgs = {
  id: Scalars['ID']['input'];
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


export type MutationUnmarkContentItemCompletedArgs = {
  externalContentId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationUnregisterPushSubscriptionArgs = {
  endpoint: Scalars['String']['input'];
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


export type MutationUpdateChurchArgs = {
  id: Scalars['ID']['input'];
  input: UpdateChurchInput;
};


export type MutationUpdateConsentArgs = {
  body?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  managedBy?: InputMaybe<Scalars['String']['input']>;
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  shortText?: InputMaybe<Scalars['String']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
  url?: InputMaybe<Scalars['String']['input']>;
};


export type MutationUpdateContentAchievementArgs = {
  id: Scalars['ID']['input'];
  input: UpdateContentAchievementInput;
};


export type MutationUpdateEventArgs = {
  id: Scalars['ID']['input'];
  input: UpdateEventInput;
};


export type MutationUpdateFeedbackTagsArgs = {
  feedbackId: Scalars['ID']['input'];
  tags: Array<Scalars['String']['input']>;
};


export type MutationUpdateProjectArgs = {
  id: Scalars['ID']['input'];
  input: UpdateProjectInput;
};


export type MutationUpdateQuizArgs = {
  id: Scalars['ID']['input'];
  input: UpdateQuizInput;
};


export type MutationUpdateQuizAnswerArgs = {
  input: UpdateQuizAnswerInput;
  responseId: Scalars['ID']['input'];
};


export type MutationUpdateQuizQuestionArgs = {
  id: Scalars['ID']['input'];
  input: UpdateQuizQuestionInput;
};


export type MutationUpdateQuizSessionArgs = {
  id: Scalars['ID']['input'];
  input: UpdateQuizSessionInput;
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


export type MutationUpdateWebhookArgs = {
  id: Scalars['ID']['input'];
  input: UpdateWebhookInput;
};

export enum NotificationType {
  AchievementUnlocked = 'ACHIEVEMENT_UNLOCKED',
  ChallengeAvailable = 'CHALLENGE_AVAILABLE',
  Generic = 'GENERIC'
}

export type NumberQuestion = QuizQuestion & {
  __typename?: 'NumberQuestion';
  id: Scalars['ID']['output'];
  maxValue?: Maybe<Scalars['Float']['output']>;
  minValue?: Maybe<Scalars['Float']['output']>;
  points?: Maybe<Scalars['Int']['output']>;
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
  pointsEarned?: Maybe<Scalars['Int']['output']>;
  question: QuizQuestion;
  submission: QuizSubmission;
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

export type OrderingQuestion = QuizQuestion & {
  __typename?: 'OrderingQuestion';
  id: Scalars['ID']['output'];
  orderingItems: Array<QuizOrderingItem>;
  points?: Maybe<Scalars['Int']['output']>;
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export type OrderingResponse = QuizResponse & {
  __typename?: 'OrderingResponse';
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  isCorrect?: Maybe<Scalars['Boolean']['output']>;
  pointsEarned?: Maybe<Scalars['Int']['output']>;
  question: QuizQuestion;
  submission: QuizSubmission;
  submittedOrder: Array<Scalars['ID']['output']>;
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor?: Maybe<Scalars['String']['output']>;
};

export type PluginChallenge = Challenge & {
  __typename?: 'PluginChallenge';
  buttonText?: Maybe<Scalars['String']['output']>;
  description: Scalars['HTML']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  event?: Maybe<Event>;
  id: Scalars['ID']['output'];
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
  name: Scalars['String']['output'];
  pluginChallengeId: Scalars['String']['output'];
  project: Project;
  publishedAt?: Maybe<Scalars['DateTime']['output']>;
  requiresSuperTeamMembership: Scalars['Boolean']['output'];
  requiresTeamMembership: Scalars['Boolean']['output'];
  startedAt?: Maybe<Scalars['DateTime']['output']>;
  userCompletedAt?: Maybe<Scalars['DateTime']['output']>;
  userEnrolledAt?: Maybe<Scalars['DateTime']['output']>;
  visibleAt?: Maybe<Scalars['DateTime']['output']>;
};

export type PredefinedQuestion = QuizQuestion & {
  __typename?: 'PredefinedQuestion';
  allowMultipleSelection: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  points?: Maybe<Scalars['Int']['output']>;
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
  pointsEarned?: Maybe<Scalars['Int']['output']>;
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
  infoMessage?: Maybe<MarkdownText>;
  infoMessageEnd?: Maybe<Scalars['DateTime']['output']>;
  infoMessageStart?: Maybe<Scalars['DateTime']['output']>;
  journal: ScoreJournalConnection;
  leaderboard: LeaderboardConnection;
  myChurchTeams: Array<Team>;
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

export type PushNotificationPreference = {
  __typename?: 'PushNotificationPreference';
  enabled: Scalars['Boolean']['output'];
  notificationType: NotificationType;
  updatedAt: Scalars['DateTime']['output'];
};

export type PushSubscription = {
  __typename?: 'PushSubscription';
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
};

export type Query = {
  __typename?: 'Query';
  achievement: Achievement;
  achievements: AchievementConnection;
  adminDashboardStats: AdminDashboardStats;
  adminScoreJournal: ScoreJournalConnection;
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
  externalContent: ExternalContent;
  externalContents: ExternalContentConnection;
  feedback: FeedbackConnection;
  fileUpload?: Maybe<FileUpload>;
  firebaseToken: FirebaseTokenResponse;
  instanceID: Scalars['String']['output'];
  me: User;
  myCurrentEvent: Event;
  myCurrentProject: Project;
  myEvents: Array<Event>;
  myProjects: Array<Project>;
  myPushNotificationPreferences: Array<PushNotificationPreference>;
  pendingConsents: Array<Consent>;
  project: Project;
  projects: ProjectConnection;
  pushNotificationsEnabled: Scalars['Boolean']['output'];
  quiz: Quiz;
  quizSession?: Maybe<QuizSession>;
  quizSessions: Array<QuizSession>;
  quizSubmission: QuizSubmission;
  quizSubmissions: QuizSubmissionConnection;
  quizzes: QuizConnection;
  scoreJournal: ScoreJournalConnection;
  streak: Streak;
  streaks: StreakConnection;
  superteam: SuperTeam;
  superteams: SuperTeamConnection;
  team: Team;
  teamByJoinCode?: Maybe<Team>;
  teams: TeamConnection;
  user: User;
  userRoles: Array<UserRole>;
  users: UserConnection;
  usersWithRole: Array<User>;
  vapidPublicKey: Scalars['String']['output'];
  webhook?: Maybe<Webhook>;
  webhooks: Array<Webhook>;
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


export type QueryAdminScoreJournalArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<ScoreJournalFilter>;
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


export type QueryExternalContentArgs = {
  id: Scalars['ID']['input'];
};


export type QueryExternalContentsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter: ExternalContentFilter;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  sortBy?: InputMaybe<ExternalContentSortBy>;
};


export type QueryFeedbackArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  filter?: InputMaybe<FeedbackFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryFileUploadArgs = {
  id: Scalars['ID']['input'];
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


export type QueryQuizSessionArgs = {
  id: Scalars['ID']['input'];
};


export type QueryQuizSessionsArgs = {
  quizId: Scalars['ID']['input'];
  state?: InputMaybe<QuizSessionState>;
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


export type QueryTeamByJoinCodeArgs = {
  code: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
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


export type QueryWebhookArgs = {
  id: Scalars['ID']['input'];
};


export type QueryWebhooksArgs = {
  projectId: Scalars['ID']['input'];
};

export type Quiz = {
  __typename?: 'Quiz';
  allowRetakes: Scalars['Boolean']['output'];
  challenge: Challenge;
  completionPoints: Scalars['Int']['output'];
  description: Scalars['String']['output'];
  endTime?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
  name: Scalars['String']['output'];
  project: Project;
  questions: Array<QuizQuestion>;
  randomizeQuestions: Scalars['Boolean']['output'];
  revealCorrectAnswers: Scalars['Boolean']['output'];
  sessions: Array<QuizSession>;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
  userActiveSession?: Maybe<QuizSession>;
  userActiveSubmission?: Maybe<QuizSubmission>;
  userCanStart: Scalars['Boolean']['output'];
  userSessions: Array<QuizSession>;
  userSubmissions: Array<QuizSubmission>;
};


export type QuizSessionsArgs = {
  state?: InputMaybe<QuizSessionState>;
};

export enum QuizAccessSourceType {
  All = 'ALL',
  Church = 'CHURCH',
  Direct = 'DIRECT',
  SuperTeam = 'SUPER_TEAM',
  Team = 'TEAM'
}

export type QuizAchievement = Achievement & {
  __typename?: 'QuizAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  awardableFrom?: Maybe<Scalars['DateTime']['output']>;
  celebratedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  descriptionCompleted: Scalars['String']['output'];
  descriptionPending: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  /** @deprecated Use imageCompletedObject instead */
  imageCompleted: Scalars['String']['output'];
  imageCompletedObject: Image;
  /** @deprecated Use imagePendingObject instead */
  imagePending: Scalars['String']['output'];
  imagePendingObject: Image;
  minScorePercentage?: Maybe<Scalars['Int']['output']>;
  name: Scalars['String']['output'];
  notificationText: Scalars['String']['output'];
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
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
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
};

export type QuizOrderingItem = {
  __typename?: 'QuizOrderingItem';
  id: Scalars['ID']['output'];
  itemText: Scalars['String']['output'];
  question: QuizQuestion;
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
  points?: Maybe<Scalars['Int']['output']>;
  questionOrder: Scalars['Int']['output'];
  questionText: Scalars['String']['output'];
  quiz: Quiz;
  timeoutSeconds?: Maybe<Scalars['Int']['output']>;
};

export enum QuizQuestionType {
  FreeText = 'FREE_TEXT',
  Json = 'JSON',
  Number = 'NUMBER',
  Ordering = 'ORDERING',
  Predefined = 'PREDEFINED'
}

export type QuizResponse = {
  answeredAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  pointsEarned?: Maybe<Scalars['Int']['output']>;
  question: QuizQuestion;
  submission: QuizSubmission;
  timeSpentSeconds?: Maybe<Scalars['Int']['output']>;
};

export type QuizSession = {
  __typename?: 'QuizSession';
  accessCount: Scalars['Int']['output'];
  createdAt: Scalars['DateTime']['output'];
  createdBy: User;
  finishAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  lockAt?: Maybe<Scalars['DateTime']['output']>;
  name?: Maybe<Scalars['String']['output']>;
  openAt?: Maybe<Scalars['DateTime']['output']>;
  quiz: Quiz;
  state: QuizSessionState;
  userHasAccess: Scalars['Boolean']['output'];
  userSubmission?: Maybe<QuizSubmission>;
};

export enum QuizSessionState {
  Draft = 'DRAFT',
  Finished = 'FINISHED',
  Locked = 'LOCKED',
  Open = 'OPEN'
}

export type QuizSubmission = {
  __typename?: 'QuizSubmission';
  autoSubmitted: Scalars['Boolean']['output'];
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
  session?: Maybe<QuizSession>;
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

export type RegisterPushSubscriptionInput = {
  auth: Scalars['String']['input'];
  endpoint: Scalars['String']['input'];
  p256dh: Scalars['String']['input'];
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
  projectId?: InputMaybe<Scalars['ID']['input']>;
  sourceType?: InputMaybe<ScoreSourceType>;
  userId?: InputMaybe<Scalars['ID']['input']>;
};

export type ScoreSource = ContentAchievement | Event | ExternalChallenge | PluginChallenge | QuizAchievement | QuizChallenge | SimpleAchievement | SimpleChallenge | StreakAchievement;

export enum ScoreSourceType {
  Achievement = 'ACHIEVEMENT',
  Challenge = 'CHALLENGE',
  Event = 'EVENT',
  Manual = 'MANUAL'
}

export type SendPushNotificationInput = {
  allUsers?: InputMaybe<Scalars['Boolean']['input']>;
  body: Scalars['String']['input'];
  eventIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  projectIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  tag?: InputMaybe<Scalars['String']['input']>;
  teamIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  title: Scalars['String']['input'];
  type: NotificationType;
  url?: InputMaybe<Scalars['String']['input']>;
  userIds?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type SendPushNotificationResult = {
  __typename?: 'SendPushNotificationResult';
  failedDeliveries: Scalars['Int']['output'];
  success: Scalars['Boolean']['output'];
  successfulDeliveries: Scalars['Int']['output'];
  totalRecipients: Scalars['Int']['output'];
};

export type SetNotificationPreferenceInput = {
  enabled: Scalars['Boolean']['input'];
  notificationType: NotificationType;
};

export type SimpleAchievement = Achievement & {
  __typename?: 'SimpleAchievement';
  achievedAt?: Maybe<Scalars['DateTime']['output']>;
  awardableFrom?: Maybe<Scalars['DateTime']['output']>;
  celebratedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  descriptionCompleted: Scalars['String']['output'];
  descriptionPending: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  /** @deprecated Use imageCompletedObject instead */
  imageCompleted: Scalars['String']['output'];
  imageCompletedObject: Image;
  /** @deprecated Use imagePendingObject instead */
  imagePending: Scalars['String']['output'];
  imagePendingObject: Image;
  name: Scalars['String']['output'];
  notificationText: Scalars['String']['output'];
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
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
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
  awardableFrom?: Maybe<Scalars['DateTime']['output']>;
  celebratedAt?: Maybe<Scalars['DateTime']['output']>;
  challenge?: Maybe<Challenge>;
  descriptionCompleted: Scalars['String']['output'];
  descriptionPending: Scalars['String']['output'];
  event?: Maybe<Event>;
  hidden: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  /** @deprecated Use imageCompletedObject instead */
  imageCompleted: Scalars['String']['output'];
  imageCompletedObject: Image;
  /** @deprecated Use imagePendingObject instead */
  imagePending: Scalars['String']['output'];
  imagePendingObject: Image;
  name: Scalars['String']['output'];
  neededStreak: Scalars['Int']['output'];
  notificationText: Scalars['String']['output'];
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

export type SubmitFeedbackInput = {
  canContactMe: Scalars['Boolean']['input'];
  device: DeviceMetadata;
  message: Scalars['String']['input'];
  projectId?: InputMaybe<Scalars['ID']['input']>;
  tags?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type SubmitQuizAnswerInput = {
  jsonResponse?: InputMaybe<Scalars['JSON']['input']>;
  numberResponse?: InputMaybe<Scalars['Float']['input']>;
  questionId: Scalars['ID']['input'];
  selectedAnswerIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  submittedOrder?: InputMaybe<Array<Scalars['ID']['input']>>;
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
  averageAge?: Maybe<Scalars['Float']['output']>;
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  joinCode: Scalars['String']['output'];
  leaderboardExcluded: Scalars['Boolean']['output'];
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

export enum TeamScoreDistributionMode {
  Each = 'EACH',
  Split = 'SPLIT'
}

export type UpdateAchievementInput = {
  awardableFrom?: InputMaybe<Scalars['DateTime']['input']>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted?: InputMaybe<Scalars['String']['input']>;
  descriptionPending?: InputMaybe<Scalars['String']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden?: InputMaybe<Scalars['Boolean']['input']>;
  imageCompleted?: InputMaybe<Scalars['String']['input']>;
  imagePending?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  notificationText?: InputMaybe<Scalars['String']['input']>;
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
  pluginChallengeId?: InputMaybe<Scalars['String']['input']>;
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

export type UpdateContentAchievementInput = {
  awardableFrom?: InputMaybe<Scalars['DateTime']['input']>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted?: InputMaybe<Scalars['String']['input']>;
  descriptionPending?: InputMaybe<Scalars['String']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden?: InputMaybe<Scalars['Boolean']['input']>;
  imageCompleted?: InputMaybe<Scalars['String']['input']>;
  imagePending?: InputMaybe<Scalars['String']['input']>;
  items?: InputMaybe<Array<ContentItemInput>>;
  name?: InputMaybe<Scalars['String']['input']>;
  notificationText?: InputMaybe<Scalars['String']['input']>;
  points?: InputMaybe<Scalars['Int']['input']>;
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
  infoMessage?: InputMaybe<Scalars['String']['input']>;
  infoMessageEnd?: InputMaybe<Scalars['DateTime']['input']>;
  infoMessageStart?: InputMaybe<Scalars['DateTime']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  rules?: InputMaybe<Scalars['String']['input']>;
  startDate?: InputMaybe<Scalars['DateTime']['input']>;
};

export type UpdateQuizAnswerInput = {
  submittedOrder?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type UpdateQuizInput = {
  allowRetakes?: InputMaybe<Scalars['Boolean']['input']>;
  completionPoints?: InputMaybe<Scalars['Int']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  image?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  randomizeQuestions?: InputMaybe<Scalars['Boolean']['input']>;
  revealCorrectAnswers?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
};

export type UpdateQuizQuestionInput = {
  allowMultipleSelection?: InputMaybe<Scalars['Boolean']['input']>;
  maxValue?: InputMaybe<Scalars['Float']['input']>;
  minValue?: InputMaybe<Scalars['Float']['input']>;
  orderingItems?: InputMaybe<Array<CreateOrderingItemInput>>;
  points?: InputMaybe<Scalars['Int']['input']>;
  predefinedAnswers?: InputMaybe<Array<CreatePredefinedAnswerInput>>;
  questionOrder?: InputMaybe<Scalars['Int']['input']>;
  questionText?: InputMaybe<Scalars['String']['input']>;
  stepValue?: InputMaybe<Scalars['Float']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
};

export type UpdateQuizSessionInput = {
  finishAt?: InputMaybe<Scalars['DateTime']['input']>;
  lockAt?: InputMaybe<Scalars['DateTime']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  openAt?: InputMaybe<Scalars['DateTime']['input']>;
};

export type UpdateStreakAchievementInput = {
  awardableFrom?: InputMaybe<Scalars['DateTime']['input']>;
  challengeId?: InputMaybe<Scalars['ID']['input']>;
  descriptionCompleted?: InputMaybe<Scalars['String']['input']>;
  descriptionPending?: InputMaybe<Scalars['String']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  hidden?: InputMaybe<Scalars['Boolean']['input']>;
  imageCompleted?: InputMaybe<Scalars['String']['input']>;
  imagePending?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  neededStreak?: InputMaybe<Scalars['Int']['input']>;
  notificationText?: InputMaybe<Scalars['String']['input']>;
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
  leaderboardExcluded?: InputMaybe<Scalars['Boolean']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateWebhookInput = {
  active?: InputMaybe<Scalars['Boolean']['input']>;
  includeEventData?: InputMaybe<Scalars['Boolean']['input']>;
  includeUserData?: InputMaybe<Scalars['Boolean']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  secret?: InputMaybe<Scalars['String']['input']>;
  url?: InputMaybe<Scalars['String']['input']>;
};

export type User = {
  __typename?: 'User';
  age?: Maybe<Scalars['Int']['output']>;
  birthdate: Scalars['String']['output'];
  church: Church;
  churchId: Scalars['ID']['output'];
  consentStatus: ConsentStatus;
  createdAt: Scalars['DateTime']['output'];
  email: Scalars['String']['output'];
  events: Array<Event>;
  gender: Gender;
  id: Scalars['ID']['output'];
  /** @deprecated Use imageObject instead */
  image?: Maybe<Scalars['String']['output']>;
  imageObject?: Maybe<Image>;
  language: Scalars['String']['output'];
  membersId: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  personUuid?: Maybe<Scalars['ID']['output']>;
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

export type UserFeedback = {
  __typename?: 'UserFeedback';
  appVersion?: Maybe<Scalars['String']['output']>;
  canContactMe: Scalars['Boolean']['output'];
  contextUrl?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['DateTime']['output'];
  handledAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['ID']['output'];
  locale?: Maybe<Scalars['String']['output']>;
  message: Scalars['String']['output'];
  platform?: Maybe<Scalars['String']['output']>;
  projectId?: Maybe<Scalars['ID']['output']>;
  screenHeight?: Maybe<Scalars['Int']['output']>;
  screenWidth?: Maybe<Scalars['Int']['output']>;
  tags: Array<Scalars['String']['output']>;
  timezone?: Maybe<Scalars['String']['output']>;
  user: User;
  userAgent?: Maybe<Scalars['String']['output']>;
  userId: Scalars['ID']['output'];
};

export type UserFilter = {
  churchId?: InputMaybe<Scalars['ID']['input']>;
  eventId?: InputMaybe<Scalars['ID']['input']>;
  gender?: InputMaybe<Gender>;
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  maxAge?: InputMaybe<Scalars['Int']['input']>;
  minAge?: InputMaybe<Scalars['Int']['input']>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
  query?: InputMaybe<Scalars['String']['input']>;
  teamId?: InputMaybe<Scalars['ID']['input']>;
};

export type UserRole = {
  __typename?: 'UserRole';
  id: Scalars['ID']['output'];
  role: RoleType;
  scope?: Maybe<RoleScope>;
  user: User;
};

export type Webhook = {
  __typename?: 'Webhook';
  active: Scalars['Boolean']['output'];
  createdAt: Scalars['DateTime']['output'];
  eventType: WebhookEventType;
  hasSecret: Scalars['Boolean']['output'];
  id: Scalars['ID']['output'];
  includeEventData: Scalars['Boolean']['output'];
  includeUserData: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  project: Project;
  recentLogs: Array<WebhookLog>;
  updatedAt: Scalars['DateTime']['output'];
  url: Scalars['String']['output'];
};


export type WebhookRecentLogsArgs = {
  first?: InputMaybe<Scalars['Int']['input']>;
};

export enum WebhookEventType {
  ExternalContentEvent = 'EXTERNAL_CONTENT_EVENT',
  PointsAwarded = 'POINTS_AWARDED',
  QuizSessionFinished = 'QUIZ_SESSION_FINISHED',
  TeamNameChanged = 'TEAM_NAME_CHANGED'
}

export type WebhookLog = {
  __typename?: 'WebhookLog';
  createdAt: Scalars['DateTime']['output'];
  durationMs: Scalars['Int']['output'];
  errorMessage?: Maybe<Scalars['String']['output']>;
  eventType: WebhookEventType;
  id: Scalars['ID']['output'];
  requestPayload: Scalars['JSON']['output'];
  responseBody?: Maybe<Scalars['String']['output']>;
  responseStatusCode?: Maybe<Scalars['Int']['output']>;
};

export type ProjectRulesQueryVariables = Exact<{ [key: string]: never; }>;


export type ProjectRulesQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', rules?: { __typename?: 'MarkdownText', markdown: string, html: string } | null } };

export type PointHistoryQueryVariables = Exact<{
  last?: InputMaybe<Scalars['Int']['input']>;
}>;


export type PointHistoryQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', journal: { __typename?: 'ScoreJournalConnection', edges: Array<{ __typename?: 'ScoreJournalEdge', node: { __typename?: 'ScoreJournal', id: string, sourceType: ScoreSourceType, reason?: string | null, points: number, createdAt: any, source?:
            | { __typename: 'ContentAchievement', id: string, name: string }
            | { __typename: 'Event', id: string, name: string }
            | { __typename: 'ExternalChallenge', id: string, name: string }
            | { __typename: 'PluginChallenge', id: string, name: string }
            | { __typename: 'QuizAchievement', id: string, name: string }
            | { __typename: 'QuizChallenge', id: string, name: string }
            | { __typename: 'SimpleAchievement', id: string, name: string }
            | { __typename: 'SimpleChallenge', id: string, name: string }
            | { __typename: 'StreakAchievement', id: string, name: string }
           | null } }> } } };

export type GetMeQueryVariables = Exact<{ [key: string]: never; }>;


export type GetMeQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string, email: string, image?: string | null, membersId: string, language: string, gender: Gender, birthdate: string, age?: number | null, createdAt: any, church: { __typename?: 'Church', id: string, name: string, country: string, category: ChurchCategory }, roles: Array<{ __typename?: 'UserRole', id: string, role: RoleType, scope?: { __typename?: 'RoleScope', id: string, type: ScopeType, church?: { __typename?: 'Church', id: string } | null, team?: { __typename?: 'Team', id: string } | null, project?: { __typename?: 'Project', id: string } | null } | null }> } };

export type ColorSetFieldsFragment = { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string };

export type BrandingColorsFieldsFragment = { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } };

export type BrandingFieldsFragment = { __typename?: 'Branding', rounding: number, logoImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, bannerImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } };

export type ImageFieldsFragment = { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null };

export type LeaderboardEntryFieldsFragment = { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> };

export type LeaderboardEntryWithDescriptionFieldsFragment = { __typename?: 'LeaderboardEntry', id: string, name: string, description: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> };

export type PredefinedAnswerFieldsFragment = { __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null };

type QuizQuestionFields_FreeTextQuestion_Fragment = { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null };

type QuizQuestionFields_JsonQuestion_Fragment = { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null };

type QuizQuestionFields_NumberQuestion_Fragment = { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null };

type QuizQuestionFields_OrderingQuestion_Fragment = { __typename: 'OrderingQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, orderingItems: Array<{ __typename?: 'QuizOrderingItem', id: string, itemText: string }> };

type QuizQuestionFields_PredefinedQuestion_Fragment = { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> };

export type QuizQuestionFieldsFragment =
  | QuizQuestionFields_FreeTextQuestion_Fragment
  | QuizQuestionFields_JsonQuestion_Fragment
  | QuizQuestionFields_NumberQuestion_Fragment
  | QuizQuestionFields_OrderingQuestion_Fragment
  | QuizQuestionFields_PredefinedQuestion_Fragment
;

type QuizQuestionUserFields_FreeTextQuestion_Fragment = { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null };

type QuizQuestionUserFields_JsonQuestion_Fragment = { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null };

type QuizQuestionUserFields_NumberQuestion_Fragment = { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null };

type QuizQuestionUserFields_OrderingQuestion_Fragment = { __typename: 'OrderingQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, orderingItems: Array<{ __typename?: 'QuizOrderingItem', id: string, itemText: string }> };

type QuizQuestionUserFields_PredefinedQuestion_Fragment = { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> };

export type QuizQuestionUserFieldsFragment =
  | QuizQuestionUserFields_FreeTextQuestion_Fragment
  | QuizQuestionUserFields_JsonQuestion_Fragment
  | QuizQuestionUserFields_NumberQuestion_Fragment
  | QuizQuestionUserFields_OrderingQuestion_Fragment
  | QuizQuestionUserFields_PredefinedQuestion_Fragment
;

export type QuizSubmissionResultFieldsFragment = { __typename?: 'QuizSubmission', id: string, completedAt?: any | null, score?: number | null, maxScore?: number | null, scorePercentage?: number | null, pointsAwarded?: number | null };

export type DeleteAchievementMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteAchievementMutation = { __typename?: 'Mutation', deleteAchievement: boolean };

export type UpdateAchievementMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateAchievementInput;
}>;


export type UpdateAchievementMutation = { __typename?: 'Mutation', updateAchievement:
    | { __typename?: 'ContentAchievement', id: string }
    | { __typename?: 'QuizAchievement', id: string }
    | { __typename?: 'SimpleAchievement', id: string }
    | { __typename?: 'StreakAchievement', id: string }
   };

export type UpdateContentAchievementMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateContentAchievementInput;
}>;


export type UpdateContentAchievementMutation = { __typename?: 'Mutation', updateContentAchievement: { __typename?: 'ContentAchievement', id: string } };

export type UpdateStreakAchievementMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateStreakAchievementInput;
}>;


export type UpdateStreakAchievementMutation = { __typename?: 'Mutation', updateStreakAchievement: { __typename?: 'StreakAchievement', id: string } };

export type CreateContentAchievementMutationVariables = Exact<{
  input: CreateContentAchievementInput;
}>;


export type CreateContentAchievementMutation = { __typename?: 'Mutation', createContentAchievement: { __typename?: 'ContentAchievement', id: string } };

export type CreateQuizAchievementMutationVariables = Exact<{
  input: CreateQuizAchievementInput;
}>;


export type CreateQuizAchievementMutation = { __typename?: 'Mutation', createQuizAchievement: { __typename?: 'QuizAchievement', id: string } };

export type CreateStreakAchievementMutationVariables = Exact<{
  input: CreateStreakAchievementInput;
}>;


export type CreateStreakAchievementMutation = { __typename?: 'Mutation', createStreakAchievement: { __typename?: 'StreakAchievement', id: string } };

export type CreateSimpleAchievementMutationVariables = Exact<{
  input: CreateSimpleAchievementInput;
}>;


export type CreateSimpleAchievementMutation = { __typename?: 'Mutation', createSimpleAchievement: { __typename?: 'SimpleAchievement', id: string } };

export type ReorderAchievementsMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  achievementIds: Array<Scalars['ID']['input']> | Scalars['ID']['input'];
}>;


export type ReorderAchievementsMutation = { __typename?: 'Mutation', reorderAchievements: Array<
    | { __typename?: 'ContentAchievement', id: string }
    | { __typename?: 'QuizAchievement', id: string }
    | { __typename?: 'SimpleAchievement', id: string }
    | { __typename?: 'StreakAchievement', id: string }
  > };

export type MarkAchievementCelebratedMutationVariables = Exact<{
  achievementId: Scalars['ID']['input'];
}>;


export type MarkAchievementCelebratedMutation = { __typename?: 'Mutation', markAchievementCelebrated: boolean };

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
    | { __typename?: 'PluginChallenge', id: string }
    | { __typename?: 'QuizChallenge', id: string }
    | { __typename?: 'SimpleChallenge', id: string }
   };

export type CreateChallengeMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  eventId?: InputMaybe<Scalars['ID']['input']>;
  input: CreateChallengeInput;
}>;


export type CreateChallengeMutation = { __typename?: 'Mutation', createChallenge:
    | { __typename?: 'ExternalChallenge', id: string }
    | { __typename?: 'PluginChallenge', id: string }
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

export type CreateConsentMutationVariables = Exact<{
  key: Scalars['String']['input'];
  title: Scalars['String']['input'];
  shortText?: InputMaybe<Scalars['String']['input']>;
  body: Scalars['String']['input'];
  url?: InputMaybe<Scalars['String']['input']>;
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  isRemote?: InputMaybe<Scalars['Boolean']['input']>;
  managedBy?: InputMaybe<Scalars['String']['input']>;
}>;


export type CreateConsentMutation = { __typename?: 'Mutation', createConsent: { __typename?: 'Consent', id: string, key: string, version: number, title: string } };

export type UpdateConsentMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  title?: InputMaybe<Scalars['String']['input']>;
  shortText?: InputMaybe<Scalars['String']['input']>;
  body?: InputMaybe<Scalars['String']['input']>;
  url?: InputMaybe<Scalars['String']['input']>;
  publishedAt?: InputMaybe<Scalars['DateTime']['input']>;
  managedBy?: InputMaybe<Scalars['String']['input']>;
}>;


export type UpdateConsentMutation = { __typename?: 'Mutation', updateConsent: { __typename?: 'Consent', id: string, key: string, version: number, title: string } };

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

export type SubmitFeedbackMutationVariables = Exact<{
  input: SubmitFeedbackInput;
}>;


export type SubmitFeedbackMutation = { __typename?: 'Mutation', submitFeedback: { __typename?: 'UserFeedback', id: string } };

export type DeleteFeedbackMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteFeedbackMutation = { __typename?: 'Mutation', deleteFeedback: boolean };

export type ForwardFeedbackToDeskMutationVariables = Exact<{
  feedbackId: Scalars['ID']['input'];
  destination: ForwardDestination;
}>;


export type ForwardFeedbackToDeskMutation = { __typename?: 'Mutation', forwardFeedbackToDesk: boolean };

export type MarkFeedbackHandledMutationVariables = Exact<{
  feedbackId: Scalars['ID']['input'];
}>;


export type MarkFeedbackHandledMutation = { __typename?: 'Mutation', markFeedbackHandled: { __typename?: 'UserFeedback', id: string, handledAt?: any | null } };

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

export type RegisterPushSubscriptionMutationVariables = Exact<{
  input: RegisterPushSubscriptionInput;
}>;


export type RegisterPushSubscriptionMutation = { __typename?: 'Mutation', registerPushSubscription: { __typename?: 'PushSubscription', id: string, createdAt: any } };

export type UnregisterPushSubscriptionMutationVariables = Exact<{
  endpoint: Scalars['String']['input'];
}>;


export type UnregisterPushSubscriptionMutation = { __typename?: 'Mutation', unregisterPushSubscription: boolean };

export type CreateQuizMutationVariables = Exact<{
  input: CreateQuizInput;
}>;


export type CreateQuizMutation = { __typename?: 'Mutation', createQuiz: { __typename?: 'Quiz', id: string, name: string } };

export type UpdateQuizMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateQuizInput;
}>;


export type UpdateQuizMutation = { __typename?: 'Mutation', updateQuiz: { __typename?: 'Quiz', id: string, name: string } };

export type AddQuizQuestionMutationVariables = Exact<{
  quizId: Scalars['ID']['input'];
  input: CreateQuizQuestionInput;
}>;


export type AddQuizQuestionMutation = { __typename?: 'Mutation', addQuizQuestion:
    | { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
    | { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
    | { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
    | { __typename: 'OrderingQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, orderingItems: Array<{ __typename?: 'QuizOrderingItem', id: string, itemText: string }> }
    | { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> }
   };

export type UpdateQuizQuestionMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateQuizQuestionInput;
}>;


export type UpdateQuizQuestionMutation = { __typename?: 'Mutation', updateQuizQuestion:
    | { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
    | { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
    | { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
    | { __typename: 'OrderingQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, orderingItems: Array<{ __typename?: 'QuizOrderingItem', id: string, itemText: string }> }
    | { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> }
   };

export type DeleteQuizQuestionMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteQuizQuestionMutation = { __typename?: 'Mutation', deleteQuizQuestion: boolean };

export type StartQuizSessionMutationVariables = Exact<{
  sessionId: Scalars['ID']['input'];
}>;


export type StartQuizSessionMutation = { __typename?: 'Mutation', startQuizSession: { __typename?: 'QuizSubmission', id: string, startedAt: any, expiresAt?: any | null, isExpired: boolean, questionOrder: Array<string>, orderedQuestions: Array<
      | { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
      | { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
      | { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
      | { __typename: 'OrderingQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, orderingItems: Array<{ __typename?: 'QuizOrderingItem', id: string, itemText: string }> }
      | { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> }
    >, quiz: { __typename?: 'Quiz', id: string, name: string, timeoutSeconds?: number | null } } };

export type SubmitQuizAnswerMutationVariables = Exact<{
  submissionId: Scalars['ID']['input'];
  input: SubmitQuizAnswerInput;
}>;


export type SubmitQuizAnswerMutation = { __typename?: 'Mutation', submitQuizAnswer:
    | { __typename: 'FreeTextResponse', textResponse: string, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
        | { __typename?: 'FreeTextQuestion', id: string }
        | { __typename?: 'JsonQuestion', id: string }
        | { __typename?: 'NumberQuestion', id: string }
        | { __typename?: 'OrderingQuestion', id: string }
        | { __typename?: 'PredefinedQuestion', id: string }
       }
    | { __typename: 'JsonResponse', id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
        | { __typename?: 'FreeTextQuestion', id: string }
        | { __typename?: 'JsonQuestion', id: string }
        | { __typename?: 'NumberQuestion', id: string }
        | { __typename?: 'OrderingQuestion', id: string }
        | { __typename?: 'PredefinedQuestion', id: string }
       }
    | { __typename: 'NumberResponse', numberResponse: number, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
        | { __typename?: 'FreeTextQuestion', id: string }
        | { __typename?: 'JsonQuestion', id: string }
        | { __typename?: 'NumberQuestion', id: string }
        | { __typename?: 'OrderingQuestion', id: string }
        | { __typename?: 'PredefinedQuestion', id: string }
       }
    | { __typename: 'OrderingResponse', isCorrect?: boolean | null, submittedOrder: Array<string>, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
        | { __typename?: 'FreeTextQuestion', id: string }
        | { __typename?: 'JsonQuestion', id: string }
        | { __typename?: 'NumberQuestion', id: string }
        | { __typename?: 'OrderingQuestion', id: string }
        | { __typename?: 'PredefinedQuestion', id: string }
       }
    | { __typename: 'PredefinedResponse', isCorrect?: boolean | null, selectedAnswerIds: Array<string>, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, selectedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, isCorrect?: boolean | null }>, question:
        | { __typename?: 'FreeTextQuestion', id: string }
        | { __typename?: 'JsonQuestion', id: string }
        | { __typename?: 'NumberQuestion', id: string }
        | { __typename?: 'OrderingQuestion', id: string }
        | { __typename?: 'PredefinedQuestion', id: string }
       }
   };

export type UpdateQuizAnswerMutationVariables = Exact<{
  responseId: Scalars['ID']['input'];
  input: UpdateQuizAnswerInput;
}>;


export type UpdateQuizAnswerMutation = { __typename?: 'Mutation', updateQuizAnswer:
    | { __typename: 'FreeTextResponse', id: string }
    | { __typename: 'JsonResponse', id: string }
    | { __typename: 'NumberResponse', id: string }
    | { __typename: 'OrderingResponse', isCorrect?: boolean | null, submittedOrder: Array<string>, id: string }
    | { __typename: 'PredefinedResponse', id: string }
   };

export type FinalizeQuizMutationVariables = Exact<{
  submissionId: Scalars['ID']['input'];
}>;


export type FinalizeQuizMutation = { __typename?: 'Mutation', finalizeQuiz: { __typename?: 'QuizSubmission', id: string, completedAt?: any | null, score?: number | null, maxScore?: number | null, scorePercentage?: number | null, pointsAwarded?: number | null } };

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

export type AdminExternalContentsQueryVariables = Exact<{
  filter: ExternalContentFilter;
  sortBy?: InputMaybe<ExternalContentSortBy>;
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type AdminExternalContentsQuery = { __typename?: 'Query', externalContents: { __typename?: 'ExternalContentConnection', edges: Array<{ __typename?: 'ExternalContentEdge', node: { __typename?: 'ExternalContent', id: string, planId: string, taskId: string, contentId?: string | null, contentType: ExternalContentType, publishedAt?: any | null, source: string, title?: string | null, translations: Array<{ __typename?: 'ExternalContentTranslation', languageCode: string, title?: string | null }> } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type AdminProjectStreaksQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectStreaksQuery = { __typename?: 'Query', streaks: { __typename?: 'StreakConnection', edges: Array<{ __typename?: 'StreakEdge', node: { __typename?: 'Streak', id: string, name: string, description: string } }> } };

export type AdminProjectQuizzesQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectQuizzesQuery = { __typename?: 'Query', quizzes: { __typename?: 'QuizConnection', edges: Array<{ __typename?: 'QuizEdge', node: { __typename?: 'Quiz', id: string, name: string } }> } };

export type GetFirebaseTokenQueryVariables = Exact<{ [key: string]: never; }>;


export type GetFirebaseTokenQuery = { __typename?: 'Query', firebaseToken: { __typename?: 'FirebaseTokenResponse', token: string, expiresIn: number } };

export type ChallengePageQueryVariables = Exact<{
  challengeId: Scalars['ID']['input'];
}>;


export type ChallengePageQuery = { __typename?: 'Query', challenge:
    | { __typename: 'ExternalChallenge', url: string, id: string, name: string, description: any, requiresTeamMembership: boolean, requiresSuperTeamMembership: boolean, userEnrolledAt?: any | null, userCompletedAt?: any | null }
    | { __typename: 'PluginChallenge', pluginChallengeId: string, id: string, name: string, description: any, requiresTeamMembership: boolean, requiresSuperTeamMembership: boolean, userEnrolledAt?: any | null, userCompletedAt?: any | null }
    | { __typename: 'QuizChallenge', id: string, name: string, description: any, requiresTeamMembership: boolean, requiresSuperTeamMembership: boolean, userEnrolledAt?: any | null, userCompletedAt?: any | null, quiz: { __typename?: 'Quiz', id: string, name: string, description: string, timeoutSeconds?: number | null, randomizeQuestions: boolean, revealCorrectAnswers: boolean, allowRetakes: boolean, completionPoints: number, endTime?: any | null, userCanStart: boolean, userActiveSubmission?: { __typename?: 'QuizSubmission', id: string } | null, userActiveSession?: { __typename?: 'QuizSession', id: string, state: QuizSessionState } | null, userSubmissions: Array<{ __typename?: 'QuizSubmission', id: string, startedAt: any, completedAt?: any | null, expiresAt?: any | null, isExpired: boolean, score?: number | null, maxScore?: number | null, scorePercentage?: number | null, pointsAwarded?: number | null, orderedQuestions: Array<
            | { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
            | { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
            | { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null }
            | { __typename: 'OrderingQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, orderingItems: Array<{ __typename?: 'QuizOrderingItem', id: string, itemText: string }> }
            | { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> }
          >, responses: Array<
            | { __typename: 'FreeTextResponse', textResponse: string, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
                | { __typename?: 'FreeTextQuestion', id: string }
                | { __typename?: 'JsonQuestion', id: string }
                | { __typename?: 'NumberQuestion', id: string }
                | { __typename?: 'OrderingQuestion', id: string }
                | { __typename?: 'PredefinedQuestion', id: string }
               }
            | { __typename: 'JsonResponse', jsonResponse: any, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
                | { __typename?: 'FreeTextQuestion', id: string }
                | { __typename?: 'JsonQuestion', id: string }
                | { __typename?: 'NumberQuestion', id: string }
                | { __typename?: 'OrderingQuestion', id: string }
                | { __typename?: 'PredefinedQuestion', id: string }
               }
            | { __typename: 'NumberResponse', numberResponse: number, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
                | { __typename?: 'FreeTextQuestion', id: string }
                | { __typename?: 'JsonQuestion', id: string }
                | { __typename?: 'NumberQuestion', id: string }
                | { __typename?: 'OrderingQuestion', id: string }
                | { __typename?: 'PredefinedQuestion', id: string }
               }
            | { __typename: 'OrderingResponse', isCorrect?: boolean | null, submittedOrder: Array<string>, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, question:
                | { __typename?: 'FreeTextQuestion', id: string }
                | { __typename?: 'JsonQuestion', id: string }
                | { __typename?: 'NumberQuestion', id: string }
                | { __typename?: 'OrderingQuestion', id: string }
                | { __typename?: 'PredefinedQuestion', id: string }
               }
            | { __typename: 'PredefinedResponse', isCorrect?: boolean | null, id: string, answeredAt?: any | null, timeSpentSeconds?: number | null, selectedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }>, question:
                | { __typename?: 'FreeTextQuestion', id: string }
                | { __typename?: 'JsonQuestion', id: string }
                | { __typename?: 'NumberQuestion', id: string }
                | { __typename?: 'OrderingQuestion', id: string }
                | { __typename?: 'PredefinedQuestion', id: string }
               }
          > }> } }
    | { __typename: 'SimpleChallenge', allowSelfCompletion: boolean, id: string, name: string, description: any, requiresTeamMembership: boolean, requiresSuperTeamMembership: boolean, userEnrolledAt?: any | null, userCompletedAt?: any | null }
   };

export type ChallengesPageQueryVariables = Exact<{ [key: string]: never; }>;


export type ChallengesPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', myTeam?: { __typename?: 'Team', joinCode: string } | null, challenges: Array<
      | { __typename: 'ExternalChallenge', url: string, id: string, name: string, description: any, buttonText: string, publishedAt?: any | null, endTime?: any | null, visibleAt?: any | null, userCompletedAt?: any | null, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
      | { __typename: 'PluginChallenge', id: string, name: string, description: any, buttonText?: string | null, publishedAt?: any | null, endTime?: any | null, visibleAt?: any | null, userCompletedAt?: any | null, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
      | { __typename: 'QuizChallenge', id: string, name: string, description: any, buttonText: string, publishedAt?: any | null, endTime?: any | null, visibleAt?: any | null, userCompletedAt?: any | null, quiz: { __typename?: 'Quiz', userCanStart: boolean, userActiveSession?: { __typename?: 'QuizSession', id: string } | null }, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
      | { __typename: 'SimpleChallenge', allowSelfCompletion: boolean, id: string, name: string, description: any, buttonText: string, publishedAt?: any | null, endTime?: any | null, visibleAt?: any | null, userCompletedAt?: any | null, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
    > } };

export type ProfilePageQueryVariables = Exact<{
  ageFilter?: InputMaybe<LeaderboardFilter>;
}>;


export type ProfilePageQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string, consentStatus: { __typename?: 'ConsentStatus', pendingConsents: Array<{ __typename: 'Consent', id: string, key: string, version: number, title: string, shortText: string, url?: string | null, managementType: ConsentManagementType, managedBy?: string | null, body: { __typename?: 'MarkdownText', html: string } }> } }, myCurrentProject: { __typename?: 'Project', id: string, name: string, infoMessageStart?: any | null, infoMessageEnd?: any | null, infoMessage?: { __typename?: 'MarkdownText', markdown: string, html: string } | null, branding: { __typename?: 'Branding', rounding: number, logoImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, bannerImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } }, achievements: Array<
      | { __typename: 'ContentAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, hidden: boolean, achievedAt?: any | null, celebratedAt?: any | null, points: number, nextItem?: { __typename?: 'ContentItem', id: string, sortOrder: number, externalContent: { __typename?: 'ExternalContent', id: string, title?: string | null, url?: string | null } } | null, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
      | { __typename: 'QuizAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, hidden: boolean, achievedAt?: any | null, celebratedAt?: any | null, points: number, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
      | { __typename: 'SimpleAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, hidden: boolean, achievedAt?: any | null, celebratedAt?: any | null, points: number, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
      | { __typename: 'StreakAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, hidden: boolean, achievedAt?: any | null, celebratedAt?: any | null, points: number, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
    >, leaderboard: { __typename?: 'LeaderboardConnection', me?: { __typename?: 'LeaderboardEntry', score: number, rank?: number | null } | null } } };

export type ConsentsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type ConsentsPageQuery = { __typename?: 'Query', me: { __typename?: 'User', consentStatus: { __typename?: 'ConsentStatus', pendingConsents: Array<{ __typename: 'Consent', id: string, key: string, version: number, title: string, shortText: string, publishedAt?: any | null, managedBy?: string | null, managementType: ConsentManagementType, url?: string | null, body: { __typename?: 'MarkdownText', html: string } }>, acceptedConsents: Array<{ __typename: 'UserConsent', id: string, action: ConsentAction, actionDate: any, consent: { __typename?: 'Consent', id: string, title: string, shortText: string, managedBy?: string | null, managementType: ConsentManagementType, url?: string | null, body: { __typename?: 'MarkdownText', html: string } } }>, rejectedConsents: Array<{ __typename: 'UserConsent', id: string, action: ConsentAction, actionDate: any, consent: { __typename?: 'Consent', id: string, title: string, shortText: string, managedBy?: string | null, managementType: ConsentManagementType, url?: string | null, body: { __typename?: 'MarkdownText', html: string } } }> } } };

export type StandingsGlobalPageQueryVariables = Exact<{
  entityType: LeaderboardEntityType;
  filter?: InputMaybe<LeaderboardFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
}>;


export type StandingsGlobalPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', id: string, leaderboard: { __typename?: 'LeaderboardConnection', edges: Array<{ __typename?: 'LeaderboardEdge', node: { __typename?: 'LeaderboardEntry', id: string, name: string, description: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } }>, me?: { __typename?: 'LeaderboardEntry', id: string, name: string, description: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } | null } } };

export type StandingsLocalPageQueryVariables = Exact<{
  filter?: InputMaybe<LeaderboardFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
}>;


export type StandingsLocalPageQuery = { __typename?: 'Query', me: { __typename?: 'User', church: { __typename?: 'Church', id: string, name: string } }, myCurrentProject: { __typename?: 'Project', id: string, personLeaderboard: { __typename?: 'LeaderboardConnection', totalCount: number, edges: Array<{ __typename?: 'LeaderboardEdge', node: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } }>, me?: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } | null }, unitLeaderboard: { __typename?: 'LeaderboardConnection', totalCount: number, edges: Array<{ __typename?: 'LeaderboardEdge', node: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } }>, me?: { __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> } | null } } };

export type StandingsUnitPageQueryVariables = Exact<{ [key: string]: never; }>;


export type StandingsUnitPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', id: string, myTeam?: { __typename?: 'Team', id: string, name: string, memberLeaderboard: Array<{ __typename?: 'LeaderboardEntry', id: string, name: string, score: number, rank?: number | null, tags: Array<LeaderboardEntryTag> }> } | null } };

export type VapidPublicKeyQueryVariables = Exact<{ [key: string]: never; }>;


export type VapidPublicKeyQuery = { __typename?: 'Query', vapidPublicKey: string };

export type CurrentProjectQueryVariables = Exact<{ [key: string]: never; }>;


export type CurrentProjectQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', branding: { __typename?: 'Branding', rounding: number, logoImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, bannerImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } };

export type AdminChurchPageQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type AdminChurchPageQuery = { __typename?: 'Query', church: { __typename?: 'Church', id: string, name: string, country: string, category: ChurchCategory } };

export type UpdateChurchMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateChurchInput;
}>;


export type UpdateChurchMutation = { __typename?: 'Mutation', updateChurch: { __typename?: 'Church', id: string, name: string, country: string, category: ChurchCategory } };

export type AdminConsentPageQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type AdminConsentPageQuery = { __typename?: 'Query', consent: { __typename?: 'Consent', id: string, key: string, version: number, title: string, shortText: string, url?: string | null, publishedAt?: any | null, managementType: ConsentManagementType, managedBy?: string | null, body: { __typename?: 'MarkdownText', markdown: string, html: string } } };

export type AdminConsentsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminConsentsPageQuery = { __typename?: 'Query', consents: Array<{ __typename?: 'Consent', id: string, key: string, version: number, title: string, shortText: string, publishedAt?: any | null, managementType: ConsentManagementType, managedBy?: string | null }> };

export type AdminFeedbackPageQueryVariables = Exact<{
  filter?: InputMaybe<FeedbackFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
}>;


export type AdminFeedbackPageQuery = { __typename?: 'Query', feedback: { __typename?: 'FeedbackConnection', totalCount: number, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null }, edges: Array<{ __typename?: 'FeedbackEdge', cursor: string, node: { __typename?: 'UserFeedback', id: string, message: string, canContactMe: boolean, userAgent?: string | null, platform?: string | null, screenWidth?: number | null, screenHeight?: number | null, appVersion?: string | null, locale?: string | null, projectId?: string | null, timezone?: string | null, contextUrl?: string | null, tags: Array<string>, createdAt: any, handledAt?: any | null, user: { __typename?: 'User', id: string, name: string, email: string } } }> } };

export type UpdateFeedbackTagsMutationVariables = Exact<{
  feedbackId: Scalars['ID']['input'];
  tags: Array<Scalars['String']['input']> | Scalars['String']['input'];
}>;


export type UpdateFeedbackTagsMutation = { __typename?: 'Mutation', updateFeedbackTags: { __typename?: 'UserFeedback', id: string, tags: Array<string> } };

export type AdminHomePageQueryVariables = Exact<{
  now: Scalars['DateTime']['input'];
}>;


export type AdminHomePageQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, name: string }, adminDashboardStats: { __typename?: 'AdminDashboardStats', totalUsers: number, totalPointsAwarded: number, newUsersLast7Days: number }, feedback: { __typename?: 'FeedbackConnection', edges: Array<{ __typename?: 'FeedbackEdge', node: { __typename?: 'UserFeedback', id: string, message: string, createdAt: any, user: { __typename?: 'User', id: string, name: string } } }> }, projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, description: string, endDate: any, startDate: any, branding: { __typename?: 'Branding', logo?: string | null, rounding: number, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string }, dark: { __typename?: 'ColorSet', accent: string } } } } }> } };

export type ChurchAdminsPageQueryVariables = Exact<{
  churchId: Scalars['ID']['input'];
}>;


export type ChurchAdminsPageQuery = { __typename?: 'Query', usersWithRole: Array<{ __typename?: 'User', id: string, name: string, email: string }>, users: { __typename?: 'UserConnection', edges: Array<{ __typename?: 'UserEdge', node: { __typename?: 'User', id: string, name: string, email: string } }> } };

export type MyChurchUnitsPageQueryVariables = Exact<{
  filter?: InputMaybe<UserFilter>;
}>;


export type MyChurchUnitsPageQuery = { __typename?: 'Query', users: { __typename?: 'UserConnection', edges: Array<{ __typename?: 'UserEdge', node: { __typename?: 'User', id: string, name: string, age?: number | null, gender: Gender, teams: Array<{ __typename?: 'Team', id: string, name: string }> } }> }, myCurrentProject: { __typename?: 'Project', id: string, name: string, myChurchTeams: Array<{ __typename?: 'Team', id: string, name: string, leaderboardExcluded: boolean, averageAge?: number | null, joinCode: string, members: Array<{ __typename?: 'TeamMember', id: string, name: string, isTeamLead: boolean, user: { __typename?: 'User', id: string, age?: number | null, gender: Gender } }> }> } };

export type AdminProjectAchievementPageQueryVariables = Exact<{
  achievementId: Scalars['ID']['input'];
}>;


export type AdminProjectAchievementPageQuery = { __typename?: 'Query', achievement:
    | { __typename: 'ContentAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, notificationText: string, achievedAt?: any | null, points: number, hidden: boolean, awardableFrom?: any | null, items: Array<{ __typename?: 'ContentItem', id: string, sortOrder: number, externalContent: { __typename?: 'ExternalContent', id: string, planId: string, taskId: string, contentId?: string | null, contentType: ExternalContentType, publishedAt?: any | null, source: string, syncedAt: any, createdAt: any, updatedAt: any, title?: string | null, translations: Array<{ __typename?: 'ExternalContentTranslation', languageCode: string, title?: string | null }> } }>, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
    | { __typename: 'QuizAchievement', minScorePercentage?: number | null, requireCompletion: boolean, id: string, name: string, descriptionPending: string, descriptionCompleted: string, notificationText: string, achievedAt?: any | null, points: number, hidden: boolean, awardableFrom?: any | null, quiz: { __typename?: 'Quiz', id: string, name: string }, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
    | { __typename: 'SimpleAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, notificationText: string, achievedAt?: any | null, points: number, hidden: boolean, awardableFrom?: any | null, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
    | { __typename: 'StreakAchievement', neededStreak: number, id: string, name: string, descriptionPending: string, descriptionCompleted: string, notificationText: string, achievedAt?: any | null, points: number, hidden: boolean, awardableFrom?: any | null, streak: { __typename?: 'Streak', id: string, name: string, description: string }, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
   };

export type AdminProjectAchievementsNewPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectAchievementsNewPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } };

export type AdminProjectChallengePageQueryVariables = Exact<{
  challengeId: Scalars['ID']['input'];
}>;


export type AdminProjectChallengePageQuery = { __typename?: 'Query', challenge:
    | { __typename: 'ExternalChallenge', url: string, id: string, name: string, description: any, image?: string | null, buttonText: string, visibleAt?: any | null, startedAt?: any | null, endTime?: any | null, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
    | { __typename: 'PluginChallenge', pluginChallengeId: string, id: string, name: string, description: any, image?: string | null, buttonText?: string | null, visibleAt?: any | null, startedAt?: any | null, endTime?: any | null, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
    | { __typename: 'QuizChallenge', id: string, name: string, description: any, image?: string | null, buttonText: string, visibleAt?: any | null, startedAt?: any | null, endTime?: any | null, quiz: { __typename?: 'Quiz', id: string, name: string, description: string, image?: string | null, timeoutSeconds?: number | null, randomizeQuestions: boolean, revealCorrectAnswers: boolean, allowRetakes: boolean, completionPoints: number, questions: Array<
          | { __typename: 'FreeTextQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
          | { __typename: 'JsonQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
          | { __typename: 'NumberQuestion', minValue?: number | null, maxValue?: number | null, stepValue?: number | null, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null }
          | { __typename: 'OrderingQuestion', id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, orderingItems: Array<{ __typename?: 'QuizOrderingItem', id: string, itemText: string }> }
          | { __typename: 'PredefinedQuestion', allowMultipleSelection: boolean, id: string, questionText: string, questionOrder: number, timeoutSeconds?: number | null, points?: number | null, predefinedAnswers: Array<{ __typename?: 'QuizPredefinedAnswer', id: string, answerText: string, answerOrder: number, isCorrect?: boolean | null }> }
        > }, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
    | { __typename: 'SimpleChallenge', allowSelfCompletion: boolean, id: string, name: string, description: any, image?: string | null, buttonText: string, visibleAt?: any | null, startedAt?: any | null, endTime?: any | null, project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } } }
   };

export type AdminProjectChallengeNewPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectChallengeNewPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, branding: { __typename?: 'Branding', colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } } }, events: { __typename?: 'EventConnection', edges: Array<{ __typename?: 'EventEdge', node: { __typename?: 'Event', id: string, name: string } }> } };

export type AdminProjectEditPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectEditPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, description: string, startDate: any, endDate: any, archivedAt?: boolean | null, infoMessageStart?: any | null, infoMessageEnd?: any | null, branding: { __typename?: 'Branding', rounding: number, logoImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, bannerImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string }, dark: { __typename?: 'ColorSet', accent: string, accentContrast: string, onAccent: string, backgroundDefault: string, backgroundRaised: string, backgroundIndent: string, textDefault: string, textMuted: string, textHint: string, shadowDefault: string, shadowBlank: string, borderDefault: string } } }, rules?: { __typename?: 'MarkdownText', markdown: string, html: string } | null, infoMessage?: { __typename?: 'MarkdownText', markdown: string, html: string } | null } };

export type AdminProjectEventPageQueryVariables = Exact<{
  eventId: Scalars['ID']['input'];
}>;


export type AdminProjectEventPageQuery = { __typename?: 'Query', event: { __typename?: 'Event', id: string, name: string, description: string, startDate: any, endDate: any, parentProject: { __typename?: 'Project', id: string, name: string } } };

export type AdminProjectPageQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type AdminProjectPageQuery = { __typename?: 'Query', project: { __typename?: 'Project', id: string, name: string, description: string, startDate: any, endDate: any, branding: { __typename?: 'Branding', rounding: number, logoImage?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string }, dark: { __typename?: 'ColorSet', accent: string } } } }, achievements: { __typename?: 'AchievementConnection', edges: Array<{ __typename?: 'AchievementEdge', node:
        | { __typename?: 'ContentAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, points: number, hidden: boolean, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
        | { __typename?: 'QuizAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, points: number, hidden: boolean, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
        | { __typename?: 'SimpleAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, points: number, hidden: boolean, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
        | { __typename?: 'StreakAchievement', id: string, name: string, descriptionPending: string, descriptionCompleted: string, points: number, hidden: boolean, imagePendingObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null }, imageCompletedObject: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } }
       }> }, challenges: { __typename?: 'ChallengeConnection', edges: Array<{ __typename?: 'ChallengeEdge', node:
        | { __typename: 'ExternalChallenge', id: string, name: string, description: any, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
        | { __typename: 'PluginChallenge', id: string, name: string, description: any, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
        | { __typename: 'QuizChallenge', id: string, name: string, description: any, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
        | { __typename: 'SimpleChallenge', id: string, name: string, description: any, imageObject?: { __typename?: 'Image', url: string, width?: number | null, height?: number | null, blurhash?: string | null } | null }
       }> } };

export type AdminProjectStreakPageQueryVariables = Exact<{
  streakId: Scalars['ID']['input'];
}>;


export type AdminProjectStreakPageQuery = { __typename?: 'Query', streak: { __typename?: 'Streak', id: string, name: string, description: string, status: number, relevantDays: Array<{ __typename?: 'DateRange', start: any, end: any }>, project: { __typename?: 'Project', id: string, name: string } } };

export type AdminProjectsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminProjectsPageQuery = { __typename?: 'Query', projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string, description: string, endDate: any, startDate: any, branding: { __typename?: 'Branding', logo?: string | null, colors: { __typename?: 'Colors', light: { __typename?: 'ColorSet', accent: string }, dark: { __typename?: 'ColorSet', accent: string } } } } }> } };

export type AdminScoresPageQueryVariables = Exact<{
  filter?: InputMaybe<ScoreJournalFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
}>;


export type AdminScoresPageQuery = { __typename?: 'Query', adminScoreJournal: { __typename?: 'ScoreJournalConnection', totalCount: number, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null }, edges: Array<{ __typename?: 'ScoreJournalEdge', cursor: string, node: { __typename?: 'ScoreJournal', id: string, points: number, sourceType: ScoreSourceType, reason?: string | null, createdAt: any, user: { __typename?: 'User', id: string, name: string }, project: { __typename?: 'Project', id: string, name: string }, awardedBy?: { __typename?: 'User', id: string, name: string } | null } }> } };

export type DeleteScoreJournalEntryMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteScoreJournalEntryMutation = { __typename?: 'Mutation', deleteScoreJournalEntry: boolean };

export type AdminScoresNewPageQueryVariables = Exact<{ [key: string]: never; }>;


export type AdminScoresNewPageQuery = { __typename?: 'Query', projects: { __typename?: 'ProjectConnection', edges: Array<{ __typename?: 'ProjectEdge', node: { __typename?: 'Project', id: string, name: string } }> } };

export type CreateScoreAdjustmentMutationVariables = Exact<{
  input: CreateScoreAdjustmentInput;
}>;


export type CreateScoreAdjustmentMutation = { __typename?: 'Mutation', createScoreAdjustment: { __typename?: 'ScoreJournal', id: string, points: number, reason?: string | null } };

export type AdminTeamPageQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type AdminTeamPageQuery = { __typename?: 'Query', team: { __typename?: 'Team', id: string, name: string, description: string, joinCode: string, leaderboardExcluded: boolean, averageAge?: number | null, members: Array<{ __typename?: 'TeamMember', id: string, name: string, isTeamLead: boolean, joinedAt: string, user: { __typename?: 'User', id: string, email: string, image?: string | null }, church: { __typename?: 'Church', id: string, name: string } }>, parentProject: { __typename?: 'Project', id: string, name: string }, superTeam?: { __typename?: 'SuperTeam', id: string, name: string } | null } };

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


export type AdminUserPageQuery = { __typename?: 'Query', user: { __typename?: 'User', id: string, personUuid?: string | null, createdAt: any, name: string, email: string, membersId: string, gender: Gender, birthdate: string, age?: number | null, image?: string | null, church: { __typename?: 'Church', id: string, name: string }, roles: Array<{ __typename?: 'UserRole', id: string, role: RoleType, scope?: { __typename?: 'RoleScope', id: string, type: ScopeType } | null }>, consentStatus: { __typename?: 'ConsentStatus', acceptedConsents: Array<{ __typename?: 'UserConsent', id: string, action: ConsentAction, actionDate: any, consent: { __typename?: 'Consent', id: string, key: string, title: string, version: number, managementType: ConsentManagementType } }>, rejectedConsents: Array<{ __typename?: 'UserConsent', id: string, action: ConsentAction, actionDate: any, consent: { __typename?: 'Consent', id: string, key: string, title: string, version: number } }>, pendingConsents: Array<{ __typename?: 'Consent', id: string, key: string, title: string, version: number }> } }, adminScoreJournal: { __typename?: 'ScoreJournalConnection', totalCount: number, edges: Array<{ __typename?: 'ScoreJournalEdge', node: { __typename?: 'ScoreJournal', id: string, points: number, sourceType: ScoreSourceType, reason?: string | null, createdAt: any, project: { __typename?: 'Project', id: string, name: string }, awardedBy?: { __typename?: 'User', id: string, name: string } | null } }> }, feedback: { __typename?: 'FeedbackConnection', totalCount: number, edges: Array<{ __typename?: 'FeedbackEdge', node: { __typename?: 'UserFeedback', id: string, message: string, canContactMe: boolean, userAgent?: string | null, platform?: string | null, screenWidth?: number | null, screenHeight?: number | null, appVersion?: string | null, createdAt: any } }> } };

export type AdminSetUserConsentMutationVariables = Exact<{
  userId: Scalars['ID']['input'];
  consentId: Scalars['ID']['input'];
  action: ConsentAction;
}>;


export type AdminSetUserConsentMutation = { __typename?: 'Mutation', adminSetUserConsent: { __typename?: 'UserConsentHistoryEntry', id: string, action: ConsentAction } };

export type AdminUsersPageQueryVariables = Exact<{
  filter?: InputMaybe<UserFilter>;
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
}>;


export type AdminUsersPageQuery = { __typename?: 'Query', users: { __typename?: 'UserConnection', totalCount: number, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null }, edges: Array<{ __typename?: 'UserEdge', cursor: string, node: { __typename?: 'User', id: string, name: string, email: string, image?: string | null, church: { __typename?: 'Church', name: string }, roles: Array<{ __typename?: 'UserRole', id: string, role: RoleType }> } }> } };

export type StandingsPageQueryVariables = Exact<{ [key: string]: never; }>;


export type StandingsPageQuery = { __typename?: 'Query', myCurrentProject: { __typename?: 'Project', myTeam?: { __typename?: 'Team', id: string } | null } };

export const ImageFieldsFragmentDoc = gql`
    fragment ImageFields on Image {
  url
  width
  height
  blurhash
}
    `;
export const ColorSetFieldsFragmentDoc = gql`
    fragment ColorSetFields on ColorSet {
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
    `;
export const BrandingColorsFieldsFragmentDoc = gql`
    fragment BrandingColorsFields on Colors {
  light {
    ...ColorSetFields
  }
  dark {
    ...ColorSetFields
  }
}
    ${ColorSetFieldsFragmentDoc}`;
export const BrandingFieldsFragmentDoc = gql`
    fragment BrandingFields on Branding {
  logoImage {
    ...ImageFields
  }
  bannerImage {
    ...ImageFields
  }
  rounding
  colors {
    ...BrandingColorsFields
  }
}
    ${ImageFieldsFragmentDoc}
${BrandingColorsFieldsFragmentDoc}`;
export const LeaderboardEntryFieldsFragmentDoc = gql`
    fragment LeaderboardEntryFields on LeaderboardEntry {
  id
  name
  score
  rank
  tags
}
    `;
export const LeaderboardEntryWithDescriptionFieldsFragmentDoc = gql`
    fragment LeaderboardEntryWithDescriptionFields on LeaderboardEntry {
  id
  name
  description
  score
  rank
  tags
}
    `;
export const PredefinedAnswerFieldsFragmentDoc = gql`
    fragment PredefinedAnswerFields on QuizPredefinedAnswer {
  id
  answerText
  answerOrder
  isCorrect
}
    `;
export const QuizQuestionFieldsFragmentDoc = gql`
    fragment QuizQuestionFields on QuizQuestion {
  __typename
  id
  questionText
  questionOrder
  timeoutSeconds
  points
  ... on PredefinedQuestion {
    allowMultipleSelection
    predefinedAnswers {
      ...PredefinedAnswerFields
    }
  }
  ... on NumberQuestion {
    minValue
    maxValue
    stepValue
  }
  ... on OrderingQuestion {
    orderingItems {
      id
      itemText
    }
  }
}
    ${PredefinedAnswerFieldsFragmentDoc}`;
export const QuizQuestionUserFieldsFragmentDoc = gql`
    fragment QuizQuestionUserFields on QuizQuestion {
  __typename
  id
  questionText
  questionOrder
  timeoutSeconds
  ... on PredefinedQuestion {
    allowMultipleSelection
    predefinedAnswers {
      ...PredefinedAnswerFields
    }
  }
  ... on NumberQuestion {
    minValue
    maxValue
    stepValue
  }
  ... on OrderingQuestion {
    orderingItems {
      id
      itemText
    }
  }
}
    ${PredefinedAnswerFieldsFragmentDoc}`;
export const QuizSubmissionResultFieldsFragmentDoc = gql`
    fragment QuizSubmissionResultFields on QuizSubmission {
  id
  completedAt
  score
  maxScore
  scorePercentage
  pointsAwarded
}
    `;
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
            ... on ContentAchievement {
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
    language
    church {
      id
      name
      country
      category
    }
    gender
    birthdate
    age
    createdAt
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
export const UpdateContentAchievementDocument = gql`
    mutation UpdateContentAchievement($id: ID!, $input: UpdateContentAchievementInput!) {
  updateContentAchievement(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateContentAchievementMutation() {
  return Urql.useMutation<UpdateContentAchievementMutation, UpdateContentAchievementMutationVariables>(UpdateContentAchievementDocument);
};
export const UpdateStreakAchievementDocument = gql`
    mutation UpdateStreakAchievement($id: ID!, $input: UpdateStreakAchievementInput!) {
  updateStreakAchievement(id: $id, input: $input) {
    id
  }
}
    `;

export function useUpdateStreakAchievementMutation() {
  return Urql.useMutation<UpdateStreakAchievementMutation, UpdateStreakAchievementMutationVariables>(UpdateStreakAchievementDocument);
};
export const CreateContentAchievementDocument = gql`
    mutation CreateContentAchievement($input: CreateContentAchievementInput!) {
  createContentAchievement(input: $input) {
    id
  }
}
    `;

export function useCreateContentAchievementMutation() {
  return Urql.useMutation<CreateContentAchievementMutation, CreateContentAchievementMutationVariables>(CreateContentAchievementDocument);
};
export const CreateQuizAchievementDocument = gql`
    mutation CreateQuizAchievement($input: CreateQuizAchievementInput!) {
  createQuizAchievement(input: $input) {
    id
  }
}
    `;

export function useCreateQuizAchievementMutation() {
  return Urql.useMutation<CreateQuizAchievementMutation, CreateQuizAchievementMutationVariables>(CreateQuizAchievementDocument);
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
export const ReorderAchievementsDocument = gql`
    mutation ReorderAchievements($projectId: ID!, $achievementIds: [ID!]!) {
  reorderAchievements(projectId: $projectId, achievementIds: $achievementIds) {
    id
  }
}
    `;

export function useReorderAchievementsMutation() {
  return Urql.useMutation<ReorderAchievementsMutation, ReorderAchievementsMutationVariables>(ReorderAchievementsDocument);
};
export const MarkAchievementCelebratedDocument = gql`
    mutation MarkAchievementCelebrated($achievementId: ID!) {
  markAchievementCelebrated(achievementId: $achievementId)
}
    `;

export function useMarkAchievementCelebratedMutation() {
  return Urql.useMutation<MarkAchievementCelebratedMutation, MarkAchievementCelebratedMutationVariables>(MarkAchievementCelebratedDocument);
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
    mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
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
export const CreateConsentDocument = gql`
    mutation CreateConsent($key: String!, $title: String!, $shortText: String, $body: String!, $url: String, $publishedAt: DateTime, $isRemote: Boolean, $managedBy: String) {
  createConsent(
    key: $key
    title: $title
    shortText: $shortText
    body: $body
    url: $url
    publishedAt: $publishedAt
    isRemote: $isRemote
    managedBy: $managedBy
  ) {
    id
    key
    version
    title
  }
}
    `;

export function useCreateConsentMutation() {
  return Urql.useMutation<CreateConsentMutation, CreateConsentMutationVariables>(CreateConsentDocument);
};
export const UpdateConsentDocument = gql`
    mutation UpdateConsent($id: ID!, $title: String, $shortText: String, $body: String, $url: String, $publishedAt: DateTime, $managedBy: String) {
  updateConsent(
    id: $id
    title: $title
    shortText: $shortText
    body: $body
    url: $url
    publishedAt: $publishedAt
    managedBy: $managedBy
  ) {
    id
    key
    version
    title
  }
}
    `;

export function useUpdateConsentMutation() {
  return Urql.useMutation<UpdateConsentMutation, UpdateConsentMutationVariables>(UpdateConsentDocument);
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
export const SubmitFeedbackDocument = gql`
    mutation SubmitFeedback($input: SubmitFeedbackInput!) {
  submitFeedback(input: $input) {
    id
  }
}
    `;

export function useSubmitFeedbackMutation() {
  return Urql.useMutation<SubmitFeedbackMutation, SubmitFeedbackMutationVariables>(SubmitFeedbackDocument);
};
export const DeleteFeedbackDocument = gql`
    mutation DeleteFeedback($id: ID!) {
  deleteFeedback(id: $id)
}
    `;

export function useDeleteFeedbackMutation() {
  return Urql.useMutation<DeleteFeedbackMutation, DeleteFeedbackMutationVariables>(DeleteFeedbackDocument);
};
export const ForwardFeedbackToDeskDocument = gql`
    mutation ForwardFeedbackToDesk($feedbackId: ID!, $destination: ForwardDestination!) {
  forwardFeedbackToDesk(feedbackId: $feedbackId, destination: $destination)
}
    `;

export function useForwardFeedbackToDeskMutation() {
  return Urql.useMutation<ForwardFeedbackToDeskMutation, ForwardFeedbackToDeskMutationVariables>(ForwardFeedbackToDeskDocument);
};
export const MarkFeedbackHandledDocument = gql`
    mutation MarkFeedbackHandled($feedbackId: ID!) {
  markFeedbackHandled(feedbackId: $feedbackId) {
    id
    handledAt
  }
}
    `;

export function useMarkFeedbackHandledMutation() {
  return Urql.useMutation<MarkFeedbackHandledMutation, MarkFeedbackHandledMutationVariables>(MarkFeedbackHandledDocument);
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
export const RegisterPushSubscriptionDocument = gql`
    mutation RegisterPushSubscription($input: RegisterPushSubscriptionInput!) {
  registerPushSubscription(input: $input) {
    id
    createdAt
  }
}
    `;

export function useRegisterPushSubscriptionMutation() {
  return Urql.useMutation<RegisterPushSubscriptionMutation, RegisterPushSubscriptionMutationVariables>(RegisterPushSubscriptionDocument);
};
export const UnregisterPushSubscriptionDocument = gql`
    mutation UnregisterPushSubscription($endpoint: String!) {
  unregisterPushSubscription(endpoint: $endpoint)
}
    `;

export function useUnregisterPushSubscriptionMutation() {
  return Urql.useMutation<UnregisterPushSubscriptionMutation, UnregisterPushSubscriptionMutationVariables>(UnregisterPushSubscriptionDocument);
};
export const CreateQuizDocument = gql`
    mutation CreateQuiz($input: CreateQuizInput!) {
  createQuiz(input: $input) {
    id
    name
  }
}
    `;

export function useCreateQuizMutation() {
  return Urql.useMutation<CreateQuizMutation, CreateQuizMutationVariables>(CreateQuizDocument);
};
export const UpdateQuizDocument = gql`
    mutation UpdateQuiz($id: ID!, $input: UpdateQuizInput!) {
  updateQuiz(id: $id, input: $input) {
    id
    name
  }
}
    `;

export function useUpdateQuizMutation() {
  return Urql.useMutation<UpdateQuizMutation, UpdateQuizMutationVariables>(UpdateQuizDocument);
};
export const AddQuizQuestionDocument = gql`
    mutation AddQuizQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
  addQuizQuestion(quizId: $quizId, input: $input) {
    ...QuizQuestionFields
  }
}
    ${QuizQuestionFieldsFragmentDoc}`;

export function useAddQuizQuestionMutation() {
  return Urql.useMutation<AddQuizQuestionMutation, AddQuizQuestionMutationVariables>(AddQuizQuestionDocument);
};
export const UpdateQuizQuestionDocument = gql`
    mutation UpdateQuizQuestion($id: ID!, $input: UpdateQuizQuestionInput!) {
  updateQuizQuestion(id: $id, input: $input) {
    ...QuizQuestionFields
  }
}
    ${QuizQuestionFieldsFragmentDoc}`;

export function useUpdateQuizQuestionMutation() {
  return Urql.useMutation<UpdateQuizQuestionMutation, UpdateQuizQuestionMutationVariables>(UpdateQuizQuestionDocument);
};
export const DeleteQuizQuestionDocument = gql`
    mutation DeleteQuizQuestion($id: ID!) {
  deleteQuizQuestion(id: $id)
}
    `;

export function useDeleteQuizQuestionMutation() {
  return Urql.useMutation<DeleteQuizQuestionMutation, DeleteQuizQuestionMutationVariables>(DeleteQuizQuestionDocument);
};
export const StartQuizSessionDocument = gql`
    mutation StartQuizSession($sessionId: ID!) {
  startQuizSession(sessionId: $sessionId) {
    id
    startedAt
    expiresAt
    isExpired
    questionOrder
    orderedQuestions {
      ...QuizQuestionUserFields
    }
    quiz {
      id
      name
      timeoutSeconds
    }
  }
}
    ${QuizQuestionUserFieldsFragmentDoc}`;

export function useStartQuizSessionMutation() {
  return Urql.useMutation<StartQuizSessionMutation, StartQuizSessionMutationVariables>(StartQuizSessionDocument);
};
export const SubmitQuizAnswerDocument = gql`
    mutation SubmitQuizAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
  submitQuizAnswer(submissionId: $submissionId, input: $input) {
    __typename
    id
    answeredAt
    timeSpentSeconds
    question {
      id
    }
    ... on PredefinedResponse {
      isCorrect
      selectedAnswerIds
      selectedAnswers {
        id
        answerText
        isCorrect
      }
    }
    ... on OrderingResponse {
      isCorrect
      submittedOrder
    }
    ... on NumberResponse {
      numberResponse
    }
    ... on FreeTextResponse {
      textResponse
    }
  }
}
    `;

export function useSubmitQuizAnswerMutation() {
  return Urql.useMutation<SubmitQuizAnswerMutation, SubmitQuizAnswerMutationVariables>(SubmitQuizAnswerDocument);
};
export const UpdateQuizAnswerDocument = gql`
    mutation UpdateQuizAnswer($responseId: ID!, $input: UpdateQuizAnswerInput!) {
  updateQuizAnswer(responseId: $responseId, input: $input) {
    __typename
    id
    ... on OrderingResponse {
      isCorrect
      submittedOrder
    }
  }
}
    `;

export function useUpdateQuizAnswerMutation() {
  return Urql.useMutation<UpdateQuizAnswerMutation, UpdateQuizAnswerMutationVariables>(UpdateQuizAnswerDocument);
};
export const FinalizeQuizDocument = gql`
    mutation FinalizeQuiz($submissionId: ID!) {
  finalizeQuiz(submissionId: $submissionId) {
    ...QuizSubmissionResultFields
  }
}
    ${QuizSubmissionResultFieldsFragmentDoc}`;

export function useFinalizeQuizMutation() {
  return Urql.useMutation<FinalizeQuizMutation, FinalizeQuizMutationVariables>(FinalizeQuizDocument);
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
export const AdminExternalContentsDocument = gql`
    query AdminExternalContents($filter: ExternalContentFilter!, $sortBy: ExternalContentSortBy, $first: Int, $after: String) {
  externalContents(filter: $filter, sortBy: $sortBy, first: $first, after: $after) {
    edges {
      node {
        id
        planId
        taskId
        contentId
        contentType
        publishedAt
        source
        title
        translations {
          languageCode
          title
        }
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
    `;

export function useAdminExternalContentsQuery(options?: Omit<Urql.UseQueryArgs<never, AdminExternalContentsQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminExternalContentsQuery, AdminExternalContentsQueryVariables | undefined>({ query: AdminExternalContentsDocument, variables: undefined, ...options });
};
export const AdminProjectStreaksDocument = gql`
    query AdminProjectStreaks($projectId: ID!) {
  streaks(filter: {projectId: $projectId}, first: 100) {
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

export function useAdminProjectStreaksQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectStreaksQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectStreaksQuery, AdminProjectStreaksQueryVariables | undefined>({ query: AdminProjectStreaksDocument, variables: undefined, ...options });
};
export const AdminProjectQuizzesDocument = gql`
    query AdminProjectQuizzes($projectId: ID!) {
  quizzes(filter: {projectId: $projectId}, first: 100) {
    edges {
      node {
        id
        name
      }
    }
  }
}
    `;

export function useAdminProjectQuizzesQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectQuizzesQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectQuizzesQuery, AdminProjectQuizzesQueryVariables | undefined>({ query: AdminProjectQuizzesDocument, variables: undefined, ...options });
};
export const GetFirebaseTokenDocument = gql`
    query GetFirebaseToken {
  firebaseToken {
    token
    expiresIn
  }
}
    `;

export function useGetFirebaseTokenQuery(options?: Omit<Urql.UseQueryArgs<never, GetFirebaseTokenQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<GetFirebaseTokenQuery, GetFirebaseTokenQueryVariables | undefined>({ query: GetFirebaseTokenDocument, variables: undefined, ...options });
};
export const ChallengePageDocument = gql`
    query ChallengePage($challengeId: ID!) {
  challenge(id: $challengeId) {
    __typename
    id
    name
    description
    requiresTeamMembership
    requiresSuperTeamMembership
    userEnrolledAt
    userCompletedAt
    ... on SimpleChallenge {
      allowSelfCompletion
    }
    ... on PluginChallenge {
      pluginChallengeId
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
        endTime
        userCanStart
        userActiveSubmission {
          id
        }
        userActiveSession {
          id
          state
        }
        userSubmissions {
          id
          startedAt
          completedAt
          expiresAt
          isExpired
          score
          maxScore
          scorePercentage
          pointsAwarded
          orderedQuestions {
            ...QuizQuestionUserFields
          }
          responses {
            __typename
            id
            answeredAt
            timeSpentSeconds
            question {
              id
            }
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
            ... on OrderingResponse {
              isCorrect
              submittedOrder
            }
          }
        }
      }
    }
  }
}
    ${QuizQuestionUserFieldsFragmentDoc}`;

export function useChallengePageQuery(options?: Omit<Urql.UseQueryArgs<never, ChallengePageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ChallengePageQuery, ChallengePageQueryVariables | undefined>({ query: ChallengePageDocument, variables: undefined, ...options });
};
export const ChallengesPageDocument = gql`
    query ChallengesPage {
  myCurrentProject {
    myTeam {
      joinCode
    }
    challenges {
      __typename
      id
      name
      description
      imageObject {
        ...ImageFields
      }
      buttonText
      publishedAt
      endTime
      visibleAt
      userCompletedAt
      ... on SimpleChallenge {
        allowSelfCompletion
      }
      ... on ExternalChallenge {
        url
      }
      ... on QuizChallenge {
        quiz {
          userCanStart
          userActiveSession {
            id
          }
        }
      }
    }
  }
}
    ${ImageFieldsFragmentDoc}`;

export function useChallengesPageQuery(options?: Omit<Urql.UseQueryArgs<never, ChallengesPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ChallengesPageQuery, ChallengesPageQueryVariables | undefined>({ query: ChallengesPageDocument, variables: undefined, ...options });
};
export const ProfilePageDocument = gql`
    query ProfilePage($ageFilter: LeaderboardFilter) {
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
        shortText
        url
        managementType
        managedBy
      }
    }
  }
  myCurrentProject {
    id
    name
    infoMessage {
      markdown
      html
    }
    infoMessageStart
    infoMessageEnd
    branding {
      ...BrandingFields
    }
    achievements {
      __typename
      id
      name
      descriptionPending
      descriptionCompleted
      imagePendingObject {
        ...ImageFields
      }
      imageCompletedObject {
        ...ImageFields
      }
      hidden
      achievedAt
      celebratedAt
      points
      ... on ContentAchievement {
        nextItem {
          id
          sortOrder
          externalContent {
            id
            title
            url
          }
        }
      }
    }
    leaderboard(entityType: PERSONS, filter: $ageFilter) {
      me {
        score
        rank
      }
    }
  }
}
    ${BrandingFieldsFragmentDoc}
${ImageFieldsFragmentDoc}`;

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
        shortText
        publishedAt
        managedBy
        managementType
        url
      }
      acceptedConsents {
        __typename
        id
        consent {
          id
          title
          body {
            html
          }
          shortText
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
          id
          title
          body {
            html
          }
          shortText
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
          ...LeaderboardEntryWithDescriptionFields
        }
      }
      me {
        ...LeaderboardEntryWithDescriptionFields
      }
    }
  }
}
    ${LeaderboardEntryWithDescriptionFieldsFragmentDoc}`;

export function useStandingsGlobalPageQuery(options?: Omit<Urql.UseQueryArgs<never, StandingsGlobalPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<StandingsGlobalPageQuery, StandingsGlobalPageQueryVariables | undefined>({ query: StandingsGlobalPageDocument, variables: undefined, ...options });
};
export const StandingsLocalPageDocument = gql`
    query StandingsLocalPage($filter: LeaderboardFilter, $first: Int) {
  me {
    church {
      id
      name
    }
  }
  myCurrentProject {
    id
    personLeaderboard: leaderboard(
      entityType: PERSONS
      filter: $filter
      first: $first
    ) {
      totalCount
      edges {
        node {
          ...LeaderboardEntryFields
        }
      }
      me {
        ...LeaderboardEntryFields
      }
    }
    unitLeaderboard: leaderboard(entityType: TEAMS, filter: $filter, first: $first) {
      totalCount
      edges {
        node {
          ...LeaderboardEntryFields
        }
      }
      me {
        ...LeaderboardEntryFields
      }
    }
  }
}
    ${LeaderboardEntryFieldsFragmentDoc}`;

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
        ...LeaderboardEntryFields
      }
    }
  }
}
    ${LeaderboardEntryFieldsFragmentDoc}`;

export function useStandingsUnitPageQuery(options?: Omit<Urql.UseQueryArgs<never, StandingsUnitPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<StandingsUnitPageQuery, StandingsUnitPageQueryVariables | undefined>({ query: StandingsUnitPageDocument, variables: undefined, ...options });
};
export const VapidPublicKeyDocument = gql`
    query VapidPublicKey {
  vapidPublicKey
}
    `;

export function useVapidPublicKeyQuery(options?: Omit<Urql.UseQueryArgs<never, VapidPublicKeyQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<VapidPublicKeyQuery, VapidPublicKeyQueryVariables | undefined>({ query: VapidPublicKeyDocument, variables: undefined, ...options });
};
export const CurrentProjectDocument = gql`
    query CurrentProject {
  myCurrentProject {
    branding {
      ...BrandingFields
    }
  }
}
    ${BrandingFieldsFragmentDoc}`;

export function useCurrentProjectQuery(options?: Omit<Urql.UseQueryArgs<never, CurrentProjectQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<CurrentProjectQuery, CurrentProjectQueryVariables | undefined>({ query: CurrentProjectDocument, variables: undefined, ...options });
};
export const AdminChurchPageDocument = gql`
    query AdminChurchPage($id: ID!) {
  church(id: $id) {
    id
    name
    country
    category
  }
}
    `;

export function useAdminChurchPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminChurchPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminChurchPageQuery, AdminChurchPageQueryVariables | undefined>({ query: AdminChurchPageDocument, variables: undefined, ...options });
};
export const UpdateChurchDocument = gql`
    mutation UpdateChurch($id: ID!, $input: UpdateChurchInput!) {
  updateChurch(id: $id, input: $input) {
    id
    name
    country
    category
  }
}
    `;

export function useUpdateChurchMutation() {
  return Urql.useMutation<UpdateChurchMutation, UpdateChurchMutationVariables>(UpdateChurchDocument);
};
export const AdminConsentPageDocument = gql`
    query AdminConsentPage($id: ID!) {
  consent(id: $id) {
    id
    key
    version
    title
    shortText
    body {
      markdown
      html
    }
    url
    publishedAt
    managementType
    managedBy
  }
}
    `;

export function useAdminConsentPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminConsentPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminConsentPageQuery, AdminConsentPageQueryVariables | undefined>({ query: AdminConsentPageDocument, variables: undefined, ...options });
};
export const AdminConsentsPageDocument = gql`
    query AdminConsentsPage {
  consents {
    id
    key
    version
    title
    shortText
    publishedAt
    managementType
    managedBy
  }
}
    `;

export function useAdminConsentsPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminConsentsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminConsentsPageQuery, AdminConsentsPageQueryVariables | undefined>({ query: AdminConsentsPageDocument, variables: undefined, ...options });
};
export const AdminFeedbackPageDocument = gql`
    query AdminFeedbackPage($filter: FeedbackFilter, $first: Int, $after: String, $last: Int, $before: String) {
  feedback(
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
        message
        canContactMe
        userAgent
        platform
        screenWidth
        screenHeight
        appVersion
        locale
        projectId
        timezone
        contextUrl
        tags
        createdAt
        handledAt
        user {
          id
          name
          email
        }
      }
    }
  }
}
    `;

export function useAdminFeedbackPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminFeedbackPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminFeedbackPageQuery, AdminFeedbackPageQueryVariables | undefined>({ query: AdminFeedbackPageDocument, variables: undefined, ...options });
};
export const UpdateFeedbackTagsDocument = gql`
    mutation UpdateFeedbackTags($feedbackId: ID!, $tags: [String!]!) {
  updateFeedbackTags(feedbackId: $feedbackId, tags: $tags) {
    id
    tags
  }
}
    `;

export function useUpdateFeedbackTagsMutation() {
  return Urql.useMutation<UpdateFeedbackTagsMutation, UpdateFeedbackTagsMutationVariables>(UpdateFeedbackTagsDocument);
};
export const AdminHomePageDocument = gql`
    query AdminHomePage($now: DateTime!) {
  me {
    id
    name
  }
  adminDashboardStats {
    totalUsers
    totalPointsAwarded
    newUsersLast7Days
  }
  feedback(first: 5) {
    edges {
      node {
        id
        message
        createdAt
        user {
          id
          name
        }
      }
    }
  }
  projects(filter: {endDateAfter: $now}) {
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
export const ChurchAdminsPageDocument = gql`
    query ChurchAdminsPage($churchId: ID!) {
  usersWithRole(role: CHURCH_ADMIN, scopeType: CHURCH, scopeId: $churchId) {
    id
    name
    email
  }
  users(filter: {churchId: $churchId}, first: 500) {
    edges {
      node {
        id
        name
        email
      }
    }
  }
}
    `;

export function useChurchAdminsPageQuery(options?: Omit<Urql.UseQueryArgs<never, ChurchAdminsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<ChurchAdminsPageQuery, ChurchAdminsPageQueryVariables | undefined>({ query: ChurchAdminsPageDocument, variables: undefined, ...options });
};
export const MyChurchUnitsPageDocument = gql`
    query MyChurchUnitsPage($filter: UserFilter) {
  users(filter: $filter, first: 1000) {
    edges {
      node {
        id
        name
        age
        gender
        teams {
          id
          name
        }
      }
    }
  }
  myCurrentProject {
    id
    name
    myChurchTeams {
      id
      name
      leaderboardExcluded
      averageAge
      joinCode
      members {
        id
        name
        isTeamLead
        user {
          id
          age
          gender
        }
      }
    }
  }
}
    `;

export function useMyChurchUnitsPageQuery(options?: Omit<Urql.UseQueryArgs<never, MyChurchUnitsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<MyChurchUnitsPageQuery, MyChurchUnitsPageQueryVariables | undefined>({ query: MyChurchUnitsPageDocument, variables: undefined, ...options });
};
export const AdminProjectAchievementPageDocument = gql`
    query AdminProjectAchievementPage($achievementId: ID!) {
  achievement(id: $achievementId) {
    __typename
    id
    name
    descriptionPending
    descriptionCompleted
    imagePendingObject {
      ...ImageFields
    }
    imageCompletedObject {
      ...ImageFields
    }
    notificationText
    achievedAt
    points
    hidden
    awardableFrom
    ... on ContentAchievement {
      items {
        id
        sortOrder
        externalContent {
          id
          planId
          taskId
          contentId
          contentType
          publishedAt
          source
          syncedAt
          createdAt
          updatedAt
          title
          translations {
            languageCode
            title
          }
        }
      }
    }
    ... on StreakAchievement {
      neededStreak
      streak {
        id
        name
        description
      }
    }
    ... on QuizAchievement {
      quiz {
        id
        name
      }
      minScorePercentage
      requireCompletion
    }
    project {
      id
      name
      branding {
        colors {
          ...BrandingColorsFields
        }
      }
    }
  }
}
    ${ImageFieldsFragmentDoc}
${BrandingColorsFieldsFragmentDoc}`;

export function useAdminProjectAchievementPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectAchievementPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectAchievementPageQuery, AdminProjectAchievementPageQueryVariables | undefined>({ query: AdminProjectAchievementPageDocument, variables: undefined, ...options });
};
export const AdminProjectAchievementsNewPageDocument = gql`
    query AdminProjectAchievementsNewPage($projectId: ID!) {
  project(id: $projectId) {
    id
    name
    branding {
      colors {
        ...BrandingColorsFields
      }
    }
  }
}
    ${BrandingColorsFieldsFragmentDoc}`;

export function useAdminProjectAchievementsNewPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectAchievementsNewPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectAchievementsNewPageQuery, AdminProjectAchievementsNewPageQueryVariables | undefined>({ query: AdminProjectAchievementsNewPageDocument, variables: undefined, ...options });
};
export const AdminProjectChallengePageDocument = gql`
    query AdminProjectChallengePage($challengeId: ID!) {
  challenge(id: $challengeId) {
    __typename
    id
    name
    description
    image
    buttonText
    visibleAt
    startedAt
    endTime
    project {
      id
      name
      branding {
        colors {
          ...BrandingColorsFields
        }
      }
    }
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
        image
        timeoutSeconds
        randomizeQuestions
        revealCorrectAnswers
        allowRetakes
        completionPoints
        questions {
          ...QuizQuestionFields
        }
      }
    }
    ... on PluginChallenge {
      pluginChallengeId
    }
  }
}
    ${BrandingColorsFieldsFragmentDoc}
${QuizQuestionFieldsFragmentDoc}`;

export function useAdminProjectChallengePageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminProjectChallengePageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminProjectChallengePageQuery, AdminProjectChallengePageQueryVariables | undefined>({ query: AdminProjectChallengePageDocument, variables: undefined, ...options });
};
export const AdminProjectChallengeNewPageDocument = gql`
    query AdminProjectChallengeNewPage($projectId: ID!) {
  project(id: $projectId) {
    id
    name
    branding {
      colors {
        ...BrandingColorsFields
      }
    }
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
    ${BrandingColorsFieldsFragmentDoc}`;

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
      ...BrandingFields
    }
    rules {
      markdown
      html
    }
    infoMessage {
      markdown
      html
    }
    infoMessageStart
    infoMessageEnd
  }
}
    ${BrandingFieldsFragmentDoc}`;

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
      logoImage {
        ...ImageFields
      }
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
        descriptionPending
        descriptionCompleted
        imagePendingObject {
          ...ImageFields
        }
        imageCompletedObject {
          ...ImageFields
        }
        points
        hidden
      }
    }
  }
  challenges(first: 50, filter: {projectId: $projectId}) {
    edges {
      node {
        __typename
        id
        name
        description
        imageObject {
          ...ImageFields
        }
      }
    }
  }
}
    ${ImageFieldsFragmentDoc}`;

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
export const AdminScoresPageDocument = gql`
    query AdminScoresPage($filter: ScoreJournalFilter, $first: Int, $after: String, $last: Int, $before: String) {
  adminScoreJournal(
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
        points
        sourceType
        reason
        createdAt
        user {
          id
          name
        }
        project {
          id
          name
        }
        awardedBy {
          id
          name
        }
      }
    }
  }
}
    `;

export function useAdminScoresPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminScoresPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminScoresPageQuery, AdminScoresPageQueryVariables | undefined>({ query: AdminScoresPageDocument, variables: undefined, ...options });
};
export const DeleteScoreJournalEntryDocument = gql`
    mutation DeleteScoreJournalEntry($id: ID!) {
  deleteScoreJournalEntry(id: $id)
}
    `;

export function useDeleteScoreJournalEntryMutation() {
  return Urql.useMutation<DeleteScoreJournalEntryMutation, DeleteScoreJournalEntryMutationVariables>(DeleteScoreJournalEntryDocument);
};
export const AdminScoresNewPageDocument = gql`
    query AdminScoresNewPage {
  projects(first: 100) {
    edges {
      node {
        id
        name
      }
    }
  }
}
    `;

export function useAdminScoresNewPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminScoresNewPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminScoresNewPageQuery, AdminScoresNewPageQueryVariables | undefined>({ query: AdminScoresNewPageDocument, variables: undefined, ...options });
};
export const CreateScoreAdjustmentDocument = gql`
    mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
  createScoreAdjustment(input: $input) {
    id
    points
    reason
  }
}
    `;

export function useCreateScoreAdjustmentMutation() {
  return Urql.useMutation<CreateScoreAdjustmentMutation, CreateScoreAdjustmentMutationVariables>(CreateScoreAdjustmentDocument);
};
export const AdminTeamPageDocument = gql`
    query AdminTeamPage($id: ID!) {
  team(id: $id) {
    id
    name
    description
    joinCode
    leaderboardExcluded
    averageAge
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
    personUuid
    createdAt
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
    consentStatus {
      acceptedConsents {
        id
        action
        actionDate
        consent {
          id
          key
          title
          version
          managementType
        }
      }
      rejectedConsents {
        id
        action
        actionDate
        consent {
          id
          key
          title
          version
        }
      }
      pendingConsents {
        id
        key
        title
        version
      }
    }
  }
  adminScoreJournal(filter: {userId: $id}, first: 20) {
    totalCount
    edges {
      node {
        id
        points
        sourceType
        reason
        createdAt
        project {
          id
          name
        }
        awardedBy {
          id
          name
        }
      }
    }
  }
  feedback(filter: {userId: $id}, first: 10) {
    totalCount
    edges {
      node {
        id
        message
        canContactMe
        userAgent
        platform
        screenWidth
        screenHeight
        appVersion
        createdAt
      }
    }
  }
}
    `;

export function useAdminUserPageQuery(options?: Omit<Urql.UseQueryArgs<never, AdminUserPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<AdminUserPageQuery, AdminUserPageQueryVariables | undefined>({ query: AdminUserPageDocument, variables: undefined, ...options });
};
export const AdminSetUserConsentDocument = gql`
    mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
  adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
    id
    action
  }
}
    `;

export function useAdminSetUserConsentMutation() {
  return Urql.useMutation<AdminSetUserConsentMutation, AdminSetUserConsentMutationVariables>(AdminSetUserConsentDocument);
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
export const StandingsPageDocument = gql`
    query StandingsPage {
  myCurrentProject {
    myTeam {
      id
    }
  }
}
    `;

export function useStandingsPageQuery(options?: Omit<Urql.UseQueryArgs<never, StandingsPageQueryVariables | undefined>, 'query'>) {
  return Urql.useQuery<StandingsPageQuery, StandingsPageQueryVariables | undefined>({ query: StandingsPageDocument, variables: undefined, ...options });
};