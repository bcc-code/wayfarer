import { useAuth0 } from '@auth0/auth0-vue'

gql(`
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
`)

export function useAuth() {
  const { reset } = useAnalytics()
  const auth0 = useAuth0()
  const config = useRuntimeConfig()

  // Wayfarer JWT stored in localStorage (exchanged from Auth0 token)
  const wayfarerToken = useLocalStorage<string>('token', () => null)
  const isLoading = useState('isLoading', () => true)
  const me = useState<GetMeQuery['me'] | null | undefined>('me', () => null)

  // Only fetch user data if we have a wayfarer token AND we're not on the callback page
  const { isAuthReady } = useAuthReady()

  const { data, fetching } = useGetMeQuery({
    pause: computed(() => !isAuthReady.value),
  })

  // If we already have user data, mark as not loading
  if (me.value) {
    isLoading.value = false
  }

  // If we don't have a token, we're not loading (we're just not authenticated)
  if (!wayfarerToken.value) {
    isLoading.value = false
  }

  // Update loading state: we're done loading when query is not fetching
  // This handles both completed queries and paused queries
  watch(
    fetching,
    (isFetching) => {
      if (!isFetching) {
        isLoading.value = false
      }
    },
    { immediate: true },
  )

  watch(
    () => data.value?.me,
    (newMe) => {
      me.value = newMe
      isLoading.value = false
    },
    { immediate: true },
  )

  // Get the Wayfarer token for API calls
  const getAccessToken = async () => {
    try {
      while (isLoading.value) {
        await new Promise((resolve) => setTimeout(resolve, 10))
      }
      return wayfarerToken.value
    } catch {
      await loginWithRedirect()
    }
  }

  const getAccessTokenSilently = async () => {
    return wayfarerToken.value
  }

  // Get Auth0 access token (for exchanging with backend)
  const getAuth0Token = async () => {
    try {
      return await auth0.getAccessTokenSilently()
    } catch {
      return null
    }
  }

  const setAccessToken = (value: string) => {
    wayfarerToken.value = value
    isLoading.value = false
  }

  const { track } = useAnalytics()

  async function loginWithRedirect() {
    return auth0.loginWithRedirect({
      appState: {
        targetUrl: window.location.pathname,
      },
    })
  }

  async function logout() {
    track(AnalyticsEvent.LogoutCompleted)
    reset()
    wayfarerToken.value = null
    me.value = null
    return auth0.logout({
      logoutParams: {
        returnTo: `${window.location.origin}/logout-callback`,
      },
    })
  }

  // Exchange Auth0 token for Wayfarer JWT
  async function exchangeToken(): Promise<boolean> {
    try {
      const auth0Token = await getAuth0Token()
      if (!auth0Token) {
        return false
      }

      const response = await $fetch<{ token: string }>(
        `${config.public.tokenUrl}?token=${auth0Token}`,
        { method: 'GET' },
      )

      if (response && response.token) {
        setAccessToken(response.token)
        return true
      }
      return false
    } catch {
      return false
    }
  }

  // Authorization
  const isSuperAdmin = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.Superadmin)
  })
  const isAdmin = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.Admin)
  })
  const isChurchAdmin = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.ChurchAdmin)
  })
  const isProjectAdmin = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.ProjectAdmin) // TODO: scope project admin to current project
  })
  const isTeamLead = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.TeamLead) // TODO: scope team lead to current team
  })
  const isM2M = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.M2M)
  })

  return {
    // Auth0 state
    isAuthenticated: auth0.isAuthenticated,
    isAuth0Loading: auth0.isLoading,

    // Token methods
    getAccessTokenSilently,
    getAccessToken,
    getAuth0Token,
    setAccessToken,
    exchangeToken,

    // Auth actions
    loginWithRedirect,
    logout,

    // Loading and user state
    isLoading,
    me,
    token: wayfarerToken,

    // Role checks
    isProjectAdmin,
    isTeamLead,
    isM2M,
    isSuperAdmin,
    isAdmin,
    isChurchAdmin,
  }
}
