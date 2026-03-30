import { signInWithCustomToken, signOut, type Auth } from 'firebase/auth'
import {
  doc,
  onSnapshot,
  type Unsubscribe,
  type Firestore,
  type DocumentSnapshot,
} from 'firebase/firestore'
import { useClientHandle } from '@urql/vue'

const NOTIFICATION_QUERY_MAP = {
  achievements: ['ProfilePageDocument'],
  challenges: [
    'ActiveChallengesPageDocument',
    'ChallengePageDocument',
    'CurrentProjectDocument',
  ],
  content: ['ProfilePageDocument'],
  quizzes: ['ActiveChallengesPageDocument', 'ChallengePageDocument'],
  projects: ['ProfilePageDocument', 'CurrentProjectDocument'],
} as const

const ADMIN_NOTIFICATION_QUERY_MAP = {
  feedback: ['AdminFeedbackPageDocument'],
} as const

const PROJECT_NOTIFICATION_QUERY_MAP = {
  quiz_sessions: ['ActiveChallengesPageDocument', 'ChallengePageDocument'],
  challenges: [
    'ActiveChallengesPageDocument',
    'ChallengePageDocument',
    'CurrentProjectDocument',
  ],
} as const

type NotificationCategory = keyof typeof NOTIFICATION_QUERY_MAP
type AdminNotificationCategory = keyof typeof ADMIN_NOTIFICATION_QUERY_MAP
type ProjectNotificationCategory = keyof typeof PROJECT_NOTIFICATION_QUERY_MAP

// Shared state - singleton pattern for app-wide listeners
const isInitialized = ref(false)
const isAuthenticated = ref(false)
const error = ref<Error | null>(null)
const listeners = new Map<string, Unsubscribe>()
let tokenRefreshTimer: ReturnType<typeof setTimeout> | null = null

