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

  return {
    getAccessTokenSilently,
    getAccessToken,
    setAccessToken,
    loginWithRedirect,
    isLoading,
    me,
    token,
  }
}
