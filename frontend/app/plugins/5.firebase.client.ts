import { initializeApp, getApps, type FirebaseApp } from 'firebase/app'
import { getAuth, type Auth } from 'firebase/auth'
import { getFirestore, type Firestore } from 'firebase/firestore'

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig()

  // Skip initialization if Firebase config is not provided
  if (
    !config.public.firebaseDatabase ||
    !config.public.firebaseApiKey ||
    !config.public.firebaseProjectId ||
    !config.public.firebaseAuthDomain
  ) {
    return {
      provide: {
        firebase: null as FirebaseApp | null,
        firebaseAuth: null as Auth | null,
        firestore: null as Firestore | null,
      },
    }
  }

  // Prevent re-initialization
  let app: FirebaseApp
  if (getApps().length === 0) {
    app = initializeApp({
      apiKey: config.public.firebaseApiKey,
      authDomain: config.public.firebaseAuthDomain,
      projectId: config.public.firebaseProjectId,
    })
  } else {
    app = getApps()[0]!
  }

  const auth: Auth = getAuth(app)
  const firestore: Firestore = config.public.firebaseDatabase
    ? getFirestore(app, config.public.firebaseDatabase)
    : getFirestore(app)

  return {
    provide: {
      firebase: app as FirebaseApp | null,
      firebaseAuth: auth as Auth | null,
      firestore: firestore as Firestore | null,
    },
  }
})
