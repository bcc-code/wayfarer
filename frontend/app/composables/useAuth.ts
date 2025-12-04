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
  const token = useCookie('token')
  const isLoading = useState('isLoading', () => true)
  const me = useState<GetMeQuery['me'] | null | undefined>('me', () => null)

  // Only fetch user data if we have a token AND we're not on the callback page
  const { isAuthReady } = useAuthReady()

  const { data, fetching } = useGetMeQuery({
    pause: computed(() => !isAuthReady.value),
  })

  // If we already have user data, mark as not loading
  if (me.value) {
    isLoading.value = false
  }

  // If we don't have a token, we're not loading (we're just not authenticated)
  if (!token.value) {
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

  const getAccessToken = async () => {
    try {
      while (isLoading.value) {
        await new Promise((resolve) => setTimeout(resolve, 10))
      }
      return token.value
    } catch {
      await loginWithRedirect()
    }
  }

  const getAccessTokenSilently = async () => {
    return token.value
  }

  const setAccessToken = (value: string) => {
    token.value = value
    isLoading.value = false
  }

  const config = useRuntimeConfig()
  const { track } = useAnalytics()

  function loginWithRedirect() {
    return navigateTo(
      `${config.public.loginUrl}?redirect=${window.location.pathname}`,
      {
        external: true,
      },
    )
  }

  function logout() {
    track(AnalyticsEvent.LogoutCompleted)
    reset()
    token.value = null
    me.value = null
    return navigateTo('/')
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
    getAccessTokenSilently,
    getAccessToken,
    setAccessToken,
    loginWithRedirect,
    logout,
    isLoading,
    me,
    token,
    isProjectAdmin,
    isTeamLead,
    isM2M,
    isSuperAdmin,
    isAdmin,
    isChurchAdmin,
  }
}
