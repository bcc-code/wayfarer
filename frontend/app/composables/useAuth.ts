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
  const token = useCookie('token')
  const isLoading = useState('isLoading', () => true)
  const me = useState<GetMeQuery['me'] | null | undefined>('me', () => null)
  const { data } = useGetMeQuery()

  watch(
    () => data.value?.me,
    (newMe) => {
      me.value = newMe
      isLoading.value = false
    },
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

  function loginWithRedirect() {
    return navigateTo(
      `${config.public.loginUrl}?redirect=${window.location.pathname}`,
      {
        external: true,
      },
    )
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
    return me.value?.roles.some((role) => role.role === RoleType.ProjectAdmin)
  })
  const isTeamLead = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.TeamLead)
  })
  const isM2M = computed(() => {
    return me.value?.roles.some((role) => role.role === RoleType.M2M)
  })

  return {
    getAccessTokenSilently,
    getAccessToken,
    setAccessToken,
    loginWithRedirect,
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
