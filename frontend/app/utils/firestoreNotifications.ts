// Which query documents refetch when a Firestore notification document
// changes. useFirestoreSync turns a ping into a refresh event for each listed
// document, and pages listen for those with useFirestoreRefresh. A document no
// map lists never refetches live (see firestoreNotifications.test.ts).
//
// ActiveChallengesPage and CompletedChallengesPage split the same challenges
// by completion, so every ping that changes one changes the other.

export const NOTIFICATION_QUERY_MAP = {
  achievements: ['ProfilePageDocument'],
  challenges: [
    'ActiveChallengesPageDocument',
    'CompletedChallengesPageDocument',
    'ChallengePageDocument',
    'CurrentProjectDocument',
  ],
  content: ['ProfilePageDocument'],
  quizzes: [
    'ActiveChallengesPageDocument',
    'CompletedChallengesPageDocument',
    'ChallengePageDocument',
  ],
  projects: ['ProfilePageDocument', 'CurrentProjectDocument'],
} as const

export const ADMIN_NOTIFICATION_QUERY_MAP = {
  feedback: ['AdminFeedbackPageDocument'],
} as const

export const PROJECT_NOTIFICATION_QUERY_MAP = {
  quiz_sessions: [
    'ActiveChallengesPageDocument',
    'CompletedChallengesPageDocument',
    'ChallengePageDocument',
  ],
  challenges: [
    'ActiveChallengesPageDocument',
    'CompletedChallengesPageDocument',
    'ChallengePageDocument',
    'CurrentProjectDocument',
  ],
  // Fired on the outgoing project when an admin switches which project is
  // current, so clients still subscribed to it refetch and re-theme.
  current_project: ['CurrentProjectDocument'],
} as const

// Every document some notification refetches
export const DISPATCHED_QUERY_DOCUMENTS: ReadonlySet<string> = new Set(
  [
    NOTIFICATION_QUERY_MAP,
    ADMIN_NOTIFICATION_QUERY_MAP,
    PROJECT_NOTIFICATION_QUERY_MAP,
  ].flatMap((map) => Object.values(map).flat()),
)