export function useFirestoreSync() {
  const { $firebaseAuth, $firestore } = useNuxtApp()
  const { me, token: wayfarerToken } = useAuth()
  const { client } = useClientHandle()

  async function fetchFirebaseToken(): Promise<{
    token: string
    expiresIn: number
  } | null> {
    try {
      const result = await client
        .query(GetFirebaseTokenDocument, {}, { requestPolicy: 'network-only' })
        .toPromise()
      if (result.error || !result.data?.firebaseToken) {
        throw new Error(result.error?.message || 'Failed to get Firebase token')
      }
      return result.data.firebaseToken
    } catch (err) {
      error.value =
        err instanceof Error ? err : new Error('Failed to fetch Firebase token')
      return null
    }
  }

  async function authenticate(): Promise<boolean> {
    const auth = $firebaseAuth as Auth | null
    if (!auth) return false

    const tokenResponse = await fetchFirebaseToken()
    if (!tokenResponse) return false

    try {
      await signInWithCustomToken(auth, tokenResponse.token)
      isAuthenticated.value = true
      scheduleTokenRefresh(tokenResponse.expiresIn)
      return true
    } catch (err) {
      error.value =
        err instanceof Error ? err : new Error('Firebase authentication failed')
      return false
    }
  }

  function scheduleTokenRefresh(expiresInSeconds: number) {
    if (tokenRefreshTimer) {
      clearTimeout(tokenRefreshTimer)
    }

    // Refresh 5 minutes before expiry, or at 90% of lifetime
    const refreshDelay = Math.min(
      (expiresInSeconds - 300) * 1000,
      expiresInSeconds * 0.9 * 1000,
    )

    if (refreshDelay > 0) {
      tokenRefreshTimer = setTimeout(async () => {
        await authenticate()
      }, refreshDelay)
    }
  }

  function subscribeToCategory(
    userId: string,
    category: NotificationCategory,
  ): Unsubscribe {
    const firestore = $firestore as Firestore | null
    if (!firestore) return () => {}
    const path = `users/${userId}/notifications/${category}`
    const docRef = doc(firestore, path)
    let isInitialSnapshot = true

    return onSnapshot(
      docRef,
      (snapshot: DocumentSnapshot) => {
        // Skip the initial snapshot that fires on subscribe
        if (isInitialSnapshot) {
          isInitialSnapshot = false
          return
        }

        if (!snapshot.exists()) return

        // Dispatch custom event for pages to handle
        const queries = NOTIFICATION_QUERY_MAP[category]
        for (const queryName of queries) {
          window.dispatchEvent(
            new CustomEvent('firestore-update', {
              detail: { query: queryName },
            }),
          )
        }
      },
      (err: Error) => {
        console.error(`Firestore listener error for ${category}:`, err)
      },
    )
  }

  function subscribeToAdminCategory(
    category: AdminNotificationCategory,
  ): Unsubscribe {
    const firestore = $firestore as Firestore | null
    if (!firestore) return () => {}
    const path = `admin/${category}`
    const docRef = doc(firestore, path)
    let isInitialSnapshot = true

    return onSnapshot(
      docRef,
      (snapshot: DocumentSnapshot) => {
        if (isInitialSnapshot) {
          isInitialSnapshot = false
          return
        }

        if (!snapshot.exists()) return

        const queries = ADMIN_NOTIFICATION_QUERY_MAP[category]
        for (const queryName of queries) {
          window.dispatchEvent(
            new CustomEvent('firestore-update', {
              detail: { query: queryName },
            }),
          )
        }
      },
      (err: Error) => {
        console.error(`Firestore admin listener error for ${category}:`, err)
      },
    )
  }

  function subscribeToProjectCategory(
    projectId: string,
    category: ProjectNotificationCategory,
  ): Unsubscribe {
    const firestore = $firestore as Firestore | null
    if (!firestore) return () => {}
    const path = `projects/${projectId}/notifications/${category}`
    const docRef = doc(firestore, path)
    let isInitialSnapshot = true

    return onSnapshot(
      docRef,
      (snapshot: DocumentSnapshot) => {
        if (isInitialSnapshot) {
          isInitialSnapshot = false
          return
        }

        if (!snapshot.exists()) return

        const queries = PROJECT_NOTIFICATION_QUERY_MAP[category]
        for (const queryName of queries) {
          window.dispatchEvent(
            new CustomEvent('firestore-update', {
              detail: { query: queryName },
            }),
          )
        }
      },
      (err: Error) => {
        console.error(
          `Firestore project listener error for ${category}:`,
          err,
        )
      },
    )
  }

  function subscribeAdmin(category: AdminNotificationCategory): () => void {
    if (!isAuthenticated.value) {
      console.warn('Cannot subscribe to admin notifications: not authenticated')
      return () => {}
    }

    const key = `admin/${category}`
    if (listeners.has(key)) {
      return () => {}
    }

    const unsubscribe = subscribeToAdminCategory(category)
    listeners.set(key, unsubscribe)

    return () => {
      const unsub = listeners.get(key)
      if (unsub) {
        unsub()
        listeners.delete(key)
      }
    }
  }

  function subscribeProject(
    projectId: string,
    category: ProjectNotificationCategory,
  ): () => void {
    if (!isAuthenticated.value) {
      console.warn(
        'Cannot subscribe to project notifications: not authenticated',
      )
      return () => {}
    }

    const key = `project/${projectId}/${category}`
    if (listeners.has(key)) {
      return () => {}
    }

    const unsubscribe = subscribeToProjectCategory(projectId, category)
    listeners.set(key, unsubscribe)

    return () => {
      const unsub = listeners.get(key)
      if (unsub) {
        unsub()
        listeners.delete(key)
      }
    }
  }

  async function initialize() {
    if (isInitialized.value) return
    if (!$firestore || !$firebaseAuth) return
    if (!me.value?.id) return

    const success = await authenticate()
    if (!success) return

    const userId = me.value.id
    const categories: NotificationCategory[] = [
      'achievements',
      'challenges',
      'content',
      'quizzes',
      'projects',
    ]

    for (const category of categories) {
      const key = `${userId}/${category}`
      if (!listeners.has(key)) {
        const unsubscribe = subscribeToCategory(userId, category)
        listeners.set(key, unsubscribe)
      }
    }

    isInitialized.value = true
  }

  function cleanup() {
    for (const unsubscribe of listeners.values()) {
      unsubscribe()
    }
    listeners.clear()

    if (tokenRefreshTimer) {
      clearTimeout(tokenRefreshTimer)
      tokenRefreshTimer = null
    }

    const auth = $firebaseAuth as Auth | null
    if (auth) {
      signOut(auth).catch(console.error)
    }

    isInitialized.value = false
    isAuthenticated.value = false
  }

  // Watch for user login/logout
  watch(
    () => me.value?.id,
    (userId, oldUserId) => {
      if (userId && !oldUserId) {
        initialize()
      } else if (!userId && oldUserId) {
        cleanup()
      }
    },
  )

  // Watch for Wayfarer token changes
  watch(wayfarerToken, (newToken) => {
    if (!newToken) {
      cleanup()
    }
  })

  return {
    isInitialized: readonly(isInitialized),
    isAuthenticated: readonly(isAuthenticated),
    error: readonly(error),
    initialize,
    cleanup,
    subscribeAdmin,
    subscribeProject,
  }
}
